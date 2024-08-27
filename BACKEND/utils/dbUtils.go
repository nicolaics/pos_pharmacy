package utils

import (
	"database/sql"

	"github.com/nicolaics/pos_pharmacy/types"
)

func FindCustomerID(db *sql.DB, customerName string) (int, error) {
	rows, err := db.Query("SELECT * FROM customer WHERE name = ? ", customerName)

	if err != nil {
		return -1, err
	}

	customer := new(types.Customer)

	for rows.Next() {
		customer, err = ScanRowIntoCustomer(rows)

		if err != nil {
			return -1, err
		}
	}

	return customer.ID, nil
}

func ScanRowIntoCustomer(rows *sql.Rows) (*types.Customer, error) {
	customer := new(types.Customer)

	err := rows.Scan(
		&customer.ID,
		&customer.Name,
		&customer.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return customer, nil
}

func FindCashierID(db *sql.DB, cashierName string) (int, error) {
	rows, err := db.Query("SELECT * FROM cashier WHERE name = ? ", cashierName)

	if err != nil {
		return -1, err
	}

	cashier := new(types.Cashier)

	for rows.Next() {
		cashier, err = ScanRowIntoCashier(rows)

		if err != nil {
			return -1, err
		}
	}

	return cashier.ID, nil
}

func ScanRowIntoCashier(rows *sql.Rows) (*types.Cashier, error) {
	cashier := new(types.Cashier)

	err := rows.Scan(
		&cashier.ID,
		&cashier.Name,
		&cashier.Password,
		&cashier.Admin,
		&cashier.CreatedAt,
		&cashier.LastLoggedIn,
	)

	if err != nil {
		return nil, err
	}

	return cashier, nil
}

func FindPaymentMethodID(db *sql.DB, paymentMethodName string) (int, error) {
	rows, err := db.Query("SELECT * FROM payment_method WHERE method = ? ", paymentMethodName)

	if err != nil {
		return -1, err
	}

	paymentMethod := new(types.PaymentMethod)

	for rows.Next() {
		paymentMethod, err = ScanRowIntoPaymentMethod(rows)

		if err != nil {
			return -1, err
		}
	}

	return paymentMethod.ID, nil
}

func ScanRowIntoPaymentMethod(rows *sql.Rows) (*types.PaymentMethod, error) {
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

func ScanRowIntoSupplier(rows *sql.Rows) (*types.Supplier, error) {
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
