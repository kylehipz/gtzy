package store

import (
	"database/sql"

	"gtzy/internal/models"
)

type BloodSugarStore struct{ DB *sql.DB }

const bloodSugarCols = `id, value_mgdl, taken_at, meal_tag, notes, source, seq_number, created_at`

func scanBloodSugar(row interface{ Scan(...any) error }) (models.BloodSugarReading, error) {
	var r models.BloodSugarReading
	err := row.Scan(&r.ID, &r.ValueMgdl, &r.TakenAt, &r.MealTag, &r.Notes, &r.Source, &r.SeqNumber, &r.CreatedAt)
	return r, err
}

// List returns readings whose taken_at falls within [from, to] (RFC3339 or
// YYYY-MM-DD bounds; empty bounds are unbounded), newest first.
func (s *BloodSugarStore) List(from, to string) ([]models.BloodSugarReading, error) {
	q := `SELECT ` + bloodSugarCols + ` FROM blood_sugar_readings WHERE 1=1`
	var args []any
	if from != "" {
		q += ` AND taken_at >= ?`
		args = append(args, from)
	}
	if to != "" {
		q += ` AND taken_at <= ?`
		args = append(args, to)
	}
	q += ` ORDER BY taken_at DESC, id DESC`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []models.BloodSugarReading{}
	for rows.Next() {
		r, err := scanBloodSugar(rows)
		if err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// RecentMeter returns the most recently inserted meter-sourced readings (highest
// ids first), up to limit. Used to describe what a background sync just imported.
func (s *BloodSugarStore) RecentMeter(limit int) ([]models.BloodSugarReading, error) {
	rows, err := s.DB.Query(
		`SELECT `+bloodSugarCols+` FROM blood_sugar_readings WHERE source = 'meter' ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []models.BloodSugarReading{}
	for rows.Next() {
		r, err := scanBloodSugar(rows)
		if err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

func (s *BloodSugarStore) Get(id int64) (models.BloodSugarReading, error) {
	row := s.DB.QueryRow(`SELECT `+bloodSugarCols+` FROM blood_sugar_readings WHERE id = ?`, id)
	return scanBloodSugar(row)
}

// MaxMeterSeq returns the highest sequence number already synced from the meter,
// or 0 if none. Used to request only newer records on the next sync.
func (s *BloodSugarStore) MaxMeterSeq() (int64, error) {
	var seq sql.NullInt64
	err := s.DB.QueryRow(
		`SELECT MAX(seq_number) FROM blood_sugar_readings WHERE source = 'meter' AND seq_number IS NOT NULL`,
	).Scan(&seq)
	if err != nil {
		return 0, err
	}
	if !seq.Valid {
		return 0, nil
	}
	return seq.Int64, nil
}

type BloodSugarInput struct {
	ValueMgdl float64
	TakenAt   string
	MealTag   string
	Notes     string
	Source    string
	SeqNumber *int64
}

func (s *BloodSugarStore) Create(in BloodSugarInput) (models.BloodSugarReading, error) {
	source := in.Source
	if source == "" {
		source = "manual"
	}
	now := NowUTC()
	res, err := s.DB.Exec(
		`INSERT INTO blood_sugar_readings (value_mgdl, taken_at, meal_tag, notes, source, seq_number, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.ValueMgdl, in.TakenAt, in.MealTag, in.Notes, source, in.SeqNumber, now,
	)
	if err != nil {
		return models.BloodSugarReading{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.BloodSugarReading{}, err
	}
	return s.Get(id)
}

// CreateMany inserts meter-sourced readings, skipping any whose seq_number is
// already present (dedup enforced by idx_bsr_meter_seq). Returns the number of
// rows actually inserted.
func (s *BloodSugarStore) CreateMany(inputs []BloodSugarInput) (int, error) {
	inserted := 0
	now := NowUTC()
	for _, in := range inputs {
		source := in.Source
		if source == "" {
			source = "meter"
		}
		res, err := s.DB.Exec(
			`INSERT INTO blood_sugar_readings (value_mgdl, taken_at, meal_tag, notes, source, seq_number, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (seq_number) WHERE source = 'meter' AND seq_number IS NOT NULL DO NOTHING`,
			in.ValueMgdl, in.TakenAt, in.MealTag, in.Notes, source, in.SeqNumber, now,
		)
		if err != nil {
			return inserted, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	return inserted, nil
}

type BloodSugarPatch struct {
	ValueMgdl *float64
	TakenAt   *string
	MealTag   *string
	Notes     *string
}

func (s *BloodSugarStore) Update(id int64, p BloodSugarPatch) (models.BloodSugarReading, error) {
	if p.ValueMgdl != nil {
		if _, err := s.DB.Exec(`UPDATE blood_sugar_readings SET value_mgdl = ? WHERE id = ?`, *p.ValueMgdl, id); err != nil {
			return models.BloodSugarReading{}, err
		}
	}
	if p.TakenAt != nil {
		if _, err := s.DB.Exec(`UPDATE blood_sugar_readings SET taken_at = ? WHERE id = ?`, *p.TakenAt, id); err != nil {
			return models.BloodSugarReading{}, err
		}
	}
	if p.MealTag != nil {
		if _, err := s.DB.Exec(`UPDATE blood_sugar_readings SET meal_tag = ? WHERE id = ?`, *p.MealTag, id); err != nil {
			return models.BloodSugarReading{}, err
		}
	}
	if p.Notes != nil {
		if _, err := s.DB.Exec(`UPDATE blood_sugar_readings SET notes = ? WHERE id = ?`, *p.Notes, id); err != nil {
			return models.BloodSugarReading{}, err
		}
	}
	return s.Get(id)
}

func (s *BloodSugarStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM blood_sugar_readings WHERE id = ?`, id)
	return err
}
