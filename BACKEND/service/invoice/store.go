package invoice

import (
	"database/sql"
	_ "fmt"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateInvoice(invoice types.Invoice) error {
	cashierID, err := utils.FindCashierID(s.db, invoice.CashierName)
	if err != nil {
		return err
	}

	customerID, err := utils.FindCustomerID(s.db, invoice.CustomerName)
	if err != nil {
		return err
	}

	paymentMethodID, err := utils.FindPaymentMethodID(s.db, invoice.PaymentMethodName)
	if err != nil {
		return err
	}

	cmd := `INSERT INTO invoice
			(number, cashier_name_id, customer_id, subtotal,
			discount, total_price, paid_amount, change_amount,
			payment_method_id, description, invoice_date)
			VALUES (?, ?, ?, ?,
					?, ?, ?, ?,
					?, ?, ?)`

	_, err = s.db.Exec(cmd, invoice.Number, cashierID, customerID, invoice.Subtotal,
							invoice.Discount, invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
							paymentMethodID, invoice.Description, invoice.InvoiceDate)

	if err != nil {
		return err
	}

	return nil
}
