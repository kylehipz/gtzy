package store

import (
	"database/sql"

	"gtzy/internal/models"
)

type JournalStore struct{ DB *sql.DB }

const journalCols = `id, date, title, content, mood, created_at, updated_at`

func scanJournal(row interface{ Scan(...any) error }) (models.JournalEntry, error) {
	var e models.JournalEntry
	err := row.Scan(&e.ID, &e.Date, &e.Title, &e.Content, &e.Mood, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (s *JournalStore) List(date, from, to string) ([]models.JournalEntry, error) {
	q := `SELECT ` + journalCols + ` FROM journal_entries WHERE 1=1`
	var args []any
	if date != "" {
		q += ` AND date = ?`
		args = append(args, date)
	}
	if from != "" {
		q += ` AND date >= ?`
		args = append(args, from)
	}
	if to != "" {
		q += ` AND date <= ?`
		args = append(args, to)
	}
	q += ` ORDER BY date DESC, created_at DESC`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.JournalEntry{}
	for rows.Next() {
		e, err := scanJournal(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *JournalStore) Get(id int64) (models.JournalEntry, error) {
	row := s.DB.QueryRow(`SELECT `+journalCols+` FROM journal_entries WHERE id = ?`, id)
	return scanJournal(row)
}

type JournalInput struct {
	Date    string
	Title   string
	Content string
	Mood    *string
}

func (s *JournalStore) Create(in JournalInput) (models.JournalEntry, error) {
	now := NowUTC()
	res, err := s.DB.Exec(
		`INSERT INTO journal_entries (date, title, content, mood, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		in.Date, in.Title, in.Content, in.Mood, now, now,
	)
	if err != nil {
		return models.JournalEntry{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.JournalEntry{}, err
	}
	return s.Get(id)
}

type JournalPatch struct {
	Title   *string
	Content *string
	Mood    **string
}

func (s *JournalStore) Update(id int64, p JournalPatch) (models.JournalEntry, error) {
	if p.Title != nil {
		if _, err := s.DB.Exec(`UPDATE journal_entries SET title = ? WHERE id = ?`, *p.Title, id); err != nil {
			return models.JournalEntry{}, err
		}
	}
	if p.Content != nil {
		if _, err := s.DB.Exec(`UPDATE journal_entries SET content = ? WHERE id = ?`, *p.Content, id); err != nil {
			return models.JournalEntry{}, err
		}
	}
	if p.Mood != nil {
		if _, err := s.DB.Exec(`UPDATE journal_entries SET mood = ? WHERE id = ?`, *p.Mood, id); err != nil {
			return models.JournalEntry{}, err
		}
	}
	if _, err := s.DB.Exec(`UPDATE journal_entries SET updated_at = ? WHERE id = ?`, NowUTC(), id); err != nil {
		return models.JournalEntry{}, err
	}
	return s.Get(id)
}

func (s *JournalStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM journal_entries WHERE id = ?`, id)
	return err
}
