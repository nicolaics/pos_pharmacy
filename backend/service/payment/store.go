package payment

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

func (s *Store) GetPaymentMethodByName(paymentMethodName string) (*types.PaymentMethod, error) {
	query := `SELECT COUNT(*) FROM payment_method WHERE name = ?`
	row := s.db.QueryRow(query, strings.ToUpper(paymentMethodName))
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int
	err := row.Scan(&count)

	if count == 0 {
		err = s.CreatePaymentMethod(paymentMethodName)
		if err != nil {
			return nil, err
		}
	}

	row = s.db.QueryRow("SELECT * FROM payment_method WHERE name = ? ", strings.ToUpper(paymentMethodName))
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	paymentMethod, err := scanRowIntoPaymentMethod(row)
	if err != nil {
		return nil, err
	}

	return paymentMethod, nil
}

func (s *Store) GetPaymentMethodByID(id int) (*types.PaymentMethod, error) {
	row := s.db.QueryRow("SELECT * FROM payment_method WHERE id = ? ", id)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}

		return nil, row.Err()
	}

	paymentMethod, err := scanRowIntoPaymentMethod(row)
	if err != nil {
		return nil, err
	}

	return paymentMethod, nil
}

func (s *Store) CreatePaymentMethod(paymentMethodName string) error {
	_, err := s.db.Exec("INSERT INTO payment_method (name) VALUES (?)", strings.ToUpper(paymentMethodName))

	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPaymentMethod(row *sql.Row) (*types.PaymentMethod, error) {
	paymentMethod := new(types.PaymentMethod)

	err := row.Scan(
		&paymentMethod.ID,
		&paymentMethod.Name,
		&paymentMethod.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	paymentMethod.CreatedAt = paymentMethod.CreatedAt.Local()

	return paymentMethod, nil
}
