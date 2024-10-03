package consumeway

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

func (s *Store) GetConsumeWayByName(consumeWayName string) (*types.ConsumeWay, error) {
	rows, err := s.db.Query("SELECT * FROM consume_way WHERE name = ? ", strings.ToLower(consumeWayName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	consumeWay := new(types.ConsumeWay)

	for rows.Next() {
		consumeWay, err = scanRowIntoConsumeWay(rows)

		if err != nil {
			return nil, err
		}
	}

	if consumeWay.ID == 0 {
		return nil, fmt.Errorf("consume way not found")
	}

	return consumeWay, nil
}

func (s *Store) GetConsumeWayByID(id int) (*types.ConsumeWay, error) {
	rows, err := s.db.Query("SELECT * FROM consume_way WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	consumeWay := new(types.ConsumeWay)

	for rows.Next() {
		consumeWay, err = scanRowIntoConsumeWay(rows)

		if err != nil {
			return nil, err
		}
	}

	return consumeWay, nil
}


func (s *Store) CreateConsumeWay(consumeWayName string) error {
	_, err := s.db.Exec("INSERT INTO consume_way (name) VALUES (?)", strings.ToLower(consumeWayName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoConsumeWay(rows *sql.Rows) (*types.ConsumeWay, error) {
	consumeWay := new(types.ConsumeWay)

	err := rows.Scan(
		&consumeWay.ID,
		&consumeWay.Name,
		&consumeWay.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	consumeWay.CreatedAt = consumeWay.CreatedAt.Local()

	return consumeWay, nil
}
