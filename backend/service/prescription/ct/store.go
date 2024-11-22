package ct

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

func (s *Store) GetConsumeTimeByName(consumeTimeName string) (*types.ConsumeTime, error) {
	rows, err := s.db.Query("SELECT * FROM consume_time WHERE name = ? ", strings.ToLower(consumeTimeName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	consumeTime := new(types.ConsumeTime)

	for rows.Next() {
		consumeTime, err = scanRowIntoConsumeTime(rows)

		if err != nil {
			return nil, err
		}
	}

	if consumeTime.ID == 0 {
		return nil, fmt.Errorf("consume time not found")
	}

	return consumeTime, nil
}

func (s *Store) GetConsumeTimeByID(id int) (*types.ConsumeTime, error) {
	rows, err := s.db.Query("SELECT * FROM consume_time WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	consumeTime := new(types.ConsumeTime)

	for rows.Next() {
		consumeTime, err = scanRowIntoConsumeTime(rows)

		if err != nil {
			return nil, err
		}
	}

	return consumeTime, nil
}

func (s *Store) CreateConsumeTime(consumeTimeName string) error {
	_, err := s.db.Exec("INSERT INTO consume_time (name) VALUES (?)", strings.ToLower(consumeTimeName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoConsumeTime(rows *sql.Rows) (*types.ConsumeTime, error) {
	consumeTime := new(types.ConsumeTime)

	err := rows.Scan(
		&consumeTime.ID,
		&consumeTime.Name,
		&consumeTime.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	consumeTime.CreatedAt = consumeTime.CreatedAt.Local()

	return consumeTime, nil
}
