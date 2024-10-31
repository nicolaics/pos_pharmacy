package mf

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

func (s *Store) GetMfByName(mfName string) (*types.Mf, error) {
	rows, err := s.db.Query("SELECT * FROM mf WHERE name = ? ", strings.ToLower(mfName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mf := new(types.Mf)

	for rows.Next() {
		mf, err = scanRowIntoMf(rows)

		if err != nil {
			return nil, err
		}
	}

	if mf.ID == 0 {
		return nil, fmt.Errorf("mf not found")
	}

	return mf, nil
}

func (s *Store) GetMfByID(id int) (*types.Mf, error) {
	rows, err := s.db.Query("SELECT * FROM mf WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mf := new(types.Mf)

	for rows.Next() {
		mf, err = scanRowIntoMf(rows)

		if err != nil {
			return nil, err
		}
	}

	return mf, nil
}


func (s *Store) CreateMf(mfName string) error {
	_, err := s.db.Exec("INSERT INTO mf (name) VALUES (?)", strings.ToLower(mfName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoMf(rows *sql.Rows) (*types.Mf, error) {
	mf := new(types.Mf)

	err := rows.Scan(
		&mf.ID,
		&mf.Name,
		&mf.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	mf.CreatedAt = mf.CreatedAt.Local()

	return mf, nil
}
