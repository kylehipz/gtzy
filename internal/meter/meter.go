package meter

import (
	"context"
	"database/sql"
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

// adapter is the process-wide BLE adapter (BlueZ on Linux). bleMu serializes all
// access to it: the background Watcher and any concurrent manual sync must not
// scan/connect at the same time. The Watcher releases bleMu between scan cycles
// (see watcher.go) so a manual sync is never starved.
var (
	adapter = bluetooth.DefaultAdapter
	bleMu   sync.Mutex
)

// scanWindow is how long a single sync attempt waits for the meter to advertise
// before giving up for this cycle. The meter advertises when you press save, so
// this is the window in which we can catch it.
const scanWindow = 8 * time.Second

// SyncInto scans for the meter, and if it is currently advertising, pulls records
// newer than what is already stored and inserts them. Returns how many records
// were fetched from the meter and how many were newly inserted (dedup drops the
// rest). When the meter is not present it returns (0, 0, nil) — that is the
// normal idle case, not an error. Holds bleMu for the whole scan+pull so it is
// safe to call from both the HTTP handler and the Watcher.
func SyncInto(ctx context.Context, db *sql.DB) (fetched, inserted int, err error) {
	bleMu.Lock()
	defer bleMu.Unlock()

	bs := &store.BloodSugarStore{DB: db}
	lastSeq, err := bs.MaxMeterSeq()
	if err != nil {
		return 0, 0, err
	}

	addr, found, err := findMeter(ctx, scanWindow)
	if err != nil {
		return 0, 0, err
	}
	if !found {
		return 0, 0, nil
	}

	inputs, err := pull(ctx, addr, lastSeq)
	if err != nil {
		return 0, 0, err
	}
	n, err := bs.CreateMany(inputs)
	if err != nil {
		return len(inputs), 0, err
	}
	return len(inputs), n, nil
}

// findMeter scans for the paired meter and returns its address once it is seen
// advertising. It matches by GTZY_METER_ADDR when set, otherwise by an advertised
// name containing "accu-chek". If the window elapses with no match it returns
// found=false and a nil error (the meter is simply asleep / out of range).
//
// adapter.Scan blocks until StopScan is called, so this waits for Scan to fully
// return before it returns — leaving the adapter idle and safe for the next
// operation (a subsequent scan or connect). Returning early while a scan is
// still tearing down causes BlueZ "operation already in progress" errors.
func findMeter(ctx context.Context, window time.Duration) (bluetooth.Address, bool, error) {
	if err := adapter.Enable(); err != nil {
		return bluetooth.Address{}, false, fmt.Errorf("enable bluetooth adapter: %w", err)
	}
	wantAddr := strings.ToLower(os.Getenv("GTZY_METER_ADDR"))

	scanCtx, cancel := context.WithTimeout(ctx, window)
	defer cancel()

	// Stop the blocking scan once we match or the window elapses.
	go func() {
		<-scanCtx.Done()
		_ = adapter.StopScan()
	}()

	found := make(chan bluetooth.Address, 1)
	// We always stop the scan ourselves (on match or timeout), so Scan's returned
	// error is the expected "scan stopped" signal and is intentionally ignored.
	_ = adapter.Scan(func(a *bluetooth.Adapter, r bluetooth.ScanResult) {
		match := false
		if wantAddr != "" {
			match = strings.ToLower(r.Address.String()) == wantAddr
		} else {
			name := strings.ToLower(r.LocalName())
			match = strings.Contains(name, "accu-chek") || strings.Contains(name, "meter")
		}
		if match {
			select {
			case found <- r.Address:
				cancel() // trigger StopScan
			default:
			}
		}
	})

	select {
	case addr := <-found:
		return addr, true, nil
	default:
		return bluetooth.Address{}, false, nil
	}
}

// pull connects to an advertising meter and reads its stored records with
// sequence number greater than lastSeq (or all records when lastSeq <= 0),
// decoding each per the Bluetooth Glucose Profile.
func pull(ctx context.Context, addr bluetooth.Address, lastSeq int64) ([]store.BloodSugarInput, error) {
	device, err := adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("connect to meter %s (is it awake and bonded?): %w", addr.String(), err)
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
