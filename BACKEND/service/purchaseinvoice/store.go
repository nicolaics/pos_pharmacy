package purchaseinvoice

import (
	"database/sql"
	"fmt"
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

func (s *Store) GetPurchaseInvoicesByNumber(number int) ([]types.PurchaseInvoice, error) {
	query := "SELECT * FROM purchase_invoice WHERE number = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoice, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoice(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseInvoiceByID(id int) (*types.PurchaseInvoice, error) {
	query := "SELECT * FROM purchase_invoice WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoice := new(types.PurchaseInvoice)

	for rows.Next() {
		purchaseInvoice, err = scanRowIntoPurchaseInvoice(rows)

		if err != nil {
			return nil, err
		}
	}

	if purchaseInvoice.ID == 0 {
		return nil, fmt.Errorf("purchase invoice not found")
	}

	return purchaseInvoice, nil
}

func (s *Store) GetPurchaseInvoiceID(number int, companyId int, supplierId int, subtotal float64, totalPrice float64, userId int, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM purchase_invoice 
				WHERE number = ? AND company_id = ? AND supplier_id = ? 
				AND subtotal = ? AND total_price = ? AND user_id = ? AND invoice_date = ? 
				AND deleted_at IS NULL`

	rows, err := s.db.Query(query, number, companyId, supplierId, subtotal, totalPrice, userId, invoiceDate)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var purchaseInvoiceId int

	for rows.Next() {
		err = rows.Scan(&purchaseInvoiceId)
		if err != nil {
			return -1, err
		}
	}

	if purchaseInvoiceId == 0 {
		return -1, fmt.Errorf("purchase invoice not found")
	}

	return purchaseInvoiceId, nil
}

func (s *Store) CreatePurchaseInvoice(purchaseInvoice types.PurchaseInvoice) error {
	values := "?"
	for i := 0; i < 10; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_invoice (
		number, company_id, supplier_id, subtotal, discount, tax, total_price, 
		description, user_id, invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseInvoice.Number, purchaseInvoice.CompanyID, purchaseInvoice.SupplierID,
		purchaseInvoice.Subtotal, purchaseInvoice.Discount, purchaseInvoice.Tax,
		purchaseInvoice.TotalPrice, purchaseInvoice.Description, purchaseInvoice.UserID,
		purchaseInvoice.InvoiceDate, purchaseInvoice.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseMedicineItems(purchaseMedItem types.PurchaseMedicineItem) error {
	values := "?"
	for i := 0; i < 9; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_medicine_items (
				purchase_invoice_id, medicine_id, qty, unit_id, 
				purchase_price, purchase_discount, purchase_tax, 
				subtotal, batch_number, expired_date
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseMedItem.PurchaseInvoiceID, purchaseMedItem.MedicineID, purchaseMedItem.Qty,
		purchaseMedItem.UnitID, purchaseMedItem.PurchasePrice, purchaseMedItem.PurchaseDiscount,
		purchaseMedItem.PurchaseTax, purchaseMedItem.Subtotal, purchaseMedItem.BatchNumber,
		purchaseMedItem.ExpDate)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPurchaseInvoicesByDate(startDate time.Time, endDate time.Time) ([]types.PurchaseInvoice, error) {
	query := `SELECT * FROM purchase_invoice 
				WHERE (invoice_date BETWEEN DATE(?) AND DATE(?)) 
				AND deleted_at IS NULL 
				ORDER BY invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoice, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoice(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseMedicineItems(purchaseInvoiceId int) ([]types.PurchaseMedicineItemsReturn, error) {
	query := `SELECT 
			pmi.id, 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.purchase_price, pmi.purchase_discount, 
			pmi.purchase_tax, pmi.subtotal, pmi.batch_number, pmi.expired_date 
			
			FROM purchase_medicine_items as pmi 
			JOIN purchase_invoice as pi ON pmi.purchase_invoice_id = pi.id 
			JOIN medicine ON pmi.medicine_id = medicine.id 
			JOIN unit ON pmi.unit_id = unit.id 
			WHERE pi.id = ? AND pi.deleted_at IS NULL`

	rows, err := s.db.Query(query, purchaseInvoiceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (s *Store) DeletePurchaseInvoice(purchaseInvoice *types.PurchaseInvoice, userId int) error {
	query := "UPDATE purchase_invoice SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, purchaseInvoice.ID)
	if err != nil {
		return err
	}

	data, err := s.GetPurchaseInvoiceByID(purchaseInvoice.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "purchase-invoice", userId, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeletePurchaseMedicineItems(purchaseInvoice *types.PurchaseInvoice, userId int) error {
	data, err := s.GetPurchaseMedicineItems(purchaseInvoice.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"purchase_invoice": purchaseInvoice,
		"deleted_medicine_items": data,
	}

	err = logger.WriteLog("delete", "purchase-invoice", userId, purchaseInvoice.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM purchase_medicine_items WHERE purchase_invoice_id = ? ", purchaseInvoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseInvoice(piid int, purchaseInvoice types.PurchaseInvoice, userId int) error {
	data, err := s.GetPurchaseInvoiceByID(piid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "purchase-invoice", userId, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE purchase_invoice SET 
				number = ?, company_id = ?, supplier_id = ?, subtotal = ?, discount = ?, 
				tax = ?, total_price = ?, description = ?, invoice_date = ?, last_modified = ?,
				last_modified_by_user_id = ? 
				 WHERE id = ?`

	_, err = s.db.Exec(query,
		purchaseInvoice.Number, purchaseInvoice.CompanyID, purchaseInvoice.SupplierID,
		purchaseInvoice.Subtotal, purchaseInvoice.Discount, purchaseInvoice.Tax,
		purchaseInvoice.TotalPrice, purchaseInvoice.Description, purchaseInvoice.InvoiceDate,
		time.Now(), purchaseInvoice.LastModifiedByUserID, piid)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPurchaseInvoice(rows *sql.Rows) (*types.PurchaseInvoice, error) {
	purchaseInvoice := new(types.PurchaseInvoice)

	err := rows.Scan(
		&purchaseInvoice.ID,
		&purchaseInvoice.Number,
		&purchaseInvoice.CompanyID,
		&purchaseInvoice.SupplierID,
		&purchaseInvoice.Subtotal,
		&purchaseInvoice.Discount,
		&purchaseInvoice.Tax,
		&purchaseInvoice.TotalPrice,
		&purchaseInvoice.Description,
		&purchaseInvoice.UserID,
		&purchaseInvoice.InvoiceDate,
		&purchaseInvoice.CreatedAt,
		&purchaseInvoice.LastModified,
		&purchaseInvoice.LastModifiedByUserID,
		&purchaseInvoice.DeletedAt,
		&purchaseInvoice.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	purchaseInvoice.InvoiceDate = purchaseInvoice.InvoiceDate.Local()
	purchaseInvoice.CreatedAt = purchaseInvoice.CreatedAt.Local()
	purchaseInvoice.LastModified = purchaseInvoice.LastModified.Local()

	return purchaseInvoice, nil
}

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
