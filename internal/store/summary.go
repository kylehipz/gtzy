package store

import (
	"database/sql"

	"gtzy/internal/models"
)

type SummaryStore struct{ DB *sql.DB }

func (s *SummaryStore) Get(periodType, periodKey string) (models.AISummary, bool, error) {
	var sum models.AISummary
	err := s.DB.QueryRow(
		`SELECT id, period_type, period_key, content, model, created_at
		 FROM ai_summaries WHERE period_type = ? AND period_key = ?`,
		periodType, periodKey,
	).Scan(&sum.ID, &sum.PeriodType, &sum.PeriodKey, &sum.Content, &sum.Model, &sum.CreatedAt)
	if err == sql.ErrNoRows {
		return models.AISummary{}, false, nil
	}
	if err != nil {
		return models.AISummary{}, false, err
	}
	return sum, true, nil
}

func (s *SummaryStore) Upsert(periodType, periodKey, content, model string) (models.AISummary, error) {
	now := NowUTC()
	_, err := s.DB.Exec(
		`INSERT INTO ai_summaries (period_type, period_key, content, model, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (period_type, period_key)
		 DO UPDATE SET content = excluded.content, model = excluded.model, created_at = excluded.created_at`,
		periodType, periodKey, content, model, now,
	)
	if err != nil {
		return models.AISummary{}, err
	}
	sum, _, err := s.Get(periodType, periodKey)
	return sum, err
}
