package unit

import (
	"database/sql"
	"strings"

	"github.com/nicolaics/pharmacon/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUnitByName(unitName string) (*types.Unit, error) {
	query := `SELECT COUNT(*) FROM unit WHERE name = ?`
	row := s.db.QueryRow(query, strings.ToUpper(unitName))
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int
	err := row.Scan(&count)

	if count == 0 {
		err = s.CreateUnit(unitName)
		if err != nil {
			return nil, err
		}
	}

	row = s.db.QueryRow("SELECT * FROM unit WHERE name = ?", strings.ToUpper(unitName))
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	unit, err := scanRowIntoUnit(row)
	if err != nil {
		return nil, err
	}

	return unit, nil
}

func (s *Store) GetUnitByID(id int) (*types.Unit, error) {
	row := s.db.QueryRow("SELECT * FROM unit WHERE id = ? ", id)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}
		return nil, row.Err()
	}

	unit, err := scanRowIntoUnit(row)
	if err != nil {
		return nil, err
	}

	return unit, nil
}

func (s *Store) CreateUnit(unitName string) error {
	_, err := s.db.Exec("INSERT INTO unit (name) VALUES (?)", strings.ToUpper(unitName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoUnit(row *sql.Row) (*types.Unit, error) {
	unit := new(types.Unit)

	err := row.Scan(
		&unit.ID,
		&unit.Name,
		&unit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	unit.CreatedAt = unit.CreatedAt.Local()

	return unit, nil
}
