package det

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetDetByName(detName string) (*types.Det, error) {
	rows, err := s.db.Query("SELECT * FROM det WHERE name = ? ", strings.ToLower(detName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	det := new(types.Det)

	for rows.Next() {
		det, err = scanRowIntoDet(rows)

		if err != nil {
			return nil, err
		}
	}

	if det.ID == 0 {
		return nil, fmt.Errorf("det not found")
	}

	return det, nil
}

func (s *Store) GetDetByID(id int) (*types.Det, error) {
	rows, err := s.db.Query("SELECT * FROM det WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	det := new(types.Det)

	for rows.Next() {
		det, err = scanRowIntoDet(rows)

		if err != nil {
			return nil, err
		}
	}

	return det, nil
}


func (s *Store) CreateDet(detName string) error {
	_, err := s.db.Exec("INSERT INTO det (name) VALUES (?)", strings.ToLower(detName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoDet(rows *sql.Rows) (*types.Det, error) {
	det := new(types.Det)

	err := rows.Scan(
		&det.ID,
		&det.Name,
		&det.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	det.CreatedAt = det.CreatedAt.Local()

	return det, nil
}
