package supplier

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

func (s *Store) GetSupplierByName(name string) (*types.Supplier, error) {
	query := "SELECT * FROM supplier WHERE name = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	supplier := new(types.Supplier)

	for rows.Next() {
		supplier, err = scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}
	}

	if supplier.ID == 0 {
		return nil, fmt.Errorf("supplier not found")
	}

	return supplier, nil
}

func (s *Store) GetSupplierBySearchName(name string) ([]types.Supplier, error) {
	query := "SELECT COUNT(*) FROM supplier WHERE name = ? AND deleted_at IS NULL"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	suppliers := make([]types.Supplier, 0)

	if count == 0 {
		query = "SELECT * FROM supplier WHERE name LIKE ? AND deleted_at IS NULL"
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
			supplier, err := scanRowIntoSupplier(rows)

			if err != nil {
				return nil, err
			}

			suppliers = append(suppliers, *supplier)
		}

		return suppliers, nil
	}

	query = "SELECT * FROM supplier WHERE name = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		supplier, err := scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}

		suppliers = append(suppliers, *supplier)
	}

	return suppliers, nil
}

func (s *Store) GetSupplierBySearchContactPersonName(name string) ([]types.Supplier, error) {
	query := "SELECT COUNT(*) FROM supplier WHERE contact_person_name = ? AND deleted_at IS NULL"
	row := s.db.QueryRow(query, name)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var count int

	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	suppliers := make([]types.Supplier, 0)

	if count == 0 {
		query = "SELECT * FROM supplier WHERE contact_person_name LIKE ? AND deleted_at IS NULL"
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
			supplier, err := scanRowIntoSupplier(rows)

			if err != nil {
				return nil, err
			}

			suppliers = append(suppliers, *supplier)
		}

		return suppliers, nil
	}

	query = "SELECT * FROM supplier WHERE contact_person_name = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		supplier, err := scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}

		suppliers = append(suppliers, *supplier)
	}

	return suppliers, nil
}

func (s *Store) GetSupplierByID(id int) (*types.Supplier, error) {
	query := "SELECT * FROM supplier WHERE id = ? AND deleted_at IS NULL"
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	supplier := new(types.Supplier)

	for rows.Next() {
		supplier, err = scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}
	}

	if supplier.ID == 0 {
		return nil, fmt.Errorf("supplier not found")
	}

	return supplier, nil
}

func (s *Store) CreateSupplier(supplier types.Supplier) error {
	values := "?"
	for i := 0; i < 7; i++ {
		values += ", ?"
	}

	query := `INSERT INTO supplier (
		name, address, company_phone_number, contact_person_name, 
		contact_person_number, terms, vendor_is_taxable, last_modified_by_user_id
	) VALUES (` + values + `)`

	_, err := s.db.Exec(query,
		supplier.Name, supplier.Address, supplier.CompanyPhoneNumber,
		supplier.ContactPersonName, supplier.ContactPersonNumber,
		supplier.Terms, supplier.VendorIsTaxable, supplier.LastModifiedByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllSuppliers() ([]types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := make([]types.Supplier, 0)

	for rows.Next() {
		supplier, err := scanRowIntoSupplier(rows)

		if err != nil {
			return nil, err
		}

		suppliers = append(suppliers, *supplier)
	}

	return suppliers, nil
}

func (s *Store) DeleteSupplier(supplier *types.Supplier, user *types.User) error {
	query := "UPDATE supplier SET deleted_at = ?, deleted_by_user_id = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), user.ID, supplier.ID)
	if err != nil {
		return err
	}

	data, err := s.GetSupplierByID(supplier.ID)
	if err != nil {
		return err
	}

	err = logger.WriteLog("delete", "supplier", user.Name, data.ID, data)
	if err != nil {
		return fmt.Errorf("error write log file")
	}

	return nil
}

func (s *Store) ModifySupplier(sid int, newSupplierData types.Supplier, user *types.User) error {
	data, err := s.GetSupplierByID(sid)
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

	query := `UPDATE suppplier SET 
				name = ?, address = ?, company_phone_number = ?, contact_person_name = ?, 
				contact_person_number = ?, terms = ?, vendor_is_taxable = ?, 
				last_modified = ?, last_modified_by_user_id = ? 
				WHERE id = ?`
	_, err = s.db.Exec(query,
		newSupplierData.Name, newSupplierData.Address, newSupplierData.CompanyPhoneNumber,
		newSupplierData.ContactPersonName, newSupplierData.ContactPersonNumber,
		newSupplierData.Terms, newSupplierData.VendorIsTaxable, time.Now(),
		newSupplierData.LastModifiedByUserID, sid)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoSupplier(rows *sql.Rows) (*types.Supplier, error) {
	supplier := new(types.Supplier)

	err := rows.Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.Address,
		&supplier.CompanyPhoneNumber,
		&supplier.ContactPersonName,
		&supplier.ContactPersonNumber,
		&supplier.Terms,
		&supplier.VendorIsTaxable,
		&supplier.CreatedAt,
		&supplier.LastModified,
		&supplier.LastModifiedByUserID,
		&supplier.DeletedAt,
		&supplier.DeletedByUserID,
	)

	if err != nil {
		return nil, err
	}

	supplier.CreatedAt = supplier.CreatedAt.Local()
	supplier.LastModified = supplier.LastModified.Local()

	return supplier, nil
}
