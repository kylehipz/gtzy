package meter

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"

	"gtzy/internal/store"
)

// Standard Bluetooth Glucose Profile UUIDs.
var (
	glucoseServiceUUID = bluetooth.New16BitUUID(0x1808) // Glucose
	glucoseMeasUUID    = bluetooth.New16BitUUID(0x2A18) // Glucose Measurement
	racpUUID           = bluetooth.New16BitUUID(0x2A52) // Record Access Control Point
)

// adapter is the process-wide BLE adapter (BlueZ on Linux). It is a package var
// so it can be enabled once and reused across syncs.
var adapter = bluetooth.DefaultAdapter

// Sync connects to the paired glucose meter, requests stored records newer than
// lastSeq (or all records when lastSeq <= 0), decodes them, and returns them as
// store inputs. Dedup against already-synced records is handled by the caller's
// DB unique index. The meter must already be bonded to this machine (pair once
// out of band with `bluetoothctl`); set GTZY_METER_ADDR to its MAC, otherwise
// Sync scans for a device whose advertised name contains "accu-chek".
func Sync(ctx context.Context, lastSeq int64) ([]store.BloodSugarInput, error) {
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable bluetooth adapter: %w", err)
	}

	device, err := connect(ctx)
	if err != nil {
		return nil, err
	}
	defer device.Disconnect() //nolint:errcheck // best-effort cleanup

	measChar, racpChar, err := discover(device)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var measurements []measurement
	racpDone := make(chan error, 1)

	if err := measChar.EnableNotifications(func(buf []byte) {
		m, perr := parseMeasurement(buf)
		if perr != nil {
			return // skip records we can't decode rather than fail the whole sync
		}
		mu.Lock()
		measurements = append(measurements, m)
		mu.Unlock()
	}); err != nil {
		return nil, fmt.Errorf("subscribe to glucose measurements: %w", err)
	}

	if err := racpChar.EnableNotifications(func(buf []byte) {
		// RACP response is opcode 0x06 (Response Code Response):
		// [0x06, operator, request-opcode, response-code]. 0x01 = success.
		if len(buf) >= 1 && buf[0] == 0x06 {
			if len(buf) >= 4 && buf[3] != 0x01 {
				racpDone <- fmt.Errorf("meter reported RACP error code %d", buf[3])
				return
			}
			racpDone <- nil
		}
	}); err != nil {
		return nil, fmt.Errorf("subscribe to RACP: %w", err)
	}

	if _, err := racpChar.Write(reportRecordsCmd(lastSeq)); err != nil {
		return nil, fmt.Errorf("write RACP report-records command: %w", err)
	}

	select {
	case err := <-racpDone:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("timed out waiting for meter records: %w", ctx.Err())
	}

	mu.Lock()
	defer mu.Unlock()
	inputs := make([]store.BloodSugarInput, 0, len(measurements))
	for _, m := range measurements {
		if !m.HasValue {
			continue
		}
		seq := int64(m.SeqNumber)
		inputs = append(inputs, store.BloodSugarInput{
			ValueMgdl: m.ValueMgdl,
			TakenAt:   m.Time.Format(time.RFC3339),
			Source:    "meter",
			SeqNumber: &seq,
		})
	}
	return inputs, nil
}

// reportRecordsCmd builds a Record Access Control Point "Report Stored Records"
// command (opcode 0x01). With no prior sync it asks for all records (operator
// 0x01); otherwise it asks for records with sequence number >= lastSeq+1
// (operator 0x03 "greater than or equal to", filter type 0x01 "sequence number").
func reportRecordsCmd(lastSeq int64) []byte {
	if lastSeq <= 0 {
		return []byte{0x01, 0x01}
	}
	next := uint16(lastSeq + 1)
	return []byte{0x01, 0x03, 0x01, byte(next), byte(next >> 8)}
}

// connect returns a connected Device, either by the configured address or by
// scanning for the meter by advertised name.
func connect(ctx context.Context) (bluetooth.Device, error) {
	if addrStr := os.Getenv("GTZY_METER_ADDR"); addrStr != "" {
		var addr bluetooth.Address
		addr.Set(addrStr)
		dev, err := adapter.Connect(addr, bluetooth.ConnectionParams{})
		if err != nil {
			return bluetooth.Device{}, fmt.Errorf("connect to meter %s (is it bonded and awake?): %w", addrStr, err)
		}
		return dev, nil
	}

	found := make(chan bluetooth.ScanResult, 1)
	scanErr := make(chan error, 1)
	var once sync.Once
	go func() {
		err := adapter.Scan(func(a *bluetooth.Adapter, r bluetooth.ScanResult) {
			name := strings.ToLower(r.LocalName())
			if strings.Contains(name, "accu-chek") || strings.Contains(name, "meter") {
				once.Do(func() {
					_ = a.StopScan()
					found <- r
				})
			}
		})
		if err != nil {
			scanErr <- err
		}
	}()

	select {
	case r := <-found:
		dev, err := adapter.Connect(r.Address, bluetooth.ConnectionParams{})
		if err != nil {
			return bluetooth.Device{}, fmt.Errorf("connect to scanned meter %s: %w", r.Address.String(), err)
		}
		return dev, nil
	case err := <-scanErr:
		return bluetooth.Device{}, fmt.Errorf("scan for meter: %w", err)
	case <-ctx.Done():
		_ = adapter.StopScan()
		return bluetooth.Device{}, fmt.Errorf("no glucose meter found (set GTZY_METER_ADDR to its MAC): %w", ctx.Err())
	}
}

// discover finds the Glucose Measurement and RACP characteristics on the device.
func discover(device bluetooth.Device) (meas, racp bluetooth.DeviceCharacteristic, err error) {
	services, err := device.DiscoverServices([]bluetooth.UUID{glucoseServiceUUID})
	if err != nil {
		return meas, racp, fmt.Errorf("discover glucose service: %w", err)
	}
	if len(services) == 0 {
		return meas, racp, fmt.Errorf("meter does not expose the Glucose service (0x1808)")
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{glucoseMeasUUID, racpUUID})
	if err != nil {
		return meas, racp, fmt.Errorf("discover glucose characteristics: %w", err)
	}

	var haveMeas, haveRACP bool
	for _, c := range chars {
		switch c.UUID() {
		case glucoseMeasUUID:
			meas, haveMeas = c, true
		case racpUUID:
			racp, haveRACP = c, true
		}
	}
	if !haveMeas || !haveRACP {
		return meas, racp, fmt.Errorf("meter missing required glucose characteristics (meas=%t racp=%t)", haveMeas, haveRACP)
	}
	return meas, racp, nil
}
