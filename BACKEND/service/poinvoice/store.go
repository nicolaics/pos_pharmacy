package poinvoice

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

func (s *Store) GetPurchaseOrderInvoicesByNumber(number int) ([]types.PurchaseOrderInvoice, error) {
	rows, err := s.db.Query("SELECT * FROM purchase_order_invoice WHERE number = ?", number)
	if err != nil {
		return nil, err
	}

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoice, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoice(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderInvoices = append(purchaseOrderInvoices, *purchaseOrderInvoice)
	}

	return purchaseOrderInvoices, nil
}

func (s *Store) GetPurchaseOrderInvoiceByID(id int) (*types.PurchaseOrderInvoice, error) {
	rows, err := s.db.Query("SELECT * FROM purchase_order_invoice WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	purchaseOrderInvoice := new(types.PurchaseOrderInvoice)

	for rows.Next() {
		purchaseOrderInvoice, err = scanRowIntoPurchaseOrderInvoice(rows)

		if err != nil {
			return nil, err
		}
	}

	if purchaseOrderInvoice.ID == 0 {
		return nil, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrderInvoice, nil
}

func (s *Store) GetPurchaseOrderInvoiceByAll(number int, companyId int, supplierId int, cashierId int, totalItems int, invoiceDate time.Time) (*types.PurchaseOrderInvoice, error) {
	query := "SELECT * FROM purchase_order_invoice WHERE number = ? AND company_id ? AND "
	query += "supplier_id = ? AND cashierId = ? AND total_items = ? AND invoice_date ?"

	rows, err := s.db.Query(query, number, companyId, supplierId, cashierId, totalItems, invoiceDate)
	if err != nil {
		return nil, err
	}

	purchaseOrderInvoice := new(types.PurchaseOrderInvoice)

	for rows.Next() {
		purchaseOrderInvoice, err = scanRowIntoPurchaseOrderInvoice(rows)

		if err != nil {
			return nil, err
		}
	}

	if purchaseOrderInvoice.ID == 0 {
		return nil, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrderInvoice, nil
}

func (s *Store) CreatePurchaseOrderInvoice(poInvoice types.PurchaseOrderInvoice) error {
	fields := "number, company_id, supplier_id, cashier_id, total_items, "
	fields += "invoice_date, last_modified"
	values := "?"

	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO purchase_order_invoice (%s) VALUES (%s)", fields, values),
						poInvoice.Number, poInvoice.CompanyID, poInvoice.SupplierID,
						poInvoice.CashierID, poInvoice.TotalItems, poInvoice.InvoiceDate,
						poInvoice.LastModified)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseOrderItems(purchaseMedItem types.PurchaseOrderItem) error {
	fields := "purchase_order_invoice_id, medicine_id, order_qty, received_qty, unit_id, remarks"
	values := "?"

	for i := 0; i < 5; i++ {
		values += ", ?"
	}

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO purchase_order_items (%s) VALUES (%s)", fields, values),
						purchaseMedItem.PurchaseOrderInvoiceID, purchaseMedItem.MedicineID, purchaseMedItem.OrderQty,
						purchaseMedItem.ReceivedQty, purchaseMedItem.UnitID, purchaseMedItem.Remarks)
	if err != nil {
		return err
	}

	return nil
}


func (s *Store) GetPurchaseOrderInvoices(startDate time.Time, endDate time.Time) ([]types.PurchaseOrderInvoice, error) {
	query := fmt.Sprintf("SELECT * FROM purchase_order_invoice WHERE invoice_date BETWEEN DATE('%s') AND DATE('%s') ORDER BY invoice_date DESC",
				startDate, endDate)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoice, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoice(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderInvoices = append(purchaseOrderInvoices, *purchaseOrderInvoice)
	}

	return purchaseOrderInvoices, nil
}


func (s *Store) GetPurchaseOrderItems(purchaseOrderInvoiceId int) ([]types.PurchaseOrderItemsReturn, error) {
	query := "SELECT "

	query += "poi.id, medicine.barcode, medicine.name, poi.order_qty, poi.received_qty, "
	query += "unit.unit, poi.remarks "

	query += "FROM purchase_order_items as poi "
	query += "JOIN purchase_order_invoice ON poi.purchase_order_invoice_id = purchase_order_invoice.id "
	query += "JOIN medicine ON pmi.medicine_id = medicine.id "
	query += "JOIN unit ON pmi.unit_id = unit.id "
	query += "WHERE purchase_order_invoice.id = ? "

	rows, err := s.db.Query(query, purchaseOrderInvoiceId)
	if err != nil {
		return nil, err
	}

	purchaseOrderItems := make([]types.PurchaseOrderItemsReturn, 0)

	for rows.Next() {
		purchaseOrderItem, err := scanRowIntoPurchaseOrderItems(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderItems = append(purchaseOrderItems, *purchaseOrderItem)
	}

	return purchaseOrderItems, nil
}

func (s *Store) DeletePurchaseOrderInvoice(purchaseOrderInvoice *types.PurchaseOrderInvoice) error {
	_, err := s.db.Exec("DELETE FROM purchase_order_invoice WHERE id = ?", purchaseOrderInvoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeletePurchaseOrderItems(purchaseOrderInvoiceId int) error {
	_, err := s.db.Exec("DELETE FROM purchase_order_items WHERE purchase_order_invoice_id = ? ", purchaseOrderInvoiceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseOrderInvoice(id int, purchaseOrderInvoice types.PurchaseOrderInvoice) error {
	fields := "number = ?, company_id = ?, supplier_id = ?, cashier_id = ?, total_items = ?, "
	fields += "invoice_date = ?, last_modified = ?"

	query := fmt.Sprintf("UPDATE purchase_order_invoice SET %s WHERE id = ?", fields)

	_, err := s.db.Exec(query,
						purchaseOrderInvoice.Number, purchaseOrderInvoice.CompanyID, purchaseOrderInvoice.SupplierID,
						purchaseOrderInvoice.CashierID, purchaseOrderInvoice.TotalItems, purchaseOrderInvoice.InvoiceDate,
						purchaseOrderInvoice.LastModified, id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPurchaseOrderInvoice(rows *sql.Rows) (*types.PurchaseOrderInvoice, error) {
	purchaseOrderInvoice := new(types.PurchaseOrderInvoice)

	err := rows.Scan(
		&purchaseOrderInvoice.ID,
		&purchaseOrderInvoice.Number,
		&purchaseOrderInvoice.CompanyID,
		&purchaseOrderInvoice.SupplierID,
		&purchaseOrderInvoice.CashierID,
		&purchaseOrderInvoice.TotalItems,
		&purchaseOrderInvoice.InvoiceDate,
		&purchaseOrderInvoice.LastModified,
		&purchaseOrderInvoice.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	purchaseOrderInvoice.InvoiceDate = purchaseOrderInvoice.InvoiceDate.Local()
	purchaseOrderInvoice.CreatedAt = purchaseOrderInvoice.CreatedAt.Local()
	purchaseOrderInvoice.LastModified = purchaseOrderInvoice.LastModified.Local()

	return purchaseOrderInvoice, nil
}

func scanRowIntoPurchaseOrderItems(rows *sql.Rows) (*types.PurchaseOrderItemsReturn, error) {
	purchaseOrderItem := new(types.PurchaseOrderItemsReturn)

	err := rows.Scan(
		&purchaseOrderItem.ID,
		&purchaseOrderItem.MedicineBarcode,
		&purchaseOrderItem.MedicineName,
		&purchaseOrderItem.OrderQty,
		&purchaseOrderItem.ReceivedQty,
		&purchaseOrderItem.Unit,
		&purchaseOrderItem.Remarks,
	)

	if err != nil {
		return nil, err
	}

	return purchaseOrderItem, nil
}