package purchaseinvoice

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

func (s *Store) GetPurchaseInvoicesByNumber(number int) ([]types.PurchaseInvoice, error) {
	query := "SELECT * FROM purchase_invoice WHERE number = ? AND deleted_at IS NULL ORDER BY invoice_date DESC"
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
	query := "SELECT * FROM purchase_invoice WHERE id = ? AND deleted_at IS NULL ORDER BY invoice_date DESC"
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

func (s *Store) GetPurchaseInvoiceID(number int, supplierId int, subtotal float64, totalPrice float64, invoiceDate time.Time) (int, error) {
	query := `SELECT id FROM purchase_invoice 
				WHERE number = ? AND supplier_id = ? 
				AND subtotal = ? AND total_price = ? AND invoice_date = ? 
				AND deleted_at IS NULL 
				ORDER BY invoice_date DESC`

	rows, err := s.db.Query(query, number, supplierId, subtotal, totalPrice, invoiceDate)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var purchaseInvoiceId int

	for rows.Next() {
		err = rows.Scan(&purchaseInvoiceId)
		if err != nil {
			return 0, err
		}
	}

	if purchaseInvoiceId == 0 {
		return 0, fmt.Errorf("purchase invoice not found")
	}

	return purchaseInvoiceId, nil
}

func (s *Store) CreatePurchaseInvoice(purchaseInvoice types.PurchaseInvoice) error {
	values := "?"
	for i := 0; i < 12; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_invoice (
		number, supplier_id, purchase_order_number, subtotal, discount_percentage, 
		discount_amount, tax_percentage, tax_amount,  
		total_price, description, user_id, invoice_date, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseInvoice.Number, purchaseInvoice.SupplierID,
		purchaseInvoice.PurchaseOrderNumber, purchaseInvoice.Subtotal,
		purchaseInvoice.DiscountPercentage, purchaseInvoice.DiscountAmount, 
		purchaseInvoice.TaxPercentage, purchaseInvoice.TaxAmount, purchaseInvoice.TotalPrice,
		purchaseInvoice.Description, purchaseInvoice.UserID, purchaseInvoice.InvoiceDate,
		purchaseInvoice.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePurchaseMedicineItem(purchaseMedItem types.PurchaseMedicineItem) error {
	values := "?"
	for i := 0; i < 11; i++ {
		values += ", ?"
	}

	query := `INSERT INTO purchase_medicine_item (
				purchase_invoice_id, medicine_id, qty, unit_id, 
				price, discount_percentage, discount_amount, 
				tax_percentage, tax_amount, 
				subtotal, batch_number, expired_date
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		purchaseMedItem.PurchaseInvoiceID, purchaseMedItem.MedicineID, purchaseMedItem.Qty,
		purchaseMedItem.UnitID, purchaseMedItem.Price, purchaseMedItem.DiscountPercentage,
		purchaseMedItem.DiscountAmount, purchaseMedItem.TaxPercentage, 
		purchaseMedItem.TaxAmount, purchaseMedItem.Subtotal, purchaseMedItem.BatchNumber,
		purchaseMedItem.ExpDate)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPurchaseMedicineItem(purchaseInvoiceId int) ([]types.PurchaseMedicineItemReturn, error) {
	query := `SELECT 
			pmi.id, 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.price, pmi.discount_percentage, pmi.discount_amount, 
			pmi.tax_percentage, pmi.tax_amount, 
			pmi.subtotal, pmi.batch_number, pmi.expired_date 
			FROM purchase_medicine_item as pmi 
			JOIN purchase_invoice as pi ON pmi.purchase_invoice_id = pi.id 
			JOIN medicine ON pmi.medicine_id = medicine.id 
			JOIN unit ON pmi.unit_id = unit.id 
			WHERE pi.id = ? AND pi.deleted_at IS NULL`

	rows, err := s.db.Query(query, purchaseInvoiceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseMedicineItems := make([]types.PurchaseMedicineItemReturn, 0)

	for rows.Next() {
		purchaseMedicineItem, err := scanRowIntoPurchaseMedicineItem(rows)

		if err != nil {
			return nil, err
		}

		purchaseMedicineItems = append(purchaseMedicineItems, *purchaseMedicineItem)
	}

	return purchaseMedicineItems, nil
}

func (s *Store) GetPurchaseInvoicesByDate(startDate time.Time, endDate time.Time) ([]types.PurchaseInvoiceListsReturnPayload, error) {
	query := `SELECT pi.id, pi.number, 
				supplier.name, 
				pi.purchase_order_number, 
				pi.total_price, pi.description, 
				user.name, 
				pi.invoice_date, pi.pdf_url  
				FROM purchase_invoice AS pi 
				JOIN supplier ON supplier.id = pi.supplier_id 
				JOIN user ON user.id = pi.user_id 
				WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
				AND pi.deleted_at IS NULL 
				ORDER BY pi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]types.PurchaseInvoiceListsReturnPayload, error) {
	query := `SELECT COUNT(*) 
				FROM purchase_invoice 
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

	purchaseInvoices := make([]types.PurchaseInvoiceListsReturnPayload, 0)

	if count == 0 {
		query = `SELECT pi.id, pi.number, 
					supplier.name, 
					pi.purchase_order_number, 
					pi.total_price, pi.description, 
					user.name, 
					pi.invoice_date, pi.pdf_url 
					FROM purchase_invoice AS pi 
					JOIN supplier ON supplier.id = pi.supplier_id 
					JOIN user ON user.id = pi.user_id 
					WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
					AND number LIKE ?
					AND pi.deleted_at IS NULL 
					ORDER BY pi.invoice_date DESC`

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
			purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)
			if err != nil {
				return nil, err
			}

			purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
		}

		return purchaseInvoices, nil
	}

	query = `SELECT pi.id, pi.number, 
					supplier.name, 
					pi.purchase_order_number, 
					pi.total_price, pi.description, 
					user.name, 
					pi.invoice_date 
					FROM purchase_invoice AS pi 
					JOIN supplier ON supplier.id = pi.supplier_id 
					JOIN user ON user.id = pi.user_id 
					WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
					AND number = ?
					AND pi.deleted_at IS NULL 
					ORDER BY pi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseInvoicesByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]types.PurchaseInvoiceListsReturnPayload, error) {
	query := `SELECT pi.id, pi.number, 
				supplier.name, 
				pi.purchase_order_number, 
				pi.total_price, pi.description, 
				user.name, 
				pi.invoice_date, pi.pdf_url 
				FROM purchase_invoice AS pi 
				JOIN supplier ON supplier.id = pi.supplier_id 
				JOIN user ON user.id = pi.user_id 
				WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
				AND pi.supplier_id = ? 
				AND pi.deleted_at IS NULL 
				ORDER BY pi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.PurchaseInvoiceListsReturnPayload, error) {
	query := `SELECT pi.id, pi.number, 
				supplier.name, 
				pi.purchase_order_number, 
				pi.total_price, pi.description, 
				user.name, 
				pi.invoice_date, pi.pdf_url 
				FROM purchase_invoice AS pi 
				JOIN supplier ON supplier.id = pi.supplier_id 
				JOIN user ON user.id = pi.user_id 
				WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
				AND pi.user_id = ? 
				AND pi.deleted_at IS NULL 
				ORDER BY pi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) GetPurchaseInvoicesByDateAndPONumber(startDate time.Time, endDate time.Time, poiNumber int) ([]types.PurchaseInvoiceListsReturnPayload, error) {
	query := `SELECT pi.id, pi.number, 
				supplier.name, 
				pi.purchase_order_number, 
				pi.total_price, pi.description, 
				user.name, 
				pi.invoice_date, pi.pdf_url 
				FROM purchase_invoice AS pi 
				JOIN supplier ON supplier.id = pi.supplier_id 
				JOIN user ON user.id = pi.user_id 
				WHERE pi.invoice_date >= ? AND pi.invoice_date < ? 
				AND pi.purchase_order_number = ? 
				AND pi.deleted_at IS NULL 
				ORDER BY pi.invoice_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, poiNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	purchaseInvoices := make([]types.PurchaseInvoiceListsReturnPayload, 0)

	for rows.Next() {
		purchaseInvoice, err := scanRowIntoPurchaseInvoiceLists(rows)

		if err != nil {
			return nil, err
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	}

	return purchaseInvoices, nil
}

func (s *Store) DeletePurchaseInvoice(purchaseInvoice *types.PurchaseInvoice, user *types.User) error {
	query := "UPDATE purchase_invoice SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, purchaseInvoice.ID)
	if err != nil {
		return err
	}

	data, err := s.GetPurchaseInvoiceByID(purchaseInvoice.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "purchase-invoice", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeletePurchaseMedicineItem(purchaseInvoice *types.PurchaseInvoice, user *types.User) error {
	data, err := s.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"purchase_invoice":      purchaseInvoice,
		"deleted_medicine_item": data,
	}

	err = logger.WriteLog("delete", "purchase-invoice", user.Name, purchaseInvoice.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM purchase_medicine_item WHERE purchase_invoice_id = ? ", purchaseInvoice.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPurchaseInvoice(piid int, purchaseInvoice types.PurchaseInvoice, user *types.User) error {
	data, err := s.GetPurchaseInvoiceByID(piid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "purchase-invoice", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE purchase_invoice SET 
				number = ?, supplier_id = ?, purchase_order_number = ?, 
				subtotal = ?, discount_percentage = ?, discount_amount = ?, 
				tax_percentage = ?, tax_amount = ?, total_price = ?, description = ?, 
				invoice_date = ?, last_modified = ?, last_modified_by_user_id = ? 
				 WHERE id = ?`

	_, err = s.db.Exec(query,
		purchaseInvoice.Number, purchaseInvoice.SupplierID,
		purchaseInvoice.PurchaseOrderNumber, purchaseInvoice.Subtotal,
		purchaseInvoice.DiscountPercentage, purchaseInvoice.DiscountAmount,
		purchaseInvoice.TaxPercentage, purchaseInvoice.TaxAmount,
		purchaseInvoice.TotalPrice,
		purchaseInvoice.Description, purchaseInvoice.InvoiceDate,
		time.Now(), purchaseInvoice.LastModifiedByUserID, piid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AbsoluteDeletePurchaseInvoice(pi types.PurchaseInvoice) error {
	query := `SELECT id FROM purchase_invoice 
				WHERE number = ? AND supplier_id = ? 
				AND purchase_order_number = ? AND subtotal = ? 
				AND discount_percentage = ? AND discount_amount = ? 
				AND tax_percentage = ? AND tax_amount = ? AND total_price = ? 
				AND description = ? AND invoice_date = ?`

	rows, err := s.db.Query(query, pi.Number, pi.SupplierID, pi.PurchaseOrderNumber,
		pi.Subtotal, pi.DiscountPercentage, pi.DiscountAmount, 
		pi.TaxPercentage, pi.TaxAmount, pi.TotalPrice,
		pi.Description, pi.InvoiceDate)
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

	query = "DELETE FROM purchase_medicine_item WHERE purchase_invoice_id = ?"
	_, _ = s.db.Exec(query, id)

	query = `DELETE FROM purchase_invoice WHERE id = ?`
	_, _ = s.db.Exec(query, id)

	return nil
}

func (s *Store) UpdatePDFUrl(piId int, pdfUrl string) error {
	query := `UPDATE purchase_invoice SET pdf_url = ? WHERE id = ?`
	_, err := s.db.Exec(query, pdfUrl, piId)
	if err != nil {
		return err
	}

	return nil
}

// false means doesn't exist
func (s *Store) IsPDFUrlExist(pdfUrl string) (bool, error) {
	query := `SELECT COUNT(*) FROM purchase_invoice WHERE pdf_url = ?`
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

func scanRowIntoPurchaseInvoice(rows *sql.Rows) (*types.PurchaseInvoice, error) {
	purchaseInvoice := new(types.PurchaseInvoice)

	err := rows.Scan(
		&purchaseInvoice.ID,
		&purchaseInvoice.Number,
		&purchaseInvoice.SupplierID,
		&purchaseInvoice.PurchaseOrderNumber,
		&purchaseInvoice.Subtotal,
		&purchaseInvoice.DiscountPercentage,
		&purchaseInvoice.DiscountAmount,
		&purchaseInvoice.TaxPercentage,
		&purchaseInvoice.TaxAmount,
		&purchaseInvoice.TotalPrice,
		&purchaseInvoice.Description,
		&purchaseInvoice.UserID,
		&purchaseInvoice.InvoiceDate,
		&purchaseInvoice.CreatedAt,
		&purchaseInvoice.LastModified,
		&purchaseInvoice.LastModifiedByUserID,
		&purchaseInvoice.PdfURL,
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

func scanRowIntoPurchaseInvoiceLists(rows *sql.Rows) (*types.PurchaseInvoiceListsReturnPayload, error) {
	purchaseInvoice := new(types.PurchaseInvoiceListsReturnPayload)

	err := rows.Scan(
		&purchaseInvoice.ID,
		&purchaseInvoice.Number,
		&purchaseInvoice.SupplierName,
		&purchaseInvoice.PurchaseOrderNumber,
		&purchaseInvoice.TotalPrice,
		&purchaseInvoice.Description,
		&purchaseInvoice.UserName,
		&purchaseInvoice.InvoiceDate,
		&purchaseInvoice.PdfURL,
	)

	if err != nil {
		return nil, err
	}

	purchaseInvoice.InvoiceDate = purchaseInvoice.InvoiceDate.Local()

	return purchaseInvoice, nil
}

func scanRowIntoPurchaseMedicineItem(rows *sql.Rows) (*types.PurchaseMedicineItemReturn, error) {
	purchaseMedicineItem := new(types.PurchaseMedicineItemReturn)

	err := rows.Scan(
		&purchaseMedicineItem.ID,
		&purchaseMedicineItem.MedicineBarcode,
		&purchaseMedicineItem.MedicineName,
		&purchaseMedicineItem.Qty,
		&purchaseMedicineItem.Unit,
		&purchaseMedicineItem.Price,
		&purchaseMedicineItem.DiscountPercentage,
		&purchaseMedicineItem.DiscountAmount,
		&purchaseMedicineItem.TaxPercentage,
		&purchaseMedicineItem.TaxAmount,
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
