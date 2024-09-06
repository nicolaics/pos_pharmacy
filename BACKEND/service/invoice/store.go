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
	query := "SELECT * FROM invoice WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (s *Store) GetInvoiceID(number int, userId int, customerId int, totalPrice float64, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM invoice 
				WHERE number = ? AND user_id = ? AND customer_id = ? AND 
				total_price = ? AND invoice_date ? AND deleted_at IS NULL`

	rows, err := s.db.Query(query, number, userId, customerId, totalPrice, invoiceDate)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var invoiceId int

	for rows.Next() {
		err = rows.Scan(&invoiceId)
		if err != nil {
			return -1, err
		}
	}

	if invoiceId == 0 {
		return -1, fmt.Errorf("invoice not found")
	}

	return invoiceId, nil
}

func (s *Store) GetInvoicesByNumber(number int) ([]types.Invoice, error) {
	query := "SELECT * FROM invoice WHERE number = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	query := `SELECT * FROM invoice 
				WHERE (invoice_date BETWEEN DATE(?) AND DATE(?)) 
					AND deleted_at IS NULL 
				ORDER BY invoice_date DESC`
	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	values := "?"
	for i := 0; i < 12; i++ {
		values += ", ?"
	}

	query := `INSERT INTO invoice (
			number, user_id, customer_id, subtotal, discount, tax, 
			total_price, paid_amount, change_amount, payment_method_id, description, 
			invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		invoice.Number, invoice.UserID, invoice.CustomerID,
		invoice.Subtotal, invoice.Discount, invoice.Tax,
		invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
		invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate,
		invoice.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateMedicineItems(medicineItem types.MedicineItems) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO medicine_items (
		invoice_id, medicine_id, qty, unit_id, price, discount, subtotal
	) VALUES (` + values + `)`
	_, err := s.db.Exec(query,
		medicineItem.InvoiceID, medicineItem.MedicineID, medicineItem.Qty,
		medicineItem.UnitID, medicineItem.Price, medicineItem.Discount,
		medicineItem.Subtotal)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetMedicineItems(invoiceId int) ([]types.MedicineItemReturnPayload, error) {
	query := `SELECT 
				mi.id, 
				medicine.barcode, medicine.name, 
				mi.qty, 
				unit.name, 
				mi.price, mi.discount, mi.subtotal 
				FROM medicine_items as mi 
				JOIN invoice ON mi.invoice_id = invoice.id 
				JOIN medicine ON mi.medicine_id = medicine.id 
				JOIN unit ON mi.unit_id = unit.id 
				WHERE invoice.id = ? AND invoice.deleted_at IS NULL`

	rows, err := s.db.Query(query, invoiceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (s *Store) DeleteInvoice(invoice *types.Invoice, userId int) error {
	query := "UPDATE invoice SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, invoice.ID)
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
	query := `UPDATE invoice SET 
			number = ?, user_id = ?, customer_id = ?, subtotal = ?, discount = ?, 
			tax = ?, total_price = ?, paid_amount = ?, change_amount = ?, 
			payment_method_id = ?, description = ?, invoice_date = ?, last_modified = ?,
			last_modified_by_user_id = ? 
			WHERE id = ? AND deleted_at IS NULL`

	_, err := s.db.Exec(query,
		invoice.Number, invoice.UserID, invoice.CustomerID,
		invoice.Subtotal, invoice.Discount, invoice.Tax,
		invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
		invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate,
		time.Now(), invoice.LastModifiedByUserID, id)
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
		&invoice.UserID,
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
		&invoice.LastModified,
		&invoice.LastModifiedByUserID,
		&invoice.DeletedAt,
		&invoice.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	invoice.InvoiceDate = invoice.InvoiceDate.Local()
	invoice.CreatedAt = invoice.CreatedAt.Local()
	invoice.LastModified = invoice.LastModified.Local()

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
