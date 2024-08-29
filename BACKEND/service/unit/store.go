package unit

import (
	"database/sql"
	"strings"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUnitByName(unitName string) (*types.Unit, error) {
	rows, err := s.db.Query("SELECT * FROM unit WHERE unit = ? ", strings.ToUpper(unitName))
	if err != nil {
		return nil, err
	}

	unit := new(types.Unit)

	for rows.Next() {
		unit, err = scanRowIntoUnit(rows)

		if err != nil {
			return nil, err
		}
	}

	return unit, nil
}

func (s *Store) CreateUnit(unitName string) error {
	_, err := s.db.Exec("INSERT INTO unit (unit) VALUES (?)", strings.ToUpper(unitName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoUnit(rows *sql.Rows) (*types.Unit, error) {
	unit := new(types.Unit)

	err := rows.Scan(
		&unit.ID,
		&unit.Unit,
		&unit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	unit.CreatedAt = unit.CreatedAt.Local()

	return unit, nil
}
