package medicine

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

func (s *Store) GetMedicineByName(name string) (*types.Medicine, error) {
	query := "SELECT * FROM medicine WHERE name = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicine := new(types.Medicine)

	for rows.Next() {
		medicine, err = scanRowIntoMedicine(rows)

		if err != nil {
			return nil, err
		}
	}

	if medicine.ID == 0 {
		return nil, fmt.Errorf("medicine not found")
	}

	return medicine, nil
}

func (s *Store) GetMedicineByID(id int) (*types.MedicineListsReturnPayload, error) {
	query := `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.id = ? 
					AND med.deleted_at IS NULL`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicine := new(types.MedicineListsReturnPayload)

	for rows.Next() {
		medicine, err = scanRowIntoMedicineLists(rows)

		if err != nil {
			return nil, err
		}
	}

	if medicine.ID == 0 {
		return nil, fmt.Errorf("medicine not found")
	}

	return medicine, nil
}

func (s *Store) GetMedicineByBarcode(barcode string) (*types.Medicine, error) {
	query := "SELECT * FROM medicine WHERE barcode = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, barcode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicine := new(types.Medicine)

	for rows.Next() {
		medicine, err = scanRowIntoMedicine(rows)

		if err != nil {
			return nil, err
		}
	}

	if medicine.ID == 0 {
		return nil, fmt.Errorf("medicine not found")
	}

	return medicine, nil
}

func (s *Store) GetMedicinesBySearchName(name string) ([]types.MedicineListsReturnPayload, error) {
	query := "SELECT COUNT(*) FROM medicine WHERE name = ? AND deleted_at IS NULL"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	medicines := make([]types.MedicineListsReturnPayload, 0)

	if count == 0 {
		query = `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.name LIKE ? 
					AND med.deleted_at IS NULL`
		searchVal := "%"


		for _, val := range name {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			medicine, err := scanRowIntoMedicineLists(rows)

			if err != nil {
				return nil, err
			}

			medicines = append(medicines, *medicine)
		}

		return medicines, nil
	}

	query = `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.name = ? 
					AND med.deleted_at IS NULL`
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		medicine, err := scanRowIntoMedicineLists(rows)

		if err != nil {
			return nil, err
		}

		medicines = append(medicines, *medicine)
	}

	return medicines, nil
}

func (s *Store) GetMedicinesBySearchBarcode(barcode string) ([]types.MedicineListsReturnPayload, error) {
	query := "SELECT COUNT(*) FROM medicine WHERE barcode = ? AND deleted_at IS NULL"
	row := s.db.QueryRow(query, barcode)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	medicines := make([]types.MedicineListsReturnPayload, 0)

	if count == 0 {
		query = `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.barcode LIKE ? 
					AND med.deleted_at IS NULL`
		searchVal := "%"


		for _, val := range barcode {
			if string(val) != " " {
				searchVal += (string(val) + "%")
			}
		}

		rows, err := s.db.Query(query, searchVal)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			medicine, err := scanRowIntoMedicineLists(rows)

			if err != nil {
				return nil, err
			}

			medicines = append(medicines, *medicine)
		}

		return medicines, nil
	}

	query = `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.barcode = ? 
					AND med.deleted_at IS NULL`
	rows, err := s.db.Query(query, barcode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		medicine, err := scanRowIntoMedicineLists(rows)

		if err != nil {
			return nil, err
		}

		medicines = append(medicines, *medicine)
	}

	return medicines, nil
}

func (s *Store) GetMedicinesByDescription(description string) ([]types.MedicineListsReturnPayload, error) {
	query := `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.description LIKE ? AND med.deleted_at IS NULL`

	searchVal := "%"
	for _, val := range description {
		if string(val) != " " {
			searchVal += (string(val) + "%")
		}
	}

	rows, err := s.db.Query(query, searchVal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicines := make([]types.MedicineListsReturnPayload, 0)

	for rows.Next() {
		medicine, err := scanRowIntoMedicineLists(rows)

		if err != nil {
			return nil, err
		}

		medicines = append(medicines, *medicine)
	}

	return medicines, nil
}

func (s *Store) CreateMedicine(med types.Medicine, userId int) error {
	values := "?"
	for i := 0; i < 18; i++ {
		values += ", ?"
	}

	query := `INSERT INTO medicine (
		barcode, name, qty, first_unit_id, first_subtotal, first_discount, first_price, 
		second_unit_id, second_unit_to_first_unit_ratio, second_subtotal, second_discount, second_price, 
		third_unit_id, third_unit_to_first_unit_ratio, third_subtotal, third_discount, third_price, description, 
		last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		med.Barcode, med.Name, med.Qty,
		med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
		med.SecondUnitID, med.SecondUnitToFirstUnitRatio, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
		med.ThirdUnitID, med.ThirdUnitToFirstUnitRatio, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
		med.Description, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllMedicines() ([]types.MedicineListsReturnPayload, error) {
	query := `SELECT med.id, med.barcode, med.name, med.qty, 
					uot.name AS unit_one, 
					med.first_discount, med.first_price, 
					utt.name AS unit_two, 
					med.second_unit_to_first_unit_ratio, 
					med.second_discount, med.second_price, 
					utht.name AS unit_three, 
					med.third_unit_to_first_unit_ratio, 
					med.third_discount, med.third_price, 
					med.description, med.created_at, 
					med.last_modified, user.name 
					FROM medicine AS med 
					JOIN unit AS uot ON med.first_unit_id = uot.id 
					JOIN unit AS utt ON med.second_unit_id = utt.id 
					JOIN unit AS utht ON med.third_unit_id = utht.id 
					JOIN user ON user.id = med.last_modified_by_user_id 
					WHERE med.deleted_at IS NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicines := make([]types.MedicineListsReturnPayload, 0)

	for rows.Next() {
		medicine, err := scanRowIntoMedicineLists(rows)

		if err != nil {
			return nil, err
		}

		medicines = append(medicines, *medicine)
	}

	return medicines, nil
}

func (s *Store) DeleteMedicine(med *types.Medicine, user *types.User) error {
	query := "UPDATE medicine SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, med.ID)
	if err != nil {
		return err
	}

	query = `SELECT COUNT(*) FROM main_doctor_presc_medicine_item WHERE medicine_id = ?`
	row := s.db.QueryRow(query, med.ID)
	if row.Err() != nil {
		return row.Err()
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		query = `DELETE FROM main_doctor_presc_medicine_item WHERE medicine_id = ?`
		_, err = s.db.Exec(query, med.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ModifyMedicine(mid int, med types.Medicine, user *types.User) error {
	data, err := s.GetMedicineByID(mid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "medicine", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE medicine SET 
		barcode = ?, name = ?, qty = ?, 
		first_unit_id = ?, first_subtotal = ?, first_discount = ?, first_price = ?, 
		second_unit_id = ?, second_unit_to_first_unit_ratio = ?, second_subtotal = ?, second_discount = ?, second_price = ?, 
		third_unit_id = ?, third_unit_to_first_unit_ratio = ?, third_subtotal = ?, third_discount = ?, third_price = ?, description = ?, 
		last_modified = ?, last_modified_by_user_id = ?
	WHERE id = ?`

	_, err = s.db.Exec(query,
		med.Barcode, med.Name, med.Qty,
		med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
		med.SecondUnitID, med.SecondUnitToFirstUnitRatio, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
		med.ThirdUnitID, med.ThirdUnitToFirstUnitRatio, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
		med.Description, time.Now(), user.ID, mid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateMedicineStock(mid int, newStock float64, user *types.User) error {
	data, err := s.GetMedicineByID(mid)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "medicine", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE medicine SET 
		qty = ?, last_modified = ?, last_modified_by_user_id = ?
	WHERE id = ?`

	_, err = s.db.Exec(query, newStock, time.Now(), user.ID, mid)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoMedicine(rows *sql.Rows) (*types.Medicine, error) {
	medicine := new(types.Medicine)

	err := rows.Scan(
		&medicine.ID,
		&medicine.Barcode,
		&medicine.Name,
		&medicine.Qty,
		&medicine.FirstUnitID,
		&medicine.FirstSubtotal,
		&medicine.FirstDiscount,
		&medicine.FirstPrice,
		&medicine.SecondUnitID,
		&medicine.SecondUnitToFirstUnitRatio,
		&medicine.SecondSubtotal,
		&medicine.SecondDiscount,
		&medicine.SecondPrice,
		&medicine.ThirdUnitID,
		&medicine.ThirdUnitToFirstUnitRatio,
		&medicine.ThirdSubtotal,
		&medicine.ThirdDiscount,
		&medicine.ThirdPrice,
		&medicine.Description,
		&medicine.CreatedAt,
		&medicine.LastModified,
		&medicine.LastModifiedByUserID,
		&medicine.DeletedAt,
		&medicine.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	medicine.CreatedAt = medicine.CreatedAt.Local()
	medicine.LastModified = medicine.LastModified.Local()

	return medicine, nil
}

func scanRowIntoMedicineLists(rows *sql.Rows) (*types.MedicineListsReturnPayload, error) {
	medicine := new(types.MedicineListsReturnPayload)

	err := rows.Scan(
		&medicine.ID,
		&medicine.Barcode,
		&medicine.Name,
		&medicine.Qty,
		&medicine.FirstUnitName,
		&medicine.FirstDiscount,
		&medicine.FirstPrice,
		&medicine.SecondUnitName,
		&medicine.SecondUnitToFirstUnitRatio,
		&medicine.SecondDiscount,
		&medicine.SecondPrice,
		&medicine.ThirdUnitName,
		&medicine.ThirdUnitToFirstUnitRatio,
		&medicine.ThirdDiscount,
		&medicine.ThirdPrice,
		&medicine.Description,
		&medicine.CreatedAt,
		&medicine.LastModified,
		&medicine.LastModifiedByUserName,
	)

	if err != nil {
		return nil, err
	}

	medicine.CreatedAt = medicine.CreatedAt.Local()
	medicine.LastModified = medicine.LastModified.Local()

	return medicine, nil
}
