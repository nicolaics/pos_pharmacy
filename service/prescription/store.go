package prescription

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

func (s *Store) GetPrescriptionsByDate(startDate time.Time, endDate time.Time) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice AS i ON p.invoice_id = i.id 
					JOIN customer ON i.customer_id = customer.id 
					JOIN user ON user.id = p.user_id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					ORDER BY p.prescription_date DESC`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionsByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT COUNT(*)
					FROM prescription 
					WHERE (prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND number = ? 
					AND deleted_at IS NULL 
					ORDER BY p.prescription_date DESC`

	row := s.db.QueryRow(query, startDate, endDate, number)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	if count == 0 {
		query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.number LIKE ? 
					ORDER BY p.prescription_date DESC`

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
			prescription, err := scanRowIntoPrescriptionLists(rows)
			if err != nil {
				return nil, err
			}

			prescriptions = append(prescriptions, *prescription)
		}

		return prescriptions, nil
	}

	query = `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.number = ? 
					ORDER BY p.prescription_date DESC`
	
	rows, err := s.db.Query(query, startDate, endDate, number)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionsByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.user_id = ? 
					ORDER BY p.prescription_date DESC`

	rows, err := s.db.Query(query, startDate, endDate, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionsByDateAndPatientID(startDate time.Time, endDate time.Time, pid int) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.patient_id = ? 
					ORDER BY p.prescription_date DESC`
	
	rows, err := s.db.Query(query, startDate, endDate, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionsByDateAndDoctorID(startDate time.Time, endDate time.Time, did int) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.doctor_id = ? 
					ORDER BY p.prescription_date DESC`
	
	rows, err := s.db.Query(query, startDate, endDate, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionsByDateAndInvoiceID(startDate time.Time, endDate time.Time, iid int) ([]types.PrescriptionListsReturnPayload, error) {
	query := `SELECT p.id, p.number, p.prescription_date, 
					patient.name, doctor.name, 
					p.qty, p.price, p.total_price, p.description, 
					user.name, 
					i.number, 
					customer.name, 
					i.total_price, i.invoice_date 
					FROM prescription AS p 
					JOIN patient ON p.patient_id = patient.id 
					JOIN doctor ON p.doctor_id = doctor.id 
					JOIN invoice ON p.invoice_id = invoice.id 
					JOIN customer ON invoice.customer_id = customer.id 
					WHERE (p.prescription_date BETWEEN DATE(?) AND DATE(?)) 
					AND p.deleted_at IS NULL 
					AND p.invoice_id = ? 
					ORDER BY p.prescription_date DESC`
	
	rows, err := s.db.Query(query, startDate, endDate, iid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescriptions := make([]types.PrescriptionListsReturnPayload, 0)

	for rows.Next() {
		prescription, err := scanRowIntoPrescriptionLists(rows)

		if err != nil {
			return nil, err
		}

		prescriptions = append(prescriptions, *prescription)
	}

	return prescriptions, nil
}

func (s *Store) GetPrescriptionID(invoiceId int, number int, date time.Time, patientId int, totalPrice float64, doctorId int) (int, error) {
	query := `SELECT id FROM prescription 
				WHERE invoice_id = ? AND number = ? AND prescription_date = ? 
				AND patient_id = ? AND total_price = ? AND doctor_id = ? 
				AND deleted_at IS NULL`

	rows, err := s.db.Query(query, invoiceId, number, date, patientId, totalPrice, doctorId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var prescriptionId int

	for rows.Next() {
		err = rows.Scan(&prescriptionId)
		if err != nil {
			return 0, err
		}
	}

	if prescriptionId == 0 {
		return 0, fmt.Errorf("prescription not found")
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

func (s *Store) DeletePrescription(prescription *types.Prescription, user *types.User) error {
	query := "UPDATE prescription SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, prescription.ID)
	if err != nil {
		return err
	}

	data, err := s.GetPrescriptionByID(prescription.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "prescription", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) DeletePrescriptionMedicineItems(prescription *types.Prescription, user *types.User) error {
	data, err := s.GetPrescriptionMedicineItems(prescription.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"prescription": prescription,
		"deleted_medicine_items": data,
	}

	err = logger.WriteLog("delete", "prescription", user.Name, prescription.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM prescription_medicine_items WHERE prescription_id = ? ", prescription.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyPrescription(id int, prescription types.Prescription, user *types.User) error {
	data, err := s.GetPrescriptionByID(prescription.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"previous_data": data,
	}

	err = logger.WriteLog("modify", "prescription", user.Name, data.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	query := `UPDATE prescription SET 
				number = ?, prescription_date = ?, patient_id = ?, doctor_id = ?, 
				qty = ?, price = ?, total_price = ?, description = ?, 
				last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`

	_, err = s.db.Exec(query,
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

	return prescription, nil
}

func scanRowIntoPrescriptionLists(rows *sql.Rows) (*types.PrescriptionListsReturnPayload, error) {
	prescription := new(types.PrescriptionListsReturnPayload)

	err := rows.Scan(
		&prescription.ID,
		&prescription.Number,
		&prescription.PrescriptionDate,
		&prescription.PatientName,
		&prescription.DoctorName,
		&prescription.Qty,
		&prescription.Price,
		&prescription.TotalPrice,
		&prescription.Description,
		&prescription.UserName,
		&prescription.Invoice.Number,
		&prescription.Invoice.CustomerName,
		&prescription.Invoice.TotalPrice,
		&prescription.Invoice.InvoiceDate,
	)

	if err != nil {
		return nil, err
	}

	prescription.PrescriptionDate = prescription.PrescriptionDate.Local()
	prescription.Invoice.InvoiceDate = prescription.Invoice.InvoiceDate.Local()

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
