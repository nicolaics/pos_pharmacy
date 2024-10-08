package production

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

func (s *Store) GetProductionByBatchNumber(number int) (*types.Production, error) {
	query := "SELECT * FROM production WHERE batch_number = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	production := new(types.Production)

	for rows.Next() {
		production, err = scanRowIntoProduction(rows)

		if err != nil {
			return nil, err
		}
	}

	if production.ID == 0 {
		return nil, fmt.Errorf("production not found")
	}

	return production, nil
}

func (s *Store) GetProductionByID(id int) (*types.Production, error) {
	query := "SELECT * FROM production WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	production := new(types.Production)

	for rows.Next() {
		production, err = scanRowIntoProduction(rows)

		if err != nil {
			return nil, err
		}
	}

	if production.ID == 0 {
		return nil, fmt.Errorf("production not found")
	}

	return production, nil
}

func (s *Store) GetProductionsByDate(startDate time.Time, endDate time.Time) ([]types.Production, error) {
	query := `SELECT * FROM production 
				WHERE (production_date BETWEEN DATE(?) AND DATE(?)) 
				AND deleted_at IS NULL 
				ORDER BY production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.Production, 0)

	for rows.Next() {
		production, err := scanRowIntoProduction(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionID(batchNumber int, producedMedId int, prodDate time.Time, totalCost float64, userId int) (int, error) {
	query := `SELECT id FROM production 
				WHERE batch_number = ? AND produced_med_id = ? AND production_date = ? 
				AND total_cost = ? AND user_id = ? 
				AND deleted_at IS NULL`

	rows, err := s.db.Query(query, batchNumber, producedMedId, prodDate, totalCost, userId)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var productionId int

	for rows.Next() {
		err = rows.Scan(&productionId)
		if err != nil {
			return -1, err
		}
	}

	if productionId == 0 {
		return -1, fmt.Errorf("production not found")
	}

	return productionId, nil
}

func (s *Store) GetNumberOfProductions() (int, error) {
	query := `SELECT COUNT(*) FROM production 
				WHERE deleted_at IS NULL`

	row := s.db.QueryRow(query)
	if row.Err() != nil {
		return -1, row.Err()
	}

	var numberOfProductions int

	err := row.Scan(&numberOfProductions)
	if err != nil {
		return -1, err
	}

	return numberOfProductions, nil
}

func (s *Store) CreateProduction(production types.Production) error {
	values := "?"
	for i := 0; i < 9; i++ {
		values += ", ?"
	}

	query := `INSERT INTO production (
		batch_number, produced_medicine_id, produced_qty, production_date, description,  
		updated_to_stock, updated_to_account, total_cost, user_id, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		production.BatchNumber, production.ProducedMedicineID, production.ProducedQty,
		production.ProductionDate, production.Description, production.UpdatedToStock,
		production.UpdatedToAccount, production.TotalCost, production.UserID, production.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateProductionMedicineItems(prodMedItems types.ProductionMedicineItems) error {
	values := "?"
	for i := 0; i < 4; i++ {
		values += ", ?"
	}

	query := `INSERT INTO production_medicine_items (
				production_id, medicine_id, qty, unit_id, cost
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		prodMedItems.ProductionID, prodMedItems.MedicineID,
		prodMedItems.Qty, prodMedItems.UnitID, prodMedItems.Cost)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetProductionMedicineItems(productionId int) ([]types.ProductionMedicineItemRow, error) {
	query := `SELECT 
			pmi.id, 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.cost 
			
			FROM production_medicine_items as pmi 
			JOIN production as prod ON pmi.production_id = prod.id 
			JOIN medicine ON pmi.medicine_id = medicine.id 
			JOIN unit ON pmi.unit_id = unit.id 
			WHERE prod.id = ? AND prod.deleted_at IS NULL`

	rows, err := s.db.Query(query, productionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescMedItems := make([]types.ProductionMedicineItemRow, 0)

	for rows.Next() {
		prescMedItem, err := scanRowIntoProductionMedicineItems(rows)

		if err != nil {
			return nil, err
		}

		prescMedItems = append(prescMedItems, *prescMedItem)
	}

	return prescMedItems, nil
}

func (s *Store) DeleteProduction(production *types.Production, userId int) error {
	query := "UPDATE production SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, production.ID)
	if err != nil {
		return err
	}

	data, err := s.GetProductionByID(production.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "production", userId, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeleteProductionMedicineItems(production *types.Production, userId int) error {
	data, err := s.GetProductionMedicineItems(production.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"production": production,
		"deleted_medicine_items": data,
	}

	err = logger.WriteLog("delete", "prescription", userId, production.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM production_medicine_items WHERE production_id = ? ", production.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyProduction(id int, production types.Production, userId int) error {
	data, err := s.GetProductionByID(production.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "production", userId, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE production SET 
				batch_number = ?, produced_medicine_id = ?, produced_qty = ?, production_date = ?, 
				description = ?, updated_to_stock = ?, updated_to_account = ?, total_cost = ?, 
				last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query,
		production.BatchNumber, production.ProducedMedicineID, production.ProducedQty,
		production.ProductionDate, production.Description, production.UpdatedToStock,
		production.UpdatedToAccount, production.TotalCost, time.Now(), production.LastModifiedByUserID, id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoProduction(rows *sql.Rows) (*types.Production, error) {
	production := new(types.Production)

	err := rows.Scan(
		&production.ID,
		&production.BatchNumber,
		&production.ProducedMedicineID,
		&production.ProducedQty,
		&production.ProductionDate,
		&production.Description,
		&production.UpdatedToStock,
		&production.UpdatedToAccount,
		&production.TotalCost,
		&production.UserID,
		&production.CreatedAt,
		&production.LastModified,
		&production.LastModifiedByUserID,
		&production.DeletedAt,
		&production.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	production.ProductionDate = production.ProductionDate.Local()
	production.CreatedAt = production.CreatedAt.Local()
	production.LastModified = production.LastModified.Local()

	return production, nil
}

func scanRowIntoProductionMedicineItems(rows *sql.Rows) (*types.ProductionMedicineItemRow, error) {
	prescMedItem := new(types.ProductionMedicineItemRow)

	err := rows.Scan(
		&prescMedItem.ID,
		&prescMedItem.MedicineBarcode,
		&prescMedItem.MedicineName,
		&prescMedItem.Qty,
		&prescMedItem.Unit,
		&prescMedItem.Cost,
	)

	if err != nil {
		return nil, err
	}

	return prescMedItem, nil
}
