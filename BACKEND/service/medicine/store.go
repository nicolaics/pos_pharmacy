package medicine

import (
	"database/sql"
	"fmt"

	"github.com/nicolaics/pos_pharmacy/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetMedicineByName(name string) (*types.Medicine, error) {
	rows, err := s.db.Query("SELECT * FROM medicine WHERE name = ? ", name)
	if err != nil {
		return nil, err
	}

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
	rows, err := s.db.Query("SELECT * FROM medicine WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

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
	rows, err := s.db.Query("SELECT * FROM medicine WHERE barcode = ?", barcode)
	if err != nil {
		return nil, err
	}

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

func (s *Store) CreateMedicine(med types.Medicine) error {
	fields := "barcode, name, qty, first_unit_id, first_subtotal, first_discount, first_price, "
    fields += "second_unit_id, second_subtotal, second_discount, second_price, "
	fields += "third_unit_id, third_subtotal, third_discount, third_price, description"

	values := "?"
	for i := 0; i < 15; i++ {
		values += ", ?"
	}

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO medicine (%s) VALUES (%s)", fields, values),
						med.Barcode, med.Name, med.Qty,
						med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
						med.SecondUnitID, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
						med.ThirdUnitID, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
						med.Description)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllMedicines() ([]types.Medicine, error) {
	rows, err := s.db.Query("SELECT * FROM medicine")

	if err != nil {
		return nil, err
	}

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

func (s *Store) DeleteMedicine(med *types.Medicine) error {
	_, err := s.db.Exec("DELETE FROM medicine WHERE id = ?", med.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyMedicine(id int, med types.Medicine) error {
	fields := "barcode = ?, name = ?, qty = ?, "
	fields += "first_unit_id = ?, first_subtotal = ?, first_discount = ?, first_price = ?, "
    fields += "second_unit_id = ?, second_subtotal = ?, second_discount = ?, second_price = ?, "
	fields += "third_unit_id = ?, third_subtotal = ?, third_discount = ?, third_price = ?, description = ?"

	_, err := s.db.Exec(fmt.Sprintf("UPDATE SET (%s) WHERE id = ?", fields),
						med.Barcode, med.Name, med.Qty,
						med.FirstUnitID, med.FirstSubtotal, med.FirstDiscount, med.FirstPrice,
						med.SecondUnitID, med.SecondSubtotal, med.SecondDiscount, med.SecondPrice,
						med.ThirdUnitID, med.ThirdSubtotal, med.ThirdDiscount, med.ThirdPrice,
						med.Description, id)
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
	)

	if err != nil {
		return nil, err
	}

	medicine.CreatedAt = medicine.CreatedAt.Local()

	return medicine, nil
}
