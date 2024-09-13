package poinvoice

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/nicolaics/pos_pharmacy/logger"
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
	defer rows.Close()

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
	defer rows.Close()

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

func (s *Store) GetPurchaseOrderInvoiceID(number int, companyId int, supplierId int, userId int, totalItems int, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM purchase_order_invoice 
				WHERE number = ? AND company_id ? 
				AND supplier_id = ? AND userId = ? AND total_items = ? 
				AND invoice_date ? AND deleted_at IS NULL`

	rows, err := s.db.Query(query, number, companyId, supplierId, userId, totalItems, invoiceDate)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var purchaseOrderInvoiceId int

	for rows.Next() {
		err = rows.Scan(&purchaseOrderInvoiceId)
		if err != nil {
			return -1, err
		}
	}

	if purchaseOrderInvoiceId == 0 {
		return -1, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrderInvoiceId, nil
}

func (s *Store) GetNumberOfPurchaseOrderInvoices() (int, error) {
	query := `SELECT COUNT(*) FROM purchase_order_invoice`
	row := s.db.QueryRow(query)
	if row.Err() != nil {
		return -1, row.Err()
	}

	var numberOfPurchaseOrderInvoices int

	err := row.Scan(&numberOfPurchaseOrderInvoices)
	if err != nil {
		return -1, err
	}

	return numberOfPurchaseOrderInvoices, nil
}

func (s *Store) CreatePurchaseOrderInvoice(poInvoice types.PurchaseOrderInvoice) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order_invoice (
		number, company_id, supplier_id, user_id, total_items, 
		invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		poInvoice.Number, poInvoice.CompanyID, poInvoice.SupplierID,
		poInvoice.UserID, poInvoice.TotalItems, poInvoice.InvoiceDate,
		poInvoice.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseOrderItems(purchaseMedItem types.PurchaseOrderItem) error {
	values := "?"
	for i := 0; i < 5; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order_items (
		purchase_order_invoice_id, medicine_id, order_qty, received_qty, unit_id, remarks
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseMedItem.PurchaseOrderInvoiceID, purchaseMedItem.MedicineID, purchaseMedItem.OrderQty,
		purchaseMedItem.ReceivedQty, purchaseMedItem.UnitID, purchaseMedItem.Remarks)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPurchaseOrderInvoicesByDate(startDate time.Time, endDate time.Time) ([]types.PurchaseOrderInvoiceListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_items, poi.invoice_date 
					FROM purchase_order_invoice AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE (poi.invoice_date BETWEEN DATE(?) AND DATE(?)) 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderInvoices = append(purchaseOrderInvoices, *purchaseOrderInvoice)
	}

	return purchaseOrderInvoices, nil
}

func (s *Store) GetPurchaseOrderInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]types.PurchaseOrderInvoiceListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_items, poi.invoice_date 
					FROM purchase_order_invoice AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE (poi.invoice_date BETWEEN DATE(?) AND DATE(?)) 
					AND poi.number LIKE ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	searchVal := "%"
	for _, val := range(strconv.Itoa(number)) {
		if string(val) != " " {
			searchVal += (string(val) + "%")
		}
	}

	rows, err := s.db.Query(query, startDate, endDate, searchVal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderInvoices = append(purchaseOrderInvoices, *purchaseOrderInvoice)
	}

	return purchaseOrderInvoices, nil
}

func (s *Store) GetPurchaseOrderInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.PurchaseOrderInvoiceListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_items, poi.invoice_date 
					FROM purchase_order_invoice AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE (poi.invoice_date BETWEEN DATE(?) AND DATE(?)) 
					AND poi.user_id = ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderInvoices = append(purchaseOrderInvoices, *purchaseOrderInvoice)
	}

	return purchaseOrderInvoices, nil
}

func (s *Store) GetPurchaseOrderInvoicesByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]types.PurchaseOrderInvoiceListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_items, poi.invoice_date 
					FROM purchase_order_invoice AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE (poi.invoice_date BETWEEN DATE(?) AND DATE(?)) 
					AND poi.supplier_id = ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrderInvoices := make([]types.PurchaseOrderInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrderInvoice, err := scanRowIntoPurchaseOrderInvoiceLists(rows)

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
				unit.name, 
				poit.remarks  
				FROM purchase_order_items as poit 
				JOIN purchase_order_invoice as poin 
					ON poit.purchase_order_invoice_id = purchase_order_invoice.id 
				JOIN medicine ON poit.medicine_id = medicine.id 
				JOIN unit ON poit.unit_id = unit.id 
				WHERE poin.id = ? AND poin.deleted_at IS NULL`

	rows, err := s.db.Query(query, purchaseOrderInvoiceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	data, err := s.GetPurchaseOrderInvoiceByID(purchaseOrderInvoice.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "purchase-order-invoice", userId, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeletePurchaseOrderItems(purchaseOrderInvoice *types.PurchaseOrderInvoice, userId int) error {
	data, err := s.GetPurchaseOrderItems(purchaseOrderInvoice.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"purchase_order_invoice": purchaseOrderInvoice,
		"deleted_medicine_items": data,
	}

	err = logger.WriteLog("delete", "purchase-order-invoice", userId, purchaseOrderInvoice.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}
	
	
	_, err = s.db.Exec("DELETE FROM purchase_order_items WHERE purchase_order_invoice_id = ? ", purchaseOrderInvoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseOrderInvoice(poiid int, purchaseOrderInvoice types.PurchaseOrderInvoice, userId int) error {
	data, err := s.GetPurchaseOrderInvoiceByID(poiid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_dataa": data,
	}

	err = logger.WriteLog("modify", "purchase-order-invoice", userId, poiid, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE purchase_order_invoice 
				SET number = ?, company_id = ?, supplier_id = ?, total_items = ?, 
				invoice_date = ?, last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query,
		purchaseOrderInvoice.Number, purchaseOrderInvoice.CompanyID, purchaseOrderInvoice.SupplierID,
		purchaseOrderInvoice.TotalItems, purchaseOrderInvoice.InvoiceDate,
		time.Now(), purchaseOrderInvoice.LastModifiedByUserID, poiid)
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
		&purchaseOrderInvoice.LastModifiedByUserID,
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

func scanRowIntoPurchaseOrderInvoiceLists(rows *sql.Rows) (*types.PurchaseOrderInvoiceListsReturnPayload, error) {
	purchaseOrderInvoice := new(types.PurchaseOrderInvoiceListsReturnPayload)

	err := rows.Scan(
		&purchaseOrderInvoice.ID,
		&purchaseOrderInvoice.Number,
		&purchaseOrderInvoice.SupplierName,
		&purchaseOrderInvoice.UserName,
		&purchaseOrderInvoice.TotalItems,
		&purchaseOrderInvoice.InvoiceDate,
	)

	if err != nil {
		return nil, err
	}

	purchaseOrderInvoice.InvoiceDate = purchaseOrderInvoice.InvoiceDate.Local()

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
