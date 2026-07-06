package store

import (
	"database/sql"

	"gtzy/internal/models"
)

type CategoryStore struct{ DB *sql.DB }

func (s *CategoryStore) List() ([]models.Category, error) {
	rows, err := s.DB.Query(`SELECT id, name, color, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s *CategoryStore) Create(name, color string) (models.Category, error) {
	if color == "" {
		color = "mauve"
	}
	now := NowUTC()
	res, err := s.DB.Exec(`INSERT INTO categories (name, color, created_at) VALUES (?, ?, ?)`, name, color, now)
	if err != nil {
		return models.Category{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Category{}, err
	}
	return models.Category{ID: id, Name: name, Color: color, CreatedAt: now}, nil
}

func (s *CategoryStore) Update(id int64, name, color *string) (models.Category, error) {
	if name != nil {
		if _, err := s.DB.Exec(`UPDATE categories SET name = ? WHERE id = ?`, *name, id); err != nil {
			return models.Category{}, err
		}
	}
	if color != nil {
		if _, err := s.DB.Exec(`UPDATE categories SET color = ? WHERE id = ?`, *color, id); err != nil {
			return models.Category{}, err
		}
	}
	var c models.Category
	err := s.DB.QueryRow(`SELECT id, name, color, created_at FROM categories WHERE id = ?`, id).
		Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt)
	return c, err
}

func (s *CategoryStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}
