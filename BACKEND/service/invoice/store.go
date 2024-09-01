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

/*
func (s *Store) GetPurchaseMedicineItems(purchaseInvoiceId int) ([]types.PurchaseMedicineItemsReturn, error) {
	query := "SELECT "

	query += "pmi.id, medicine.barcode, medicine.name, pmi.qty, unit.unit, pmi.purchase_price, pmi.purchase_discount, "
	query += "pmi.purchase_tax, pmi.subtotal, pmi.batch_number, pmi.expired_date "

	query += "FROM purchase_medicine_items as pmi "
	query += "JOIN purchase_invoice as pi ON pmi.purchase_invoice_id = pi.id "
	query += "JOIN medicine ON pmi.medicine_id = medicine.id "
	query += "JOIN unit ON pmi.unit_id = unit.id "
	query += "WHERE pi.id = ? "

	rows, err := s.db.Query(query, purchaseInvoiceId)
	if err != nil {
		return nil, err
	}

	purchaseMedicineItems := make([]types.PurchaseMedicineItemsReturn, 0)

	for rows.Next() {
		purchaseMedicineItem, err := scanRowIntoPurchaseMedicineItems(rows)

		if err != nil {
			return nil, err
		}

		purchaseMedicineItems = append(purchaseMedicineItems, *purchaseMedicineItem)
	}

	return purchaseMedicineItems, nil
}

func (s *Store) DeletePurchaseInvoice(purchaseInvoice *types.PurchaseInvoice) error {
	_, err := s.db.Exec("DELETE FROM purchase_invoice WHERE id = ?", purchaseInvoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeletePurchaseMedicineItems(purchaseInvoiceId int) error {
	_, err := s.db.Exec("DELETE FROM purchase_medicine_items WHERE purchase_invoice_id = ? ", purchaseInvoiceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseInvoice(id int, purchaseInvoice types.PurchaseInvoice) error {
	fields := "number = ?, company_id = ?, supplier_id = ?, subtotal = ?, discount = ?, "
	fields += "tax = ?, total_price = ?, description = ?, cashier_id = ?, invoice_date = ?"

	query := fmt.Sprintf("UPDATE purchase_invoice SET %s WHERE id = ?", fields)

	_, err := s.db.Exec(query,
						purchaseInvoice.Number, purchaseInvoice.CompanyID, purchaseInvoice.SupplierID,
						purchaseInvoice.Subtotal, purchaseInvoice.Discount, purchaseInvoice.Tax,
						purchaseInvoice.TotalPrice, purchaseInvoice.Description, purchaseInvoice.CashierID,
						purchaseInvoice.InvoiceDate, id)
	if err != nil {
		return err
	}

	return nil
}
*/

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

/*
func scanRowIntoPurchaseMedicineItems(rows *sql.Rows) (*types.PurchaseMedicineItemsReturn, error) {
	purchaseMedicineItem := new(types.PurchaseMedicineItemsReturn)

	err := rows.Scan(
		&purchaseMedicineItem.ID,
		&purchaseMedicineItem.MedicineBarcode,
		&purchaseMedicineItem.MedicineName,
		&purchaseMedicineItem.Qty,
		&purchaseMedicineItem.Unit,
		&purchaseMedicineItem.Price,
		&purchaseMedicineItem.Discount,
		&purchaseMedicineItem.Tax,
		&purchaseMedicineItem.Subtotal,
		&purchaseMedicineItem.BatchNumber,
		&purchaseMedicineItem.ExpDate,
	)

	if err != nil {
		return nil, err
	}

	purchaseMedicineItem.ExpDate = purchaseMedicineItem.ExpDate.Local()

	return purchaseMedicineItem, nil
}
*/