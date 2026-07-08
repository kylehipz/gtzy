// Package meter reads stored records off a Bluetooth Low Energy blood-glucose
// meter that implements the standard Bluetooth Glucose Profile (service 0x1808).
// gtzy acts as the BLE central: it connects to the (already-bonded) meter, asks
// for stored records via the Record Access Control Point, and decodes the
// Glucose Measurement notifications the meter streams back.
package meter

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// measurement is one decoded Glucose Measurement record (characteristic 0x2A18).
type measurement struct {
	SeqNumber uint16
	Time      time.Time
	ValueMgdl float64
	HasValue  bool
}

// mgdlPerKgL converts a glucose concentration in kg/L (the SFLOAT unit the
// Glucose Profile uses when the units flag is 0) to mg/dL. 1 mg/dL = 1e-5 kg/L.
const mgdlPerKgL = 100000.0

// mgdlPerMolL converts glucose mol/L (units flag = 1) to mg/dL:
// 1 mmol/L glucose = 18.0182 mg/dL, and 1 mol/L = 1000 mmol/L.
const mgdlPerMolL = 18018.2

// parseMeasurement decodes a Glucose Measurement (0x2A18) value per the
// Bluetooth GLS spec. Byte layout: flags(1), seq(2 LE), base time(7),
// [time offset(2 LE) if flags bit0], [concentration SFLOAT(2) + type/location(1)
// if flags bit1], [sensor status(2) if flags bit3].
func parseMeasurement(b []byte) (measurement, error) {
	if len(b) < 10 {
		return measurement{}, fmt.Errorf("glucose measurement too short: %d bytes", len(b))
	}
	flags := b[0]
	var m measurement
	m.SeqNumber = binary.LittleEndian.Uint16(b[1:3])

	t, err := parseDateTime(b[3:10])
	if err != nil {
		return measurement{}, err
	}
	m.Time = t
	off := 10

	// Time Offset Present (bit 0): minutes to add to base time.
	if flags&0x01 != 0 {
		if len(b) < off+2 {
			return measurement{}, fmt.Errorf("glucose measurement truncated at time offset")
		}
		minutes := int16(binary.LittleEndian.Uint16(b[off : off+2]))
		m.Time = m.Time.Add(time.Duration(minutes) * time.Minute)
		off += 2
	}

	// Glucose Concentration present (bit 1): SFLOAT + type/location byte.
	if flags&0x02 != 0 {
		if len(b) < off+3 {
			return measurement{}, fmt.Errorf("glucose measurement truncated at concentration")
		}
		raw := binary.LittleEndian.Uint16(b[off : off+2])
		val, ok := sfloatToFloat(raw)
		if ok {
			if flags&0x04 != 0 { // units flag: 1 = mol/L
				m.ValueMgdl = val * mgdlPerMolL
			} else { // 0 = kg/L
				m.ValueMgdl = val * mgdlPerKgL
			}
			m.ValueMgdl = math.Round(m.ValueMgdl)
			m.HasValue = true
		}
	}

	return m, nil
}

// parseDateTime decodes an org.bluetooth.date_time (7 bytes) as local wall time.
func parseDateTime(b []byte) (time.Time, error) {
	if len(b) < 7 {
		return time.Time{}, fmt.Errorf("date_time too short: %d bytes", len(b))
	}
	year := int(binary.LittleEndian.Uint16(b[0:2]))
	month := time.Month(b[2])
	day := int(b[3])
	hour := int(b[4])
	minute := int(b[5])
	sec := int(b[6])
	if year == 0 || month == 0 || day == 0 {
		return time.Time{}, fmt.Errorf("date_time has unknown fields (year=%d month=%d day=%d)", year, month, day)
	}
	return time.Date(year, month, day, hour, minute, sec, 0, time.Local), nil
}

// sfloatToFloat decodes an IEEE-11073 16-bit SFLOAT (4-bit signed exponent +
// 12-bit signed mantissa). The second return is false for the reserved special
// values (NaN, +/-INFINITY, NRes, Reserved).
func sfloatToFloat(raw uint16) (float64, bool) {
	rawMantissa := raw & 0x0FFF
	switch rawMantissa {
	case 0x07FF, // NaN
		0x0800, // NRes
		0x07FE, // +INFINITY
		0x0802, // -INFINITY
		0x0801: // Reserved
		return 0, false
	}

	mantissa := int32(rawMantissa)
	if mantissa >= 0x0800 { // sign-extend 12-bit
		mantissa -= 0x1000
	}
	exponent := int32(raw >> 12)
	if exponent >= 0x08 { // sign-extend 4-bit
		exponent -= 0x10
	}
	return float64(mantissa) * math.Pow(10, float64(exponent)), true
}
