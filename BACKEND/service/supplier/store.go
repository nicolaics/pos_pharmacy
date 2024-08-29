package supplier

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

func (s *Store) GetSupplierByName(name string) (*types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE name = ? ", name)

	if err != nil {
		return nil, err
	}

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

func (s *Store) GetSupplierByID(id int) (*types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

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
	fields := "name, address, company_phone_number, contact_person_name, contact_person_number, terms, vendor_is_taxable"
	values := "?, ?, ?, ?, ?, ?, ?"

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO supplier (%s) VALUES (%s)", fields, values),
		supplier.Name, supplier.Address, supplier.CompanyPhoneNumber,
		supplier.ContactPersonName, supplier.ContactPersonNumber,
		supplier.Terms, supplier.VendorIsTaxable)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllSuppliers() ([]types.Supplier, error) {
	rows, err := s.db.Query("SELECT * FROM supplier")

	if err != nil {
		return nil, err
	}

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

func (s *Store) DeleteSupplier(supplier *types.Supplier) error {
	_, err := s.db.Exec("DELETE FROM supplier WHERE id = ? ", supplier.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifySupplier(id int, newSupplierData types.Supplier) error {
	columns := "name = ?, address = ?, company_phone_number = ?, contact_person_name = ?, contact_person_number = ?, terms = ?, vendor_is_taxable = ?"

	_, err := s.db.Exec(fmt.Sprintf("UPDATE supplier SET %s WHERE id = ?", columns),
		newSupplierData.Name, newSupplierData.Address, newSupplierData.CompanyPhoneNumber,
		newSupplierData.ContactPersonName, newSupplierData.ContactPersonNumber,
		newSupplierData.Terms, newSupplierData.VendorIsTaxable, id)
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
		&supplier.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return supplier, nil
}
