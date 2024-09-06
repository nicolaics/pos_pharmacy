package prescription

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

func (s *Store) GetPrescriptionsByNumber(number int) ([]types.Prescription, error) {
	query := "SELECT * FROM prescription WHERE number = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.Prescription, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescription(rows)
		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionByID(id int) (*types.Prescription, error) {
	query := "SELECT * FROM prescription WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescription := new(types.Prescription)

	for rows.Next() {
		prescription, err = scanRowIntoPrescription(rows)

		if err != nil {
			return nil, err
		}
	}

	if prescription.ID == 0 {
		return nil, fmt.Errorf("prescription not found")
	}

	return prescription, nil
}

func (s *Store) GetPrescriptionsByDate(startDate time.Time, endDate time.Time) ([]types.Prescription, error) {
	query := `SELECT * FROM prescription 
				WHERE (prescription_date BETWEEN DATE(?) AND DATE(?)) 
				AND deleted_at IS NULL 
				ORDER BY prescription_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.Prescription, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescription(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionID(invoiceId int, number int, date time.Time, patientName string, totalPrice float64) (int, error) {
	query := `SELECT id FROM prescription 
				WHERE invoice_id = ? AND number = ? AND prescription_date = ? 
				AND patient_name = ? AND total_price = ? 
				AND deleted_at IS NULL`

	rows, err := s.db.Query(query, invoiceId, number, date, patientName, totalPrice)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var prescriptionId int

	for rows.Next() {
		err = rows.Scan(&prescriptionId)
		if err != nil {
			return -1, err
		}
	}

	if prescriptionId == 0 {
		return -1, fmt.Errorf("prescription not found")
	}

	return prescriptionId, nil
}

func (s *Store) CreatePrescription(prescription types.Prescription) error {
	values := "?"
	for i := 0; i < 10; i++ {
		values += ", ?"
	}

	query := `INSERT INTO prescription (
		invoice_id, number, prescription_date, patient_id, doctor_id, qty, 
		price, total_price, description, 
		user_id, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		prescription.InvoiceID, prescription.Number, prescription.PrescriptionDate,
		prescription.PatientID, prescription.DoctorID, prescription.Qty,
		prescription.Price, prescription.TotalPrice, prescription.Description,
		prescription.UserID, prescription.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePrescriptionMedicineItems(prescMedItems types.PrescriptionMedicineItems) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO prescription_medicine_items (
				prescription_id, medicine_id, qty, unit_id, 
				price, discount, subtotal
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query, 
						prescMedItems.PrescriptionID, prescMedItems.MedicineID,
						prescMedItems.Qty, prescMedItems.UnitID, prescMedItems.Price,
						prescMedItems.Discount, prescMedItems.Subtotal)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPrescriptionMedicineItems(prescriptionId int) ([]types.PrescriptionMedicineItemRow, error) {
	query := `SELECT 
			pmi.id, 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.price, pmi.discount, pmi.subtotal 
			
			FROM prescription_medicine_items as pmi 
			JOIN prescription as presc ON pmi.prescription_id = presc.id 
			JOIN medicine ON pmi.medicine_id = medicine.id 
			JOIN unit ON pmi.unit_id = unit.id 
			WHERE presc.id = ? AND presc.deleted_at IS NULL`

	rows, err := s.db.Query(query, prescriptionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescMedItems := make([]types.PrescriptionMedicineItemRow, 0)

	for rows.Next() {
		prescMedItem, err := scanRowIntoPrescriptionMedicineItems(rows)

		if err != nil {
			return nil, err
		}

		prescMedItems = append(prescMedItems, *prescMedItem)
	}

	return prescMedItems, nil
}

func (s *Store) DeletePrescription(prescription *types.Prescription, userId int) error {
	query := "UPDATE prescription SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, prescription.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeletePrescriptionMedicineItems(prescriptionId int) error {
	_, err := s.db.Exec("DELETE FROM prescription_medicine_items WHERE prescription_id = ? ", prescriptionId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPrescription(id int, prescription types.Prescription) error {
	query := `UPDATE prescription SET 
				number = ?, prescription_date = ?, patient_id = ?, doctor_id = ?, 
				qty = ?, price = ?, total_price = ?, description = ?, 
				last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err := s.db.Exec(query,
						prescription.Number, prescription.PrescriptionDate, 
						prescription.PatientID, prescription.DoctorID,
						prescription.Qty, prescription.Price,
						prescription.TotalPrice, prescription.Description, 
						time.Now(), prescription.LastModifiedByUserID, id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoPrescription(rows *sql.Rows) (*types.Prescription, error) {
	prescription := new(types.Prescription)

	err := rows.Scan(
		&prescription.ID,
		&prescription.InvoiceID,
		&prescription.Number,
		&prescription.PrescriptionDate,
		&prescription.PatientID,
		&prescription.DoctorID,
		&prescription.Qty,
		&prescription.Price,
		&prescription.TotalPrice,
		&prescription.Description,
		&prescription.CreatedAt,
		&prescription.UserID,
		&prescription.LastModified,
		&prescription.LastModifiedByUserID,
		&prescription.DeletedAt,
		&prescription.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	prescription.PrescriptionDate = prescription.PrescriptionDate.Local()
	prescription.CreatedAt = prescription.CreatedAt.Local()
	prescription.LastModified = prescription.LastModified.Local()

	return prescription, nil
}

func scanRowIntoPrescriptionMedicineItems(rows *sql.Rows) (*types.PrescriptionMedicineItemRow, error) {
	prescMedItem := new(types.PrescriptionMedicineItemRow)

	err := rows.Scan(
		&prescMedItem.ID,
		&prescMedItem.MedicineBarcode,
		&prescMedItem.MedicineName,
		&prescMedItem.Qty,
		&prescMedItem.Unit,
		&prescMedItem.Price,
		&prescMedItem.Discount,
		&prescMedItem.Subtotal,
	)

	if err != nil {
		return nil, err
	}

	return prescMedItem, nil
}
