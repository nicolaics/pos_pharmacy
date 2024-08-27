package supplier

import (
	"database/sql"
	"fmt"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetSupplierByName(name string) (*types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE name = ? ", name)

	if err != nil {
		return nil, err
	}

	supplier := new(types.Supplier)

	for rows.Next() {
		supplier, err = scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}
	}

	if supplier.ID == 0 {
		return nil, fmt.Errorf("supplier not found")
	}

	return supplier, nil
}

func (s *Store) GetSupplierByID(id int) (*types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	supplier := new(types.Supplier)

	for rows.Next() {
		supplier, err = scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}
	}

	if supplier.ID == 0 {
		return nil, fmt.Errorf("supplier not found")
	}

	return supplier, nil
}

func (s *Store) CreateSupplier(supplier types.Supplier) error {
	_, err := s.db.Exec("INSERT INTO supplier (name) VALUES (?)",
						supplier.Name)

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoSupplier(rows *sql.Rows) (*types.Supplier, error) {
	supplier := new(types.Supplier)

	err := rows.Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return supplier, nil
}