package prescription

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/types"

	dectofrac "github.com/av-elier/go-decimal-to-rational"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetPrescriptionsByNumber(number int) ([]types.Prescription, error) {
	query := "SELECT * FROM prescription WHERE number = ? AND deleted_at IS NULL ORDER BY prescription_date DESC"
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
	query := "SELECT * FROM prescription WHERE id = ? AND deleted_at IS NULL ORDER BY prescription_date DESC"
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE prescription_date >= ? AND prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
					WHERE p.prescription_date >= ? AND p.prescription_date < ? 
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
				AND deleted_at IS NULL 
				ORDER BY prescription_date DESC`

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
	for i := 0; i < 11; i++ {
		values += ", ?"
	}

	query := `INSERT INTO prescription (
		invoice_id, number, prescription_date, patient_id, doctor_id, qty, 
		price, total_price, description, 
		user_id, last_modified_by_user_id, pdf_url
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		prescription.InvoiceID, prescription.Number, prescription.PrescriptionDate,
		prescription.PatientID, prescription.DoctorID, prescription.Qty,
		prescription.Price, prescription.TotalPrice, prescription.Description,
		prescription.UserID, prescription.LastModifiedByUserID, prescription.PDFUrl)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateSetItem(medicineSet types.PrescriptionSetItem) error {
	values := "?"
	for i := 0; i < 9; i++ {
		values += ", ?"
	}

	query := `INSERT INTO prescription_set_item 
				(prescription_id, mf_id, dose_id, set_unit_id, 
				consume_time_id, det_id, prescription_set_usage_id, must_finish, 
				print_eticket) 
				VALUES (` + values + `)`
	_, err := s.db.Exec(query, medicineSet.PrescriptionID, medicineSet.MfID, medicineSet.DoseID,
		medicineSet.SetUnitID, medicineSet.ConsumeTimeID, medicineSet.DetID,
		medicineSet.UsageID, medicineSet.MustFinish, medicineSet.PrintEticket)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreatePrescriptionMedicineItem(prescMedItem types.PrescriptionMedicineItem) error {
	values := "?"
	for i := 0; i < 7; i++ {
		values += ", ?"
	}

	query := `INSERT INTO prescription_medicine_item (
				prescription_set_item_id, medicine_id, qty, unit_id, 
				price, discount_percentage, discount_amount, subtotal
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		prescMedItem.PrescriptionSetItemID, prescMedItem.MedicineID,
		prescMedItem.Qty, prescMedItem.UnitID, prescMedItem.Price,
		prescMedItem.DiscountPercentage, prescMedItem.DiscountAmount, prescMedItem.Subtotal)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPrescriptionMedicineItems(prescriptionSetItemId int) ([]types.PrescriptionMedicineItemReturn, error) {
	query := `SELECT 
			medicine.barcode, medicine.name, 
			pmi.qty, 
			unit.name, 
			pmi.price, pmi.discount_percentage, pmi.discount_amount, pmi.subtotal 
			
			FROM prescription_medicine_item as pmi 
			JOIN prescription_set_item as psi ON pmi.prescription_set_item_id = psi.id 
			JOIN medicine ON pmi.medicine_id = medicine.id 
			JOIN unit ON pmi.unit_id = unit.id 
			WHERE psi.id = ?`

	rows, err := s.db.Query(query, prescriptionSetItemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prescMedItems := make([]types.PrescriptionMedicineItemReturn, 0)

	for rows.Next() {
		prescMedItem, err := scanRowIntoPrescriptionMedicineItem(rows)
		if err != nil {
			return nil, err
		}

		var qty string

		if prescMedItem.Qty < 1.0 {
			qty = dectofrac.NewRatP(prescMedItem.Qty, 0.01).String()
		} else {
			if prescMedItem.Qty == math.Trunc(prescMedItem.Qty) {
				qty = fmt.Sprintf("%.0f", prescMedItem.Qty)
			} else {
				qty = fmt.Sprintf("%.1f", prescMedItem.Qty)
			}
		}

		prescMedItems = append(prescMedItems, types.PrescriptionMedicineItemReturn{
			MedicineBarcode:    prescMedItem.MedicineBarcode,
			MedicineName:       prescMedItem.MedicineName,
			QtyString:          qty,
			QtyFloat:           prescMedItem.Qty,
			Unit:               prescMedItem.Unit,
			Price:              prescMedItem.Price,
			DiscountPercentage: prescMedItem.DiscountPercentage,
			DiscountAmount:     prescMedItem.DiscountAmount,
			Subtotal:           prescMedItem.Subtotal,
		})
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

func (s *Store) DeletePrescriptionMedicineItem(prescription *types.Prescription, setItemId int, user *types.User) error {
	data, err := s.GetPrescriptionMedicineItems(setItemId)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"prescription":          prescription,
		"deleted_medicine_item": data,
	}

	err = logger.WriteLog("delete", "prescription", user.Name, prescription.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM prescription_medicine_item WHERE prescription_set_item_id = ? ", setItemId)
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

func (s *Store) AbsoluteDeletePrescription(presc types.Prescription) error {
	query := `SELECT id FROM prescription 
				WHERE number = ? AND prescription_date = ? 
				AND patient_id = ? AND doctor_id = ? 
				AND qty = ? AND price = ? AND total_price = ? 
				AND description = ? AND invoice_id = ?`

	rows, err := s.db.Query(query, presc.Number, presc.PrescriptionDate, presc.PatientID, presc.DoctorID,
		presc.Qty, presc.Price, presc.TotalPrice, presc.Description, presc.InvoiceID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var prescId int

	for rows.Next() {
		err = rows.Scan(&prescId)
		if err != nil {
			return nil
		}
	}

	if prescId == 0 {
		return nil
	}

	query = `SELECT id FROM prescription_set_item WHERE prescription_id = ?`
	rows, err = s.db.Query(query, prescId)
	if err != nil {
		return err
	}
	defer rows.Close()

	setItemsId := make([]int, 0)

	for rows.Next() {
		var setItemTemp int
		err = rows.Scan(&setItemTemp)
		if err != nil {
			return nil
		}

		setItemsId = append(setItemsId, setItemTemp)
	}

	if prescId == 0 {
		return nil
	}

	err = s.DeleteEticketByPrescriptionID(prescId)
	if err != nil {
		return err
	}

	for _, setItemId := range setItemsId {
		query = "DELETE FROM prescription_medicine_item WHERE prescription_set_item_id = ?"
		_, err = s.db.Exec(query, setItemId)
		if err != nil {
			return err
		}
	}

	query = "DELETE FROM prescription_set_item WHERE prescription_id = ?"
	_, err = s.db.Exec(query, prescId)
	if err != nil {
		return err
	}

	query = `DELETE FROM prescription WHERE id = ?`
	_, err = s.db.Exec(query, prescId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPrescriptionMedicineItemID(prescMedItem types.PrescriptionMedicineItem) (int, error) {
	query := `SELECT id FROM prescription_medicine_item 
				WHERE prescription_set_item_id = ? AND medicine_id = ? 
				AND qty = ? AND unit_id = ? 
				AND price = ? AND discount_percentage = ? AND discount_amount = ? 
				AND subtotal = ?`

	rows, err := s.db.Query(query,
		prescMedItem.PrescriptionSetItemID, prescMedItem.MedicineID,
		prescMedItem.Qty, prescMedItem.UnitID, prescMedItem.Price,
		prescMedItem.DiscountPercentage, prescMedItem.DiscountAmount, prescMedItem.Subtotal)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	if id == 0 {
		return 0, fmt.Errorf("medicine item not found")
	}

	return id, nil
}

func (s *Store) GetSetItemByID(id int) (*types.PrescriptionSetItem, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_item WHERE id = ? ", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicineSet := new(types.PrescriptionSetItem)

	for rows.Next() {
		medicineSet, err = scanRowIntoSetItem(rows)

		if err != nil {
			return nil, err
		}
	}

	if medicineSet.ID == 0 {
		return nil, fmt.Errorf("medicine set not found")
	}

	return medicineSet, nil
}

func (s *Store) GetSetItemsByPrescriptionID(prescId int) ([]types.PrescriptionSetItem, error) {
	rows, err := s.db.Query("SELECT * FROM prescription_set_item WHERE prescription_id = ? ", prescId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicineSets := make([]types.PrescriptionSetItem, 0)

	for rows.Next() {
		medicineSet, err := scanRowIntoSetItem(rows)
		if err != nil {
			return nil, err
		}

		medicineSets = append(medicineSets, *medicineSet)
	}

	return medicineSets, nil
}

func (s *Store) GetSetItemID(medicineSet types.PrescriptionSetItem) (int, error) {
	query := `SELECT id FROM prescription_set_item 
				WHERE prescription_id = ? AND mf_id = ? AND dose_id = ? 
				AND set_unit_id = ? AND consume_time_id = ? AND det_id = ? 
				AND prescription_set_usage_id = ? AND must_finish = ? 
				AND print_eticket = ?`
	rows, err := s.db.Query(query, medicineSet.PrescriptionID, medicineSet.MfID,
		medicineSet.DoseID, medicineSet.SetUnitID,
		medicineSet.ConsumeTimeID, medicineSet.DetID,
		medicineSet.UsageID, medicineSet.MustFinish, medicineSet.PrintEticket)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	if id == 0 {
		return 0, fmt.Errorf("medicine set not found")
	}

	return id, err
}

func (s *Store) CreateEticket(eticket types.Eticket) error {
	query := `INSERT INTO eticket 
				(prescription_id, prescription_set_item_id, number, medicine_qty, size, pdf_url) 
				VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, eticket.PrescriptionID, eticket.PrescriptionSetItemID,
		eticket.Number, eticket.MedicineQty, eticket.Size, eticket.PDFUrl)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteEticket(id int) error {
	query := `DELETE FROM eticket WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteEticketByPrescriptionID(id int) error {
	query := `DELETE FROM eticket WHERE prescription_id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetEticketID(eticket types.Eticket) (int, error) {
	query := `SELECT id FROM eticket 
				WHERE prescription_id = ? AND prescription_set_item_id = ? 
				AND number = ? AND medicine_qty = ?`
	rows, err := s.db.Query(query, eticket.PrescriptionID, eticket.PrescriptionSetItemID,
		eticket.Number, eticket.MedicineQty)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	if id == 0 {
		return 0, fmt.Errorf("eticket not found")
	}

	return id, nil
}

func (s *Store) GetEticketsByPrescriptionID(prescId int) ([]types.Eticket, error) {
	rows, err := s.db.Query("SELECT * FROM eticket WHERE prescription_id = ? ", prescId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	etickets := make([]types.Eticket, 0)

	for rows.Next() {
		eticket, err := scanRowIntoEticket(rows)
		if err != nil {
			return nil, err
		}

		etickets = append(etickets, *eticket)
	}

	return etickets, nil
}

func (s *Store) UpdatePDFUrl(tableName string, prescId int, fileName string) error {
	var query string

	if tableName == "eticket" {
		query = `UPDATE eticket SET pdf_url = ? WHERE id = ?`
	} else if tableName == "prescription" {
		query = `UPDATE prescription SET pdf_url = ? WHERE id = ?`
	} else {
		return fmt.Errorf("unknown table")
	}

	_, err := s.db.Exec(query, fileName, prescId)
	if err != nil {
		return err
	}

	return nil
}

// false means doesn't exist
func (s *Store) IsPDFUrlExist(tableName string, fileName string) (bool, error) {
	var query string

	if tableName == "eticket" {
		query = `SELECT COUNT(*) FROM eticket WHERE pdf_url = ?`
	} else if tableName == "prescription" {
		query = `SELECT COUNT(*) FROM prescription WHERE pdf_url = ?`
	} else {
		return true, fmt.Errorf("unknown table")
	}

	row := s.db.QueryRow(query, fileName)
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

func (s *Store) UpdateEticketID(eticketId int, prescSetItemId int) error {
	query := `UPDATE prescription_set_item SET eticket_id = ? WHERE id = ?`
	_, err := s.db.Exec(query, eticketId, prescSetItemId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetPrescriptionSetAndMedicineItems(prescriptionId int) ([]types.PrescriptionSetItemReturn, error) {
	query := `SELECT psi.id, 
					mf.name, dose.name, unit.name AS set_unit, 
					consume_time.name, det.name, psu.name, 
					psi.must_finish, psi.print_eticket, psi.eticket_id 
					FROM prescription_set_item AS psi 
					JOIN mf ON psi.mf_id = mf.id 
					JOIN dose ON dose.id = psi.dose_id 
					JOIN unit ON unit.id = psi.set_unit_id 
					JOIN consume_time ON consume_time.id = psi.consume_time_id 
					JOIN det ON det.id = psi.det_id 
					JOIN prescription_set_usage AS psu ON psu.id = psi.prescription_set_usage_id 
					JOIN prescription ON prescription.id = psi.prescription_id 
					WHERE psi.prescription_id = ? AND prescription.deleted_at IS NULL`

	rows, err := s.db.Query(query, prescriptionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type SetItemTemp struct {
		ID           int
		Mf           string
		Dose         string
		SetUnit      string
		ConsumeTime  string
		Det          string
		Usage        string
		MustFinish   bool
		PrintEticket bool
		EticketID    int
	}
	setItemsTemp := make([]SetItemTemp, 0)

	for rows.Next() {
		var setItem SetItemTemp
		err = rows.Scan(&setItem.ID,
			&setItem.Mf,
			&setItem.Dose,
			&setItem.SetUnit,
			&setItem.ConsumeTime,
			&setItem.Det,
			&setItem.Usage,
			&setItem.MustFinish,
			&setItem.PrintEticket,
			&setItem.EticketID)
		if err != nil {
			return nil, err
		}

		setItemsTemp = append(setItemsTemp, setItem)
	}

	setItems := make([]types.PrescriptionSetItemReturn, 0)

	for _, setItem := range setItemsTemp {
		medItems, err := s.GetPrescriptionMedicineItems(setItem.ID)
		if err != nil {
			return nil, err
		}

		setItems = append(setItems, types.PrescriptionSetItemReturn{
			ID:            setItem.ID,
			Mf:            setItem.Mf,
			Dose:          setItem.Dose,
			SetUnit:       setItem.SetUnit,
			ConsumeTime:   setItem.ConsumeTime,
			Det:           setItem.Det,
			Usage:         setItem.Usage,
			MustFinish:    setItem.MustFinish,
			PrintEticket:  setItem.PrintEticket,
			MedicineItems: medItems,
		})
	}

	return setItems, nil
}

func (s *Store) DeleteSetItem(prescription *types.Prescription, user *types.User) error {
	data, err := s.GetSetItemsByPrescriptionID(prescription.ID)
	if err != nil {
		return err
	}

	writeData := map[string]interface{}{
		"prescription":         prescription,
		"deleted_medicine_set": data,
	}

	err = logger.WriteLog("delete", "prescription", user.Name, prescription.ID, writeData)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	_, err = s.db.Exec("DELETE FROM prescription_set_item WHERE prescription_id = ? ", prescription.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsValidPrescriptionNumber(number int, startDate time.Time, endDate time.Time) (bool, error) {
	query := `SELECT COUNT(*)
					FROM prescription 
					WHERE prescription_date >= ? AND prescription_date < ? 
					AND number = ? 
					AND deleted_at IS NULL`

	row := s.db.QueryRow(query, startDate, endDate, number)
	if row.Err() != nil {
		return false, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func scanRowIntoSetItem(rows *sql.Rows) (*types.PrescriptionSetItem, error) {
	medicineSet := new(types.PrescriptionSetItem)

	err := rows.Scan(
		&medicineSet.ID,
		&medicineSet.PrescriptionID,
		&medicineSet.MfID,
		&medicineSet.DoseID,
		&medicineSet.SetUnitID,
		&medicineSet.ConsumeTimeID,
		&medicineSet.DetID,
		medicineSet.UsageID,
		&medicineSet.MustFinish,
		&medicineSet.PrintEticket,
	)

	if err != nil {
		return nil, err
	}

	return medicineSet, nil
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
		&prescription.PDFUrl,
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

func scanRowIntoPrescriptionMedicineItem(rows *sql.Rows) (*types.PrescriptionMedicineItemTemp, error) {
	prescMedItem := new(types.PrescriptionMedicineItemTemp)

	err := rows.Scan(
		&prescMedItem.MedicineBarcode,
		&prescMedItem.MedicineName,
		&prescMedItem.Qty,
		&prescMedItem.Unit,
		&prescMedItem.Price,
		&prescMedItem.DiscountPercentage,
		&prescMedItem.DiscountAmount,
		&prescMedItem.Subtotal,
	)

	if err != nil {
		return nil, err
	}

	return prescMedItem, nil
}

func scanRowIntoEticket(rows *sql.Rows) (*types.Eticket, error) {
	eticket := new(types.Eticket)

	err := rows.Scan(
		&eticket.ID,
		&eticket.PrescriptionID,
		&eticket.PrescriptionSetItemID,
		&eticket.Number,
		&eticket.MedicineQty,
		&eticket.Size,
		&eticket.PDFUrl,
		&eticket.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	eticket.CreatedAt = eticket.CreatedAt.Local()

	return eticket, nil
}
