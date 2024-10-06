package production

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

func (s *Store) GetProductionByNumber(number int) (*types.Production, error) {
	query := "SELECT * FROM production WHERE number = ? AND deleted_at IS NULL ORDER BY production_date DESC"
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
	query := "SELECT * FROM production WHERE id = ? AND deleted_at IS NULL ORDER BY production_date DESC"
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

func (s *Store) GetProductionsByDate(startDate time.Time, endDate time.Time) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					ORDER BY prod.production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionsByDateAndNumber(startDate time.Time, endDate time.Time, bn int) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					AND prod.number LIKE ? 
					ORDER BY prod.production_date DESC`

	searchVal := "%"
	for _, val := range strconv.Itoa(bn) {
		if string(val) != " " {
			searchVal += (string(val) + "%")
		}
	}

	rows, err := s.db.Query(query, startDate, endDate, searchVal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionsByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					AND prod.user_id = ? 
					ORDER BY prod.production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionsByDateAndMedicineID(startDate time.Time, endDate time.Time, mid int) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					AND prod.produced_medicine_id = ? 
					ORDER BY prod.production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, mid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionsByDateAndUpdatedToStock(startDate time.Time, endDate time.Time, uts bool) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					AND prod.updated_to_stock = ? 
					ORDER BY prod.production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
}

func (s *Store) GetProductionsByDateAndUpdatedToAccount(startDate time.Time, endDate time.Time, uta bool) ([]types.ProductionListsReturnPayload, error) {
	query := `SELECT prod.id, prod.number, 
					med.name AS produced_medicine_name, 
					prod.produced_qty, 
					unit.name, 
					prod.production_date, 
					prod.description, prod.updated_to_stock, 
					prod.updated_to_account, prod.total_cost,
					user.name 
					FROM production AS prod 
					JOIN medicine AS med ON prod.produced_medicine_id = med.id 
					JOIN user ON prod.user_id = user.id 
					JOIN unit ON unit.id = prod.produced_unit_id 
					WHERE (prod.production_date BETWEEN DATE(?) AND DATE(?)) 
					AND prod.deleted_at IS NULL 
					AND prod.updated_to_account = ? 
					ORDER BY prod.production_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uta)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productions := make([]types.ProductionListsReturnPayload, 0)

	for rows.Next() {
		production, err := scanRowIntoProductionLists(rows)

		if err != nil {
			return nil, err
		}

		productions = append(productions, *production)
	}

	return productions, nil
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
	for i := 0; i < 10; i++ {
		values += ", ?"
	}

	query := `INSERT INTO production (
		number, produced_medicine_id, produced_qty, produced_unit_id, production_date, description,  
		updated_to_stock, updated_to_account, total_cost, user_id, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		production.Number, production.ProducedMedicineID, production.ProducedQty, production.ProducedUnitID,
		production.ProductionDate, production.Description, production.UpdatedToStock,
		production.UpdatedToAccount, production.TotalCost, production.UserID, production.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateProductionMedicineItem(prodMedItem types.ProductionMedicineItem) error {
	values := "?"
	for i := 0; i < 4; i++ {
		values += ", ?"
	}

	query := `INSERT INTO production_medicine_item (
				production_id, medicine_id, qty, unit_id, cost
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		prodMedItem.ProductionID, prodMedItem.MedicineID,
		prodMedItem.Qty, prodMedItem.UnitID, prodMedItem.Cost)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetProductionMedicineItem(productionId int) ([]types.ProductionMedicineItemRow, error) {
	query := `SELECT 
			pmi.id, 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.cost 
			
			FROM production_medicine_item as pmi 
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
		prescMedItem, err := scanRowIntoProductionMedicineItem(rows)

		if err != nil {
			return nil, err
		}

		prescMedItems = append(prescMedItems, *prescMedItem)
	}

	return prescMedItems, nil
}

func (s *Store) DeleteProduction(production *types.Production, user *types.User) error {
	query := "UPDATE production SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, production.ID)
	if err != nil {
		return err
	}

	data, err := s.GetProductionByID(production.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "production", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeleteProductionMedicineItem(production *types.Production, user *types.User) error {
	data, err := s.GetProductionMedicineItem(production.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"production":            production,
		"deleted_medicine_item": data,
	}

	err = logger.WriteLog("delete", "prescription", user.Name, production.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM production_medicine_item WHERE production_id = ? ", production.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyProduction(id int, production types.Production, user *types.User) error {
	data, err := s.GetProductionByID(production.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "production", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE production SET 
				number = ?, produced_medicine_id = ?, produced_qty = ?, produced_unit_id = ?, production_date = ?, 
				description = ?, updated_to_stock = ?, updated_to_account = ?, total_cost = ?, 
				last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query,
		production.Number, production.ProducedMedicineID, production.ProducedQty, production.ProducedUnitID,
		production.ProductionDate, production.Description, production.UpdatedToStock,
		production.UpdatedToAccount, production.TotalCost, time.Now(), production.LastModifiedByUserID, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AbsoluteDeleteProduction(prod types.Production) error {
	query := `SELECT id FROM production 
				WHERE number = ? AND produced_medicine_id = ? 
				AND produced_qty = ? AND produced_unit_id = ? 
				AND production_date = ? AND description = ? 
				AND updated_to_stock = ? AND updated_to_account = ? 
				AND total_cost = ?`

	rows, err := s.db.Query(query, prod.Number, prod.ProducedMedicineID, prod.ProducedQty, prod.ProducedUnitID,
		prod.ProductionDate, prod.Description, prod.UpdatedToStock, prod.UpdatedToAccount,
		prod.TotalCost)
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

	query = "DELETE FROM production_medicine_item WHERE production_id = ?"
	_, _ = s.db.Exec(query, id)

	query = `DELETE FROM production WHERE id = ?`
	_, _ = s.db.Exec(query, id)

	return nil
}

func scanRowIntoProduction(rows *sql.Rows) (*types.Production, error) {
	production := new(types.Production)

	err := rows.Scan(
		&production.ID,
		&production.Number,
		&production.ProducedMedicineID,
		&production.ProducedQty,
		&production.ProducedUnitID,
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

func scanRowIntoProductionLists(rows *sql.Rows) (*types.ProductionListsReturnPayload, error) {
	production := new(types.ProductionListsReturnPayload)

	err := rows.Scan(
		&production.ID,
		&production.Number,
		&production.ProducedMedicineName,
		&production.ProducedQty,
		&production.ProducedUnit,
		&production.ProductionDate,
		&production.Description,
		&production.UpdatedToStock,
		&production.UpdatedToAccount,
		&production.TotalCost,
		&production.UserName,
	)

	if err != nil {
		return nil, err
	}

	production.ProductionDate = production.ProductionDate.Local()

	return production, nil
}

func scanRowIntoProductionMedicineItem(rows *sql.Rows) (*types.ProductionMedicineItemRow, error) {
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
