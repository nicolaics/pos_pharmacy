package invoice

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetInvoiceByID(id int) (*types.Invoice, error) {
	rows, err := s.db.Query("SELECT * FROM invoice WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	invoice := new(types.Invoice)

	for rows.Next() {
		invoice, err = scanRowIntoInvoice(rows)

		if err != nil {
			return nil, err
		}
	}

	if invoice.ID == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	return invoice, nil
}

func (s *Store) GetInvoiceByAll(number int, cashierId int, customerId int, totalPrice float64, invoiceDate time.Time) (*types.Invoice, error) {
	query := "SELECT * FROM invoice WHERE number = ? AND cashier_id = ? AND customer_id = ? AND "
	query += "total_price = ? AND invoice_date ?"

	rows, err := s.db.Query(query, number, cashierId, customerId, totalPrice, invoiceDate)
	if err != nil {
		return nil, err
	}

	invoice := new(types.Invoice)

	for rows.Next() {
		invoice, err = scanRowIntoInvoice(rows)

		if err != nil {
			return nil, err
		}
	}

	if invoice.ID == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	return invoice, nil
}

func (s *Store) GetInvoicesByNumber(number int) ([]types.Invoice, error) {
	rows, err := s.db.Query("SELECT * FROM invoice WHERE number = ? ", number)
	if err != nil {
		return nil, err
	}

	invoices := make([]types.Invoice, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoice(rows)

		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]types.Invoice, error) {
	query := fmt.Sprintf("SELECT * FROM invoice WHERE invoice_date BETWEEN DATE('%s') AND DATE('%s') ORDER BY invoice_date DESC",
				startDate, endDate)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	invoices := make([]types.Invoice, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoice(rows)

		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) CreateInvoice(invoice types.Invoice) error {
	fields := "number, cashier_id, customer_id, subtotal, discount, tax, "
	fields += "total_price, paid_amount, change_amount, payment_method_id, description, "
	fields += "invoice_date"
	values := "?"

	for i := 0; i < 11; i++ {
		values += ", ?"
	}

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO invoice (%s) VALUES (%s)", fields, values),
						invoice.Number, invoice.CashierID, invoice.CustomerID,
						invoice.Subtotal, invoice.Discount, invoice.Tax,
						invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
						invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateMedicineItems(medicineItem types.MedicineItems) error {
	fields := "invoice_id, medicine_id, qty, unit_id, "
	fields += "price, discount, subtotal"
	values := "?"

	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO medicine_items (%s) VALUES (%s)", fields, values),
						medicineItem.InvoiceID, medicineItem.MedicineID, medicineItem.Qty,
						medicineItem.UnitID, medicineItem.Price, medicineItem.Discount,
						medicineItem.Subtotal)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetMedicineItems(invoiceId int) ([]types.MedicineItemReturnPayload, error) {
	query := "SELECT "

	query += "mi.id, medicine.barcode, medicine.name, mi.qty, unit.unit, mi.price, mi.discount, "
	query += "mi.subtotal "

	query += "FROM medicine_items as mi "
	query += "JOIN invoice ON mi.invoice_id = invoice.id "
	query += "JOIN medicine ON mi.medicine_id = medicine.id "
	query += "JOIN unit ON mi.unit_id = unit.id "
	query += "WHERE invoice.id = ? "

	rows, err := s.db.Query(query, invoiceId)
	if err != nil {
		return nil, err
	}

	medicineItems := make([]types.MedicineItemReturnPayload, 0)

	for rows.Next() {
		medicineItem, err := scanRowIntoMedicineItems(rows)

		if err != nil {
			return nil, err
		}

		medicineItems = append(medicineItems, *medicineItem)
	}

	return medicineItems, nil
}

func (s *Store) DeleteInvoice(invoice *types.Invoice) error {
	_, err := s.db.Exec("DELETE FROM invoice WHERE id = ?", invoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteMedicineItems(invoiceId int) error {
	_, err := s.db.Exec("DELETE FROM medicine_items WHERE invoice_id = ? ", invoiceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyInvoice(id int, invoice types.Invoice) error {
	fields := "number = ?, cashier_id = ?, customer_id = ?, subtotal = ?, discount = ?, "
	fields += "tax = ?, total_price = ?, paid_amount = ?, change_amount = ?, "
	fields += "payment_method_id = ?, description = ?, invoice_date = ?"

	query := fmt.Sprintf("UPDATE invoice SET %s WHERE id = ?", fields)

	_, err := s.db.Exec(query,
						invoice.Number, invoice.CashierID, invoice.CustomerID,
						invoice.Subtotal, invoice.Discount, invoice.Tax,
						invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
						invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate,
						id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoInvoice(rows *sql.Rows) (*types.Invoice, error) {
	invoice := new(types.Invoice)

	err := rows.Scan(
		&invoice.ID,
		&invoice.Number,
		&invoice.CashierID,
		&invoice.CustomerID,
		&invoice.Subtotal,
		&invoice.Discount,
		&invoice.Tax,
		&invoice.TotalPrice,
		&invoice.PaidAmount,
		&invoice.ChangeAmount,
		&invoice.PaymentMethodID,
		&invoice.Description,
		&invoice.InvoiceDate,
		&invoice.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	invoice.InvoiceDate = invoice.InvoiceDate.Local()
	invoice.CreatedAt = invoice.CreatedAt.Local()

	return invoice, nil
}

func scanRowIntoMedicineItems(rows *sql.Rows) (*types.MedicineItemReturnPayload, error) {
	medicineItem := new(types.MedicineItemReturnPayload)

	err := rows.Scan(
		&medicineItem.ID,
		&medicineItem.MedicineBarcode,
		&medicineItem.MedicineName,
		&medicineItem.Qty,
		&medicineItem.Unit,
		&medicineItem.Price,
		&medicineItem.Discount,
		&medicineItem.Subtotal,
	)

	if err != nil {
		return nil, err
	}

	return medicineItem, nil
}
