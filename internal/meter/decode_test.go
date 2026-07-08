package meter

import (
	"math"
	"testing"
	"time"
)

func TestParseMeasurement_KgL(t *testing.T) {
	// flags 0x02 (concentration present, units kg/L, no time offset);
	// seq 5; 2026-07-08 08:30:00; SFLOAT 0.001 kg/L (mantissa 100, exp -5) = 100 mg/dL.
	b := []byte{
		0x02,       // flags
		0x05, 0x00, // seq = 5
		0xEA, 0x07, // year 2026
		0x07,       // month
		0x08,       // day
		0x08,       // hour
		0x1E,       // minute 30
		0x00,       // second
		0x64, 0xB0, // SFLOAT 0xB064 => 100e-5 = 0.001
		0x00, // type/sample location
	}
	m, err := parseMeasurement(b)
	if err != nil {
		t.Fatalf("parseMeasurement: %v", err)
	}
	if m.SeqNumber != 5 {
		t.Errorf("seq = %d, want 5", m.SeqNumber)
	}
	if !m.HasValue {
		t.Fatal("HasValue = false, want true")
	}
	if m.ValueMgdl != 100 {
		t.Errorf("value = %v mg/dL, want 100", m.ValueMgdl)
	}
	want := time.Date(2026, 7, 8, 8, 30, 0, 0, time.Local)
	if !m.Time.Equal(want) {
		t.Errorf("time = %v, want %v", m.Time, want)
	}
}

func TestParseMeasurement_TimeOffset(t *testing.T) {
	// flags 0x03: time offset present (+60 min) + concentration present.
	b := []byte{
		0x03,
		0x01, 0x00, // seq 1
		0xEA, 0x07, 0x07, 0x08, 0x08, 0x00, 0x00, // 2026-07-08 08:00:00
		0x3C, 0x00, // time offset +60 minutes
		0x64, 0xB0, // 100 mg/dL
		0x00,
	}
	m, err := parseMeasurement(b)
	if err != nil {
		t.Fatalf("parseMeasurement: %v", err)
	}
	want := time.Date(2026, 7, 8, 9, 0, 0, 0, time.Local)
	if !m.Time.Equal(want) {
		t.Errorf("time = %v, want %v (base + 60m offset)", m.Time, want)
	}
}

func TestParseMeasurement_NoConcentration(t *testing.T) {
	// flags 0x00: no optional fields present.
	b := []byte{
		0x00,
		0x02, 0x00,
		0xEA, 0x07, 0x07, 0x08, 0x08, 0x00, 0x00,
	}
	m, err := parseMeasurement(b)
	if err != nil {
		t.Fatalf("parseMeasurement: %v", err)
	}
	if m.HasValue {
		t.Error("HasValue = true, want false when no concentration present")
	}
}

func TestSfloatToFloat(t *testing.T) {
	cases := []struct {
		raw  uint16
		want float64
		ok   bool
	}{
		{0xB064, 0.001, true}, // mantissa 100, exp -5
		{0x0064, 100, true},   // mantissa 100, exp 0
		{0x07FF, 0, false},    // NaN
		{0x0800, 0, false},    // NRes
		{0x07FE, 0, false},    // +INFINITY
	}
	for _, c := range cases {
		got, ok := sfloatToFloat(c.raw)
		if ok != c.ok {
			t.Errorf("sfloatToFloat(%#04x) ok = %v, want %v", c.raw, ok, c.ok)
			continue
		}
		if ok && math.Abs(got-c.want) > 1e-9 {
			t.Errorf("sfloatToFloat(%#04x) = %v, want %v", c.raw, got, c.want)
		}
	}
}

func TestReportRecordsCmd(t *testing.T) {
	if got := reportRecordsCmd(0); string(got) != string([]byte{0x01, 0x01}) {
		t.Errorf("reportRecordsCmd(0) = %v, want [1 1] (all records)", got)
	}
	// lastSeq 5 => request seq >= 6 (operator 0x03, filter 0x01, operand 6 LE).
	got := reportRecordsCmd(5)
	want := []byte{0x01, 0x03, 0x01, 0x06, 0x00}
	if string(got) != string(want) {
		t.Errorf("reportRecordsCmd(5) = %v, want %v", got, want)
	}
}
