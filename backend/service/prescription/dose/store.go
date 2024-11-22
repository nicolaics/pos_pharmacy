package dose

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

func (s *Store) GetDoseByName(doseName string) (*types.Dose, error) {
	rows, err := s.db.Query("SELECT * FROM dose WHERE name = ? ", strings.ToLower(doseName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dose := new(types.Dose)

	for rows.Next() {
		dose, err = scanRowIntoDose(rows)

		if err != nil {
			return nil, err
		}
	}

	if dose.ID == 0 {
		return nil, fmt.Errorf("dose not found")
	}

	return dose, nil
}

func (s *Store) GetDoseByID(id int) (*types.Dose, error) {
	rows, err := s.db.Query("SELECT * FROM dose WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dose := new(types.Dose)

	for rows.Next() {
		dose, err = scanRowIntoDose(rows)

		if err != nil {
			return nil, err
		}
	}

	return dose, nil
}

func (s *Store) CreateDose(doseName string) error {
	_, err := s.db.Exec("INSERT INTO dose (name) VALUES (?)", strings.ToLower(doseName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoDose(rows *sql.Rows) (*types.Dose, error) {
	dose := new(types.Dose)

	err := rows.Scan(
		&dose.ID,
		&dose.Name,
		&dose.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	dose.CreatedAt = dose.CreatedAt.Local()

	return dose, nil
}
