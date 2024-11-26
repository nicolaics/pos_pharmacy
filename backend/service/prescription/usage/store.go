package su

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nicolaics/pharmacon/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUsageByName(usageName string) (*types.Usage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE name = ? ", strings.ToLower(usageName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := new(types.Usage)

	for rows.Next() {
		usage, err = scanRowIntoUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	if usage.ID == 0 {
		return nil, fmt.Errorf("prescription set usage not found")
	}

	return usage, nil
}

func (s *Store) GetUsageByID(id int) (*types.Usage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := new(types.Usage)

	for rows.Next() {
		usage, err = scanRowIntoUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	return usage, nil
}

func (s *Store) CreateUsage(usageName string) error {
	_, err := s.db.Exec("INSERT INTO prescription_set_usage (name) VALUES (?)", strings.ToLower(usageName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoUsage(rows *sql.Rows) (*types.Usage, error) {
	usage := new(types.Usage)

	err := rows.Scan(
		&usage.ID,
		&usage.Name,
		&usage.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	usage.CreatedAt = usage.CreatedAt.Local()

	return usage, nil
}
