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
	query := "SELECT * FROM purchase_order_invoice WHERE number = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, number)
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
	query := "SELECT * FROM purchase_order_invoice WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
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

func (s *Store) GetPurchaseOrderInvoiceByAll(number int, companyId int, supplierId int, userId int, totalItems int, invoiceDate time.Time) (*types.PurchaseOrderInvoice, error) {
	query := `SELECT * FROM purchase_order_invoice WHERE number = ? AND company_id ? AND 
	supplier_id = ? AND userId = ? AND total_items = ? AND invoice_date ? AND deleted_at IS NULL`

	rows, err := s.db.Query(query, number, companyId, supplierId, userId, totalItems, invoiceDate)
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

func (s *Store) CreatePurchaseOrderInvoice(poInvoice types.PurchaseOrderInvoice, userId int) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order_invoice (
		number, company_id, supplier_id, user_id, total_items, 
		invoice_date, modified_by_user_id
	) VALUES ({values})`

	_, err := s.db.Exec(query,
		poInvoice.Number, poInvoice.CompanyID, poInvoice.SupplierID,
		poInvoice.UserID, poInvoice.TotalItems, poInvoice.InvoiceDate,
		userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseOrderItems(purchaseMedItem types.PurchaseOrderItem, userId int) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order_items (
		purchase_order_invoice_id, medicine_id, order_qty, received_qty, unit_id, 
		remarks, modified_by_user_id
	) VALUES ({values})`

	_, err := s.db.Exec(query,
		purchaseMedItem.PurchaseOrderInvoiceID, purchaseMedItem.MedicineID, purchaseMedItem.OrderQty,
		purchaseMedItem.ReceivedQty, purchaseMedItem.UnitID, purchaseMedItem.Remarks, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPurchaseOrderInvoices(startDate time.Time, endDate time.Time) ([]types.PurchaseOrderInvoice, error) {
	query := `SELECT * FROM purchase_order_invoice 
	WHERE (invoice_date BETWEEN DATE(?) AND DATE(?)) 
	AND deleted_at IS NULL ORDER BY invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
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
	query := `SELECT 
				poit.id, 
				medicine.barcode, medicine.name, 
				poit.order_qty, poit.received_qty, 
				unit.unit, 
				poit.remarks, poit.last_modified, user.name 
				FROM purchase_order_items as poit 
				JOIN purchase_order_invoice as poin 
					ON poit.purchase_order_invoice_id = purchase_order_invoice.id 
				JOIN medicine ON poit.medicine_id = medicine.id 
				JOIN unit ON poit.unit_id = unit.id 
				JOIN user ON user.id = poin.modified_by_user_id 
				WHERE poin.id = ? AND poin.deleted_at IS NULL`

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

func (s *Store) DeletePurchaseOrderInvoice(purchaseOrderInvoice *types.PurchaseOrderInvoice, userId int) error {
	query := "UPDATE purchase_order_invoice SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, purchaseOrderInvoice.ID)
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

func (s *Store) ModifyPurchaseOrderInvoice(poiid int, purchaseOrderInvoice types.PurchaseOrderInvoice, userId int) error {
	query := `UPDATE purchase_order_invoice 
				SET number = ?, company_id = ?, supplier_id = ?, user_id = ?, total_items = ?, 
				invoice_date = ?, last_modified = ?, modified_by_user_id = ? 
				WHERE id = ?`

	_, err := s.db.Exec(query,
		purchaseOrderInvoice.Number, purchaseOrderInvoice.CompanyID, purchaseOrderInvoice.SupplierID,
		purchaseOrderInvoice.UserID, purchaseOrderInvoice.TotalItems, purchaseOrderInvoice.InvoiceDate,
		purchaseOrderInvoice.LastModified, userId, poiid)
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
		&purchaseOrderInvoice.UserID,
		&purchaseOrderInvoice.TotalItems,
		&purchaseOrderInvoice.InvoiceDate,
		&purchaseOrderInvoice.CreatedAt,
		&purchaseOrderInvoice.LastModified,
		&purchaseOrderInvoice.ModifiedByUserID,
		&purchaseOrderInvoice.DeletedAt,
		&purchaseOrderInvoice.DeletedByUserID,		
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
		&purchaseOrderItem.LastModified,
		&purchaseOrderItem.ModifiedByUserName,
	)

	if err != nil {
		return nil, err
	}

	return purchaseOrderItem, nil
}
