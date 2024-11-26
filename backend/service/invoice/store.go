package invoice

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetInvoiceByID(id int) (*types.Invoice, error) {
	query := "SELECT * FROM invoice WHERE id = ? AND deleted_at IS NULL ORDER BY invoice_date DESC"
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

func (s *Store) GetInvoiceID(number int, customerId int, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM invoice 
				WHERE number = ? AND customer_id = ? 
				AND invoice_date = ? AND deleted_at IS NULL 
				ORDER BY invoice_date DESC`

	rows, err := s.db.Query(query, number, customerId, invoiceDate)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var invoiceId int

	for rows.Next() {
		err = rows.Scan(&invoiceId)
		if err != nil {
			return 0, err
		}
	}

	if invoiceId == 0 {
		return 0, fmt.Errorf("invoice not found")
	}

	return invoiceId, nil
}

func (s *Store) GetInvoicesByNumber(number int) ([]types.Invoice, error) {
	query := "SELECT * FROM invoice WHERE number LIKE ? AND deleted_at IS NULL ORDER BY invoice_date DESC"

	searchVal := "%"
	for _, val := range strconv.Itoa(number) {
		if string(val) != " " {
			searchVal += (string(val) + "%")
		}
	}

	rows, err := s.db.Query(query, searchVal)
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

func (s *Store) GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]types.InvoiceListsReturnPayload, error) {
	query := `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND invoice.deleted_at IS NULL 
				ORDER BY invoice.invoice_date DESC`
	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]types.InvoiceListsReturnPayload, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoiceLists(rows)
		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]types.InvoiceListsReturnPayload, error) {
	query := `SELECT COUNT(*) 
				FROM invoice 
				WHERE invoice_date >= ? AND invoice_date < ? 
				AND number = ? 
				AND deleted_at IS NULL`

	row := s.db.QueryRow(query, startDate, endDate, number)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	invoices := make([]types.InvoiceListsReturnPayload, 0)

	if count == 0 {
		query = `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND invoice.number LIKE ? 
					AND invoice.deleted_at IS NULL 
					ORDER BY invoice.invoice_date DESC`

		searchVal := "%"
		for _, val := range strconv.Itoa(number) {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, startDate, endDate, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			invoice, err := scanRowIntoInvoiceLists(rows)
			if err != nil {
				return nil, err
			}

			invoices = append(invoices, *invoice)
		}

		return invoices, nil
	}

	query = `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND invoice.number = ? 
					AND invoice.deleted_at IS NULL 
					ORDER BY invoice.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		invoice, err := scanRowIntoInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.InvoiceListsReturnPayload, error) {
	query := `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND invoice.deleted_at IS NULL 
				ORDER BY invoice.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]types.InvoiceListsReturnPayload, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoiceLists(rows)
		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetInvoicesByDateAndCustomerID(startDate time.Time, endDate time.Time, cid int) ([]types.InvoiceListsReturnPayload, error) {
	query := `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND customer_id = ? 
					AND invoice.deleted_at IS NULL 
					ORDER BY invoice.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]types.InvoiceListsReturnPayload, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoiceLists(rows)
		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetInvoicesByDateAndPaymentMethodID(startDate time.Time, endDate time.Time, pmid int) ([]types.InvoiceListsReturnPayload, error) {
	query := `SELECT invoice.id, invoice.number, 
					user.name, customer.name, 
					invoice.subtotal, 
					invoice.discount_percentage, invoice.discount_amount, 
					invoice.tax_percentage, invoice.tax_amount, 
					invoice.total_price, 
					payment_method.name, 
					invoice.description, invoice.invoice_date 
					FROM invoice 
					JOIN user ON user.id = invoice.user_id 
					JOIN customer ON customer.id = invoice.customer_id 
					JOIN payment_method ON payment_method.id = invoice.payment_method_id 
					WHERE invoice.invoice_date >= ? AND invoice.invoice_date < ? 
					AND invoice.deleted_at IS NULL 
				ORDER BY invoice.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, pmid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]types.InvoiceListsReturnPayload, 0)

	for rows.Next() {
		invoice, err := scanRowIntoInvoiceLists(rows)
		if err != nil {
			return nil, err
		}

		invoices = append(invoices, *invoice)
	}

	return invoices, nil
}

func (s *Store) GetNumberOfInvoices(startDate time.Time, endDate time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM invoice WHERE invoice_date >= ? AND invoice_date < ?`
	row := s.db.QueryRow(query, startDate, endDate)
	if row.Err() != nil {
		return -1, row.Err()
	}

	var numberOfInvoices int

	err := row.Scan(&numberOfInvoices)
	if err != nil {
		return -1, err
	}

	return numberOfInvoices, nil
}

func (s *Store) CreateInvoice(invoice types.Invoice) error {
	values := "?"
	for i := 0; i < 14; i++ {
		values += ", ?"
	}

	query := `INSERT INTO invoice (
			number, user_id, customer_id, subtotal, discount_percentage, discount_amount, 
			tax_percentage, tax_amount, total_price, paid_amount, change_amount, 
			payment_method_id, description, invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		invoice.Number, invoice.UserID, invoice.CustomerID,
		invoice.Subtotal, invoice.DiscountPercentage, invoice.DiscountAmount,
		invoice.TaxPercentage, invoice.TaxAmount, invoice.TotalPrice,
		invoice.PaidAmount, invoice.ChangeAmount, invoice.PaymentMethodID,
		invoice.Description, invoice.InvoiceDate, invoice.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateMedicineItem(medicineItem types.InvoiceMedicineItem) error {
	values := "?"
	for i := 0; i < 7; i++ {
		values += ", ?"
	}

	query := `INSERT INTO medicine_item (
		invoice_id, medicine_id, qty, unit_id, price, 
		discount_percentage, discount_amount, subtotal
	) VALUES (` + values + `)`
	_, err := s.db.Exec(query,
		medicineItem.InvoiceID, medicineItem.MedicineID, medicineItem.Qty,
		medicineItem.UnitID, medicineItem.Price, medicineItem.DiscountPercentage,
		medicineItem.DiscountAmount, medicineItem.Subtotal)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetMedicineItem(invoiceId int) ([]types.InvoiceMedicineItemReturnPayload, error) {
	query := `SELECT 
				mi.id, 
				medicine.barcode, medicine.name, 
				mi.qty, 
				unit.name, 
				mi.price, mi.discount_percentage, mi_discount_amount, mi.subtotal 
				FROM medicine_item as mi 
				JOIN invoice ON mi.invoice_id = invoice.id 
				JOIN medicine ON mi.medicine_id = medicine.id 
				JOIN unit ON mi.unit_id = unit.id 
				WHERE invoice.id = ? AND invoice.deleted_at IS NULL`

	rows, err := s.db.Query(query, invoiceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicineItems := make([]types.InvoiceMedicineItemReturnPayload, 0)

	for rows.Next() {
		medicineItem, err := scanRowIntoMedicineItem(rows)

		if err != nil {
			return nil, err
		}

		medicineItems = append(medicineItems, *medicineItem)
	}

	return medicineItems, nil
}

func (s *Store) DeleteInvoice(invoice *types.Invoice, user *types.User) error {
	query := "UPDATE invoice SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, invoice.ID)
	if err != nil {
		return err
	}

	data, err := s.GetInvoiceByID(invoice.ID)
	if err != nil {
		return err
	}

	err = logger.WriteServerLog("delete", "invoice", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeleteMedicineItem(invoice *types.Invoice, user *types.User) error {
	data, err := s.GetMedicineItem(invoice.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"invoice":               invoice,
		"deleted_medicine_item": data,
	}

	err = logger.WriteServerLog("delete", "invoice", user.Name, invoice.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM medicine_item WHERE invoice_id = ? ", invoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyInvoice(invoiceId int, invoice types.Invoice, user *types.User) error {
	data, err := s.GetInvoiceByID(invoiceId)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteServerLog("modify", "invoice", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE invoice SET 
			number = ?, user_id = ?, customer_id = ?, subtotal = ?, 
			discount_percentage = ?, discount_amount = ?, 
			tax_percentage = ?, tax_amount = ?, 
			total_price = ?, paid_amount = ?, change_amount = ?, 
			payment_method_id = ?, description = ?, invoice_date = ?, last_modified = ?,
			last_modified_by_user_id = ? 
			WHERE id = ? AND deleted_at IS NULL`

	_, err = s.db.Exec(query,
		invoice.Number, invoice.UserID, invoice.CustomerID,
		invoice.Subtotal, invoice.DiscountPercentage, invoice.DiscountAmount,
		invoice.TaxPercentage, invoice.TaxAmount,
		invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
		invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate,
		time.Now(), invoice.LastModifiedByUserID, invoiceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePdfUrl(invoiceId int, pdfUrl string) error {
	query := `UPDATE invoice SET pdf_url = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := s.db.Exec(query, pdfUrl, invoiceId)
	if err != nil {
		return err
	}

	return nil
}

// false means doesn't exist
func (s *Store) IsPdfUrlExist(fileName, columnName string) (bool, error) {
	var query string

	if columnName == "invoice" {
		query = `SELECT COUNT(*) FROM invoice WHERE pdf_url = ?`
	} else if columnName == "receipt" {
		query = `SELECT COUNT(*) FROM invoice WHERE receipt_pdf_url = ?`
	}

	row := s.db.QueryRow(query, fileName)
	if row.Err() != nil && row.Err() != sql.ErrNoRows {
		return true, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return (count > 0), nil
}

func (s *Store) UpdateReceiptPdfUrl(invoiceId int, receiptPdfUrl string) error {
	query := `UPDATE invoice SET receipt_pdf_url = ? WHERE id = ?`
	_, err := s.db.Exec(query, receiptPdfUrl, invoiceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AbsoluteDeleteInvoice(invoice types.Invoice) error {
	query := `SELECT id FROM invoice WHERE number = ? AND user_id = ? 
				AND customer_id = ? AND subtotal = ? 
				AND discount_percentage = ? AND discount_amount = ? 
				AND tax_percentage = ? AND tax_amount = ? 
				AND total_price = ? AND paid_amount = ? 
				AND change_amount = ? AND payment_method_id = ? 
				AND description = ? AND invoice_date = ?`

	rows, err := s.db.Query(query, invoice.Number, invoice.UserID,
		invoice.CustomerID, invoice.Subtotal,
		invoice.DiscountPercentage, invoice.DiscountAmount,
		invoice.TaxPercentage, invoice.TaxAmount,
		invoice.TotalPrice, invoice.PaidAmount, invoice.ChangeAmount,
		invoice.PaymentMethodID, invoice.Description, invoice.InvoiceDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	var id int

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil
		}
	}

	if id == 0 {
		return nil
	}

	query = "DELETE FROM medicine_item WHERE invoice_id = ?"
	_, _ = s.db.Exec(query, id)

	query = `DELETE FROM invoice WHERE id = ?`
	_, _ = s.db.Exec(query, id)

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
		&invoice.DiscountPercentage,
		&invoice.DiscountAmount,
		&invoice.TaxPercentage,
		&invoice.TaxAmount,
		&invoice.TotalPrice,
		&invoice.PaidAmount,
		&invoice.ChangeAmount,
		&invoice.PaymentMethodID,
		&invoice.Description,
		&invoice.InvoiceDate,
		&invoice.CreatedAt,
		&invoice.LastModified,
		&invoice.LastModifiedByUserID,
		&invoice.PdfUrl,
		&invoice.ReceiptPdfUrl,
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

func scanRowIntoInvoiceLists(rows *sql.Rows) (*types.InvoiceListsReturnPayload, error) {
	invoice := new(types.InvoiceListsReturnPayload)

	err := rows.Scan(
		&invoice.ID,
		&invoice.Number,
		&invoice.UserName,
		&invoice.CustomerName,
		&invoice.Subtotal,
		&invoice.DiscountPercentage,
		&invoice.DiscountAmount,
		&invoice.TaxPercentage,
		&invoice.TaxAmount,
		&invoice.TotalPrice,
		&invoice.PaymentMethodName,
		&invoice.Description,
		&invoice.InvoiceDate,
	)

	if err != nil {
		return nil, err
	}

	invoice.InvoiceDate = invoice.InvoiceDate.Local()

	return invoice, nil
}

func scanRowIntoMedicineItem(rows *sql.Rows) (*types.InvoiceMedicineItemReturnPayload, error) {
	medicineItem := new(types.InvoiceMedicineItemReturnPayload)

	err := rows.Scan(
		&medicineItem.ID,
		&medicineItem.MedicineBarcode,
		&medicineItem.MedicineName,
		&medicineItem.Qty,
		&medicineItem.Unit,
		&medicineItem.Price,
		&medicineItem.DiscountPercentage,
		&medicineItem.DiscountAmount,
		&medicineItem.Subtotal,
	)

	if err != nil {
		return nil, err
	}

	return medicineItem, nil
}
