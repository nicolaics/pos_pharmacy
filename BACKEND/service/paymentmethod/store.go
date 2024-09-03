package paymentmethod

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

func (s *Store) GetPaymentMethodByName(paymentMethodName string) (*types.PaymentMethod, error) {
	rows, err := s.db.Query("SELECT * FROM payment_method WHERE method = ? ", strings.ToUpper(paymentMethodName))
	if err != nil {
		return nil, err
	}

	paymentMethod := new(types.PaymentMethod)

	for rows.Next() {
		paymentMethod, err = scanRowIntoPaymentMethod(rows)

		if err != nil {
			return nil, err
		}
	}

	return paymentMethod, nil
}

func (s *Store) GetPaymentMethodByID(id int) (*types.PaymentMethod, error) {
	rows, err := s.db.Query("SELECT * FROM payment_method WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}

	paymentMethod := new(types.PaymentMethod)

	for rows.Next() {
		paymentMethod, err = scanRowIntoPaymentMethod(rows)

		if err != nil {
			return nil, err
		}
	}

	return paymentMethod, nil
}

func (s *Store) CreatePaymentMethod(paymentMethodName string) error {
	_, err := s.db.Exec("INSERT INTO payment_method (method) VALUES (?)", strings.ToUpper(paymentMethodName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPaymentMethod(rows *sql.Rows) (*types.PaymentMethod, error) {
	paymentMethod := new(types.PaymentMethod)

	err := rows.Scan(
		&paymentMethod.ID,
		&paymentMethod.Method,
		&paymentMethod.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	paymentMethod.CreatedAt = paymentMethod.CreatedAt.Local()

	return paymentMethod, nil
}
