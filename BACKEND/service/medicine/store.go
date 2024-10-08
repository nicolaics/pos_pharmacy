package medicine

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

func (s *Store) GetMedicineByID(id int) (*types.Medicine, error) {
	query := "SELECT * FROM medicine WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
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

func (s *Store) CreateMedicine(med types.Medicine, userId int) error {
	values := "?"
	for i := 0; i < 16; i++ {
		values += ", ?"
	}

	query := `INSERT INTO medicine (
		barcode, name, qty, first_unit_id, first_subtotal, first_discount, first_price, 
		second_unit_id, second_subtotal, second_discount, second_price, 
		third_unit_id, third_subtotal, third_discount, third_price, description, 
		last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		med.Barcode, med.Name, med.Qty,
		med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
		med.SecondUnitID, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
		med.ThirdUnitID, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
		med.Description, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllMedicines() ([]types.Medicine, error) {
	rows, err := s.db.Query("SELECT * FROM medicine WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	medicines := make([]types.Medicine, 0)

	for rows.Next() {
		medicine, err := scanRowIntoMedicine(rows)

		if err != nil {
			return nil, err
		}

		medicines = append(medicines, *medicine)
	}

	return medicines, nil
}

func (s *Store) DeleteMedicine(med *types.Medicine, userId int) error {
	query := "UPDATE medicine SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), userId, med.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyMedicine(mid int, med types.Medicine, userId int) error {
	query := `UPDATE medicine SET (
		barcode = ?, name = ?, qty = ?, 
		first_unit_id = ?, first_subtotal = ?, first_discount = ?, first_price = ?, 
		second_unit_id = ?, second_subtotal = ?, second_discount = ?, second_price = ?, 
		third_unit_id = ?, third_subtotal = ?, third_discount = ?, third_price = ?, description = ?, 
		last_modified = ?, last_modified_by_user_id = ?
	) WHERE id = ?`

	_, err := s.db.Exec(query,
		med.Barcode, med.Name, med.Qty,
		med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
		med.SecondUnitID, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
		med.ThirdUnitID, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
		med.Description, time.Now(), userId, mid)
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
		&medicine.SecondSubtotal,
		&medicine.SecondDiscount,
		&medicine.SecondPrice,
		&medicine.ThirdUnitID,
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
