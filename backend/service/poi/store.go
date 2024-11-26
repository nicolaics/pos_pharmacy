package poi

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

func (s *Store) GetPurchaseOrderByNumber(number int) (*types.PurchaseOrder, error) {
	query := "SELECT * FROM purchase_order WHERE number = ? AND deleted_at IS NULL ORDER BY invoice_date DESC"
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrder := new(types.PurchaseOrder)

	for rows.Next() {
		purchaseOrder, err = scanRowIntoPurchaseOrder(rows)

		if err != nil {
			return nil, err
		}
	}

	if purchaseOrder.ID == 0 {
		return nil, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrder, nil
}

func (s *Store) GetPurchaseOrderByID(id int) (*types.PurchaseOrder, error) {
	query := "SELECT * FROM purchase_order WHERE id = ? AND deleted_at IS NULL ORDER BY invoice_date DESC"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrder := new(types.PurchaseOrder)

	for rows.Next() {
		purchaseOrder, err = scanRowIntoPurchaseOrder(rows)

		if err != nil {
			return nil, err
		}
	}

	if purchaseOrder.ID == 0 {
		return nil, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrder, nil
}

func (s *Store) GetPurchaseOrderID(number int, supplierId int, totalItem int, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM purchase_order 
				WHERE number = ? 
				AND supplier_id = ? AND total_item = ? 
				AND invoice_date = ? AND deleted_at IS NULL 
				ORDER BY invoice_date DESC`

	rows, err := s.db.Query(query, number, supplierId, totalItem, invoiceDate)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var purchaseOrderId int

	for rows.Next() {
		err = rows.Scan(&purchaseOrderId)
		if err != nil {
			return 0, err
		}
	}

	if purchaseOrderId == 0 {
		return 0, fmt.Errorf("purchase order invoice not found")
	}

	return purchaseOrderId, nil
}

func (s *Store) GetNumberOfPurchaseOrders() (int, error) {
	query := `SELECT COUNT(*) FROM purchase_order`
	row := s.db.QueryRow(query)
	if row.Err() != nil {
		return -1, row.Err()
	}

	var numberOfPurchaseOrders int

	err := row.Scan(&numberOfPurchaseOrders)
	if err != nil {
		return -1, err
	}

	return numberOfPurchaseOrders, nil
}

func (s *Store) CreatePurchaseOrder(poInvoice types.PurchaseOrder) error {
	values := "?"
	for i := 0; i < 5; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order (
		number, supplier_id, user_id, total_item, 
		invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		poInvoice.Number, poInvoice.SupplierID,
		poInvoice.UserID, poInvoice.TotalItem, poInvoice.InvoiceDate,
		poInvoice.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseOrderItem(purchaseMedItem types.PurchaseOrderItem) error {
	values := "?"
	for i := 0; i < 5; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_order_item (
		purchase_order_id, medicine_id, order_qty, received_qty, unit_id, remarks
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseMedItem.PurchaseOrderID, purchaseMedItem.MedicineID, purchaseMedItem.OrderQty,
		purchaseMedItem.ReceivedQty, purchaseMedItem.UnitID, purchaseMedItem.Remarks)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPurchaseOrdersByDate(startDate time.Time, endDate time.Time) ([]types.PurchaseOrderListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_item, poi.invoice_date 
					FROM purchase_order AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE poi.invoice_date >= ? AND poi.invoice_date < ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrders := make([]types.PurchaseOrderListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrder, err := scanRowIntoPurchaseOrderLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrders = append(purchaseOrders, *purchaseOrder)
	}

	return purchaseOrders, nil
}

func (s *Store) GetPurchaseOrdersByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]types.PurchaseOrderListsReturnPayload, error) {
	query := `SELECT COUNT(*)
					FROM purchase_order 
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

	purchaseOrders := make([]types.PurchaseOrderListsReturnPayload, 0)

	if count == 0 {
		query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_item, poi.invoice_date 
					FROM purchase_order AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE poi.invoice_date >= ? AND poi.invoice_date < ? 
					AND poi.number LIKE ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

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
			purchaseOrder, err := scanRowIntoPurchaseOrderLists(rows)
			if err != nil {
				return nil, err
			}

			purchaseOrders = append(purchaseOrders, *purchaseOrder)
		}

		return purchaseOrders, nil
	}

	query = `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_item, poi.invoice_date 
					FROM purchase_order AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE poi.invoice_date >= ? AND poi.invoice_date < ? 
					AND poi.number = ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		purchaseOrder, err := scanRowIntoPurchaseOrderLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrders = append(purchaseOrders, *purchaseOrder)
	}

	return purchaseOrders, nil
}

func (s *Store) GetPurchaseOrdersByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.PurchaseOrderListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_item, poi.invoice_date 
					FROM purchase_order AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE poi.invoice_date >= ? AND poi.invoice_date < ? 
					AND poi.user_id = ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrders := make([]types.PurchaseOrderListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrder, err := scanRowIntoPurchaseOrderLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrders = append(purchaseOrders, *purchaseOrder)
	}

	return purchaseOrders, nil
}

func (s *Store) GetPurchaseOrdersByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]types.PurchaseOrderListsReturnPayload, error) {
	query := `SELECT poi.id, poi.number, 
					supplier.name, user.name, 
					poi.total_item, poi.invoice_date 
					FROM purchase_order AS poi 
					JOIN supplier ON poi.supplier_id = supplier.id 
					JOIN user ON poi.user_id = user.id 
					WHERE poi.invoice_date >= ? AND poi.invoice_date < ? 
					AND poi.supplier_id = ? 
					AND poi.deleted_at IS NULL 
					ORDER BY poi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrders := make([]types.PurchaseOrderListsReturnPayload, 0)

	for rows.Next() {
		purchaseOrder, err := scanRowIntoPurchaseOrderLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrders = append(purchaseOrders, *purchaseOrder)
	}

	return purchaseOrders, nil
}

func (s *Store) GetPurchaseOrderItem(purchaseOrderId int) ([]types.PurchaseOrderItemReturn, error) {
	query := `SELECT 
				poit.id, 
				medicine.barcode, medicine.name, 
				poit.order_qty, poit.received_qty, 
				unit.name, 
				poit.remarks  
				FROM purchase_order_item as poit 
				JOIN purchase_order as poin 
					ON poit.purchase_order_id = poin.id 
				JOIN medicine ON poit.medicine_id = medicine.id 
				JOIN unit ON poit.unit_id = unit.id 
				WHERE poin.id = ? AND poin.deleted_at IS NULL 
				ORDER BY poin.invoice_date DESC`

	rows, err := s.db.Query(query, purchaseOrderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseOrderItems := make([]types.PurchaseOrderItemReturn, 0)

	for rows.Next() {
		purchaseOrderItem, err := scanRowIntoPurchaseOrderItem(rows)

		if err != nil {
			return nil, err
		}

		purchaseOrderItems = append(purchaseOrderItems, *purchaseOrderItem)
	}

	return purchaseOrderItems, nil
}

func (s *Store) DeletePurchaseOrder(purchaseOrder *types.PurchaseOrder, user *types.User) error {
	query := "UPDATE purchase_order SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, purchaseOrder.ID)
	if err != nil {
		return err
	}

	data, err := s.GetPurchaseOrderByID(purchaseOrder.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "purchase-order", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeletePurchaseOrderItem(purchaseOrder *types.PurchaseOrder, user *types.User) error {
	data, err := s.GetPurchaseOrderItem(purchaseOrder.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"purchase_order":        purchaseOrder,
		"deleted_medicine_item": data,
	}

	err = logger.WriteLog("delete", "purchase-order", user.Name, purchaseOrder.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM purchase_order_item WHERE purchase_order_id = ? ", purchaseOrder.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseOrder(poiid int, purchaseOrder types.PurchaseOrder, user *types.User) error {
	data, err := s.GetPurchaseOrderByID(poiid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_dataa": data,
	}

	err = logger.WriteLog("modify", "purchase-order", user.Name, poiid, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE purchase_order 
				SET number = ?, supplier_id = ?, total_item = ?, 
				invoice_date = ?, last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query,
		purchaseOrder.Number, purchaseOrder.SupplierID,
		purchaseOrder.TotalItem, purchaseOrder.InvoiceDate,
		time.Now(), purchaseOrder.LastModifiedByUserID, poiid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdtaeReceivedQty(poinid int, newQty float64, user *types.User, mid int) error {
	data, err := s.GetPurchaseOrderItem(poinid)
	if err != nil {
		return err
	}

	purchaseOrder, err := s.GetPurchaseOrderByID(poinid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"purchase_order": purchaseOrder,
		"previous_data":  data,
	}

	err = logger.WriteLog("modify", "purchase-order", user.Name, purchaseOrder.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE purchase_order_item 
				SET received_qty = ? WHERE purchase_order_id = ? AND medicine_id = ?`

	_, err = s.db.Exec(query, newQty, poinid, mid)
	if err != nil {
		return err
	}

	query = `UPDATE purchase_order 
				SET last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query, time.Now(), user.ID, poinid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AbsoluteDeletePurchaseOrder(poi types.PurchaseOrder) error {
	query := `SELECT id FROM purchase_order 
				WHERE number = ? 
				AND supplier_id = ? AND total_item = ? 
				AND invoice_date = ?`

	rows, err := s.db.Query(query, poi.Number, poi.SupplierID, poi.TotalItem, poi.InvoiceDate)
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

	query = "DELETE FROM purchase_order_item WHERE purchase_order_id = ?"
	_, _ = s.db.Exec(query, id)

	query = `DELETE FROM purchase_order WHERE id = ?`
	_, _ = s.db.Exec(query, id)

	return nil
}

func (s *Store) UpdatePdfUrl(poId int, pdfUrl string) error {
	query := `UPDATE purchase_order SET pdf_url = ? WHERE id = ?`
	_, err := s.db.Exec(query, pdfUrl, poId)
	if err != nil {
		return err
	}

	return nil
}

// false means doesn't exist
func (s *Store) IsPdfUrlExist(pdfUrl string) (bool, error) {
	query := `SELECT COUNT(*) FROM purchase_order WHERE pdf_url = ?`
	row := s.db.QueryRow(query, pdfUrl)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return (count > 0), nil
}

func scanRowIntoPurchaseOrder(rows *sql.Rows) (*types.PurchaseOrder, error) {
	purchaseOrder := new(types.PurchaseOrder)

	err := rows.Scan(
		&purchaseOrder.ID,
		&purchaseOrder.Number,
		&purchaseOrder.SupplierID,
		&purchaseOrder.UserID,
		&purchaseOrder.TotalItem,
		&purchaseOrder.InvoiceDate,
		&purchaseOrder.CreatedAt,
		&purchaseOrder.LastModified,
		&purchaseOrder.LastModifiedByUserID,
		&purchaseOrder.DeletedAt,
		&purchaseOrder.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	purchaseOrder.InvoiceDate = purchaseOrder.InvoiceDate.Local()
	purchaseOrder.CreatedAt = purchaseOrder.CreatedAt.Local()
	purchaseOrder.LastModified = purchaseOrder.LastModified.Local()

	return purchaseOrder, nil
}

func scanRowIntoPurchaseOrderLists(rows *sql.Rows) (*types.PurchaseOrderListsReturnPayload, error) {
	purchaseOrder := new(types.PurchaseOrderListsReturnPayload)

	err := rows.Scan(
		&purchaseOrder.ID,
		&purchaseOrder.Number,
		&purchaseOrder.SupplierName,
		&purchaseOrder.UserName,
		&purchaseOrder.TotalItem,
		&purchaseOrder.InvoiceDate,
	)

	if err != nil {
		return nil, err
	}

	purchaseOrder.InvoiceDate = purchaseOrder.InvoiceDate.Local()

	return purchaseOrder, nil
}

func scanRowIntoPurchaseOrderItem(rows *sql.Rows) (*types.PurchaseOrderItemReturn, error) {
	purchaseOrderItem := new(types.PurchaseOrderItemReturn)

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
