package paymentmethod

import (
	"database/sql"
	_ "fmt"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) FindPaymentMethodID(db *sql.DB, paymentMethodName string) (int, error) {
	rows, err := db.Query("SELECT * FROM payment_method WHERE method = ? ", paymentMethodName)

	if err != nil {
		return -1, err
	}

	paymentMethod := new(types.PaymentMethod)

	for rows.Next() {
		paymentMethod, err = scanRowIntoPaymentMethod(rows)

		if err != nil {
			return -1, err
		}
	}

	return paymentMethod.ID, nil
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

	return paymentMethod, nil
}