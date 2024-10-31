package maindoctorprescmeditem

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

func (s *Store) CreateMainDoctorPrescMedItem(item types.MainDoctorPrescMedItem) error {
	query := `INSERT INTO main_doctor_presc_medicine_item (
				medicine_id, medicine_content_id, qty, unit_id, user_id, last_modified_by_user_id) 
				VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, item.MedicineID, item.MedicineContentID, item.Qty, item.UnitID, item.UserID, item.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetMainDoctorPrescMedItemByMedicineData(medId int) (*types.MainDoctorPrescMedItemReturn, error) {
	query := `SELECT medicine.name, md.qty, unit.name 
				FROM main_doctor_presc_medicine_item AS md 
				JOIN medicine ON medicine.id = md.medicine_content_id 
				JOIN unit ON unit.id = md.unit_id 
				WHERE md.medicine_id = ?`
	rows, err := s.db.Query(query, medId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medContents := make([]types.MainDoctorPrescMedContent, 0)

	for rows.Next() {
		medContent, err := scanRowIntoMedContents(rows)
		if err != nil {
			return nil, err
		}

		medContents = append(medContents, *medContent)
	}

	query = `SELECT medicine.id, medicine.name, md.last_modified, user.name 
				FROM main_doctor_presc_medicine_item AS md 
				JOIN medicine ON medicine.id = md.medicine_id 
				JOIN user ON user.id = md.last_modified_by_user_id 
				WHERE md.medicine_id = ?`
	rows, err = s.db.Query(query, medId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var temp struct {
		ID int
		Name string
		LastModified time.Time
		LastModifiedByUser string
	}

	for rows.Next() {
		err = rows.Scan(
			&temp.ID,
			&temp.Name,
			&temp.LastModified,
			&temp.LastModifiedByUser,
		)
		if err != nil {
			return nil, err
		}
	}

	if temp.ID == 0 {
		return nil, fmt.Errorf("medicine id %d not found", medId)
	}

	returnPayload := types.MainDoctorPrescMedItemReturn{
		MedicineName: temp.Name,
		MedicineContents: medContents,
		LastModified: temp.LastModified.Local(),
		LastModifiedByUserName: temp.LastModifiedByUser,
	}
	return &returnPayload, nil
}

func (s *Store) GetAllMainDoctorPrescMedItemByMedicineData() ([]types.MainDoctorPrescMedItemReturn, error) {
	query := `SELECT DISTINCT medicine_id FROM main_doctor_presc_medicine_item`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medIds := make([]int, 0)
	for rows.Next() {
		var medId int
		err = rows.Scan(&medId)
		if err != nil {
			return nil, err
		}

		medIds = append(medIds, medId)
	}

	returnPayload := make([]types.MainDoctorPrescMedItemReturn, 0)

	for _, medId := range(medIds) {
		query = `SELECT medicine.name, md.qty, unit.name 
				FROM main_doctor_presc_medicine_item AS md 
				JOIN medicine ON medicine.id = md.medicine_content_id 
				JOIN unit ON unit.id = md.unit_id 
				WHERE md.medicine_id = ?`
		rows, err = s.db.Query(query, medId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		medContents := make([]types.MainDoctorPrescMedContent, 0)

		for rows.Next() {
			medContent, err := scanRowIntoMedContents(rows)
			if err != nil {
				return nil, err
			}

			medContents = append(medContents, *medContent)
		}

		query = `SELECT medicine.id, medicine.name, md.last_modified, user.name 
				FROM main_doctor_presc_medicine_item AS md 
				JOIN medicine ON medicine.id = md.medicine_id 
				JOIN user ON user.id = md.last_modified_by_user_id 
				WHERE md.medicine_id = ?`
		rows, err = s.db.Query(query, medId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var temp struct {
			ID int
			Name string
			LastModified time.Time
			LastModifiedByUser string
		}
	
		for rows.Next() {
			err = rows.Scan(
				&temp.ID,
				&temp.Name,
				&temp.LastModified,
				&temp.LastModifiedByUser,
			)
			if err != nil {
				return nil, err
			}
		}
	
		if temp.ID == 0 {
			return nil, fmt.Errorf("medicine id %d not found", medId)
		}

		returnPayload = append(returnPayload, types.MainDoctorPrescMedItemReturn{
			MedicineName: temp.Name,
			MedicineContents: medContents,
			LastModified: temp.LastModified.Local(),
			LastModifiedByUserName: temp.LastModifiedByUser,
		})
	}

	
	return returnPayload, nil
}

func (s *Store) IsMedicineContentsExist(medId int) (bool, error) {
	query := `SELECT COUNT(*) 
				FROM main_doctor_presc_medicine_item AS md 
				JOIN medicine ON medicine.id = md.medicine_id 
				WHERE md.medicine_id = ?`
	row := s.db.QueryRow(query, medId)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return (count < 1), nil
}

func (s *Store) IsMedicineBarcodeExist(barcode string) (bool, error) {
	query := `SELECT COUNT(*) 
				FROM medicine 
				WHERE medicine.barcode = ? AND medicine.deleted_at IS NULL`
	row := s.db.QueryRow(query, barcode)
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

func (s *Store) DeleteMainDoctorPrescMedItem(medId int, user *types.User) error {
	data, err := s.GetMainDoctorPrescMedItemByMedicineData(medId)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("delete", "main-doctor-prescription-medicines", user.Name, medId, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `DELETE FROM main_doctor_presc_medicine_item WHERE medicine_id = ?`
	_, err = s.db.Exec(query, medId)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoMedContents(rows *sql.Rows) (*types.MainDoctorPrescMedContent, error) {
	medContent := new(types.MainDoctorPrescMedContent)

	err := rows.Scan(
		&medContent.Name,
		&medContent.Qty,
		&medContent.Unit,
	)

	if err != nil {
		return nil, err
	}

	return medContent, nil
}
