package customer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetCustomerByName(name string) (*types.Customer, error) {
	query := "SELECT * FROM customer WHERE name = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, strings.ToUpper(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customer := new(types.Customer)

	for rows.Next() {
		customer, err = scanRowIntoCustomer(rows)

		if err != nil {
			return nil, err
		}
	}

	if customer.ID == 0 {
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

func (s *Store) GetCustomerByID(id int) (*types.Customer, error) {
	query := "SELECT * FROM customer WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customer := new(types.Customer)

	for rows.Next() {
		customer, err = scanRowIntoCustomer(rows)

		if err != nil {
			return nil, err
		}
	}

	if customer.ID == 0 {
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

func (s *Store) CreateCustomer(customer types.Customer) error {
	_, err := s.db.Exec("INSERT INTO customer (name) VALUES (?)",
						strings.ToUpper(customer.Name))

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllCustomers() ([]types.Customer, error) {
	rows, err := s.db.Query("SELECT * FROM customer WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]types.Customer, 0)

	for rows.Next() {
		customer, err := scanRowIntoCustomer(rows)

		if err != nil {
			return nil, err
		}

		customers = append(customers, *customer)
	}

	return customers, nil
}

func (s *Store) DeleteCustomer(uid int, customer *types.Customer) error {
	query := "UPDATE customer SET deleted_at ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), uid, customer.ID)
	if err != nil {
		return err
	}

	data, err := s.GetCustomerByID(customer.ID)	
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "customer", uid, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) ModifyCustomer(id int, newName string, uid int) error {
	data, err := s.GetCustomerByID(id)
	if err != nil {
		return err
	}

	err = logger.WriteLog("modify", "customer", uid, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("UPDATE customer SET name = ? WHERE id = ? ", strings.ToUpper(newName), id)

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoCustomer(rows *sql.Rows) (*types.Customer, error) {
	customer := new(types.Customer)

	err := rows.Scan(
		&customer.ID,
		&customer.Name,
		&customer.CreatedAt,
		&customer.DeletedAt,
		&customer.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	customer.CreatedAt = customer.CreatedAt.Local()

	return customer, nil
}
