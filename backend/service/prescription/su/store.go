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

func (s *Store) GetSetUsageByName(SetUsageName string) (*types.SetUsage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE name = ? ", strings.ToLower(SetUsageName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	SetUsage := new(types.SetUsage)

	for rows.Next() {
		SetUsage, err = scanRowIntoSetUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	if SetUsage.ID == 0 {
		return nil, fmt.Errorf("prescription set usage not found")
	}

	return SetUsage, nil
}

func (s *Store) GetSetUsageByID(id int) (*types.SetUsage, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_usage WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	SetUsage := new(types.SetUsage)

	for rows.Next() {
		SetUsage, err = scanRowIntoSetUsage(rows)

		if err != nil {
			return nil, err
		}
	}

	return SetUsage, nil
}

func (s *Store) CreateSetUsage(SetUsageName string) error {
	_, err := s.db.Exec("INSERT INTO prescription_set_usage (name) VALUES (?)", strings.ToLower(SetUsageName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoSetUsage(rows *sql.Rows) (*types.SetUsage, error) {
	SetUsage := new(types.SetUsage)

	err := rows.Scan(
		&SetUsage.ID,
		&SetUsage.Name,
		&SetUsage.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	SetUsage.CreatedAt = SetUsage.CreatedAt.Local()

	return SetUsage, nil
}
