package companyprofile

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

func (s *Store) GetCompanyProfileByName(name string) (*types.CompanyProfile, error) {
	rows, err := s.db.Query("SELECT * FROM self_company_profile WHERE name = ? ", name)
	if err != nil {
		return nil, err
	}

	companyProfile := new(types.CompanyProfile)

	for rows.Next() {
		companyProfile, err = scanRowIntoCompanyProfile(rows)

		if err != nil {
			return nil, err
		}
	}

	if companyProfile.ID == 0 {
		return nil, fmt.Errorf("company profile not found")
	}

	return companyProfile, nil
}

func (s *Store) GetCompanyProfileByID(id int) (*types.CompanyProfile, error) {
	rows, err := s.db.Query("SELECT * FROM supplier WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	companyProfile := new(types.CompanyProfile)

	for rows.Next() {
		companyProfile, err = scanRowIntoCompanyProfile(rows)

		if err != nil {
			return nil, err
		}
	}

	if companyProfile.ID == 0 {
		return nil, fmt.Errorf("company profile not found")
	}

	return companyProfile, nil
}

func (s *Store) CreateCompanyProfile(companyProfile types.CompanyProfile) error {
	fields := "name, address, business_number, pharmacist, pharmacist_license_number"
	values := "?, ?, ?, ?, ?"

	_, err := s.db.Exec(fmt.Sprintf("INSERT INTO supplier (%s) VALUES (%s)", fields, values),
		companyProfile.Name, companyProfile.Address, companyProfile.BusinessNumber,
		companyProfile.Pharmacist, companyProfile.PharmacistLicenseNumber)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllCompanyProfiles() ([]types.CompanyProfile, error) {
	rows, err := s.db.Query("SELECT * FROM supplier")

	if err != nil {
		return nil, err
	}

	companyProfiles := make([]types.CompanyProfile, 0)

	for rows.Next() {
		companyProfile, err := scanRowIntoCompanyProfile(rows)

		if err != nil {
			return nil, err
		}

		companyProfiles = append(companyProfiles, *companyProfile)
	}

	return companyProfiles, nil
}

func (s *Store) DeleteCompanyProfile(companyProfile *types.CompanyProfile) error {
	_, err := s.db.Exec("DELETE FROM self_company_profile WHERE id = ? ", companyProfile.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyCompanyProfile(id int, newCompanyProfile types.CompanyProfile) error {
	columns := "name = ?, address = ?, business_number = ?, pharmacist = ?, pharmacist_license_number = ?"

	_, err := s.db.Exec(fmt.Sprintf("UPDATE supplier SET %s WHERE id = ?", columns),
		newCompanyProfile.Name, newCompanyProfile.Address, newCompanyProfile.BusinessNumber,
		newCompanyProfile.Pharmacist, newCompanyProfile.PharmacistLicenseNumber, id)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoCompanyProfile(rows *sql.Rows) (*types.CompanyProfile, error) {
	companyProfile := new(types.CompanyProfile)

	err := rows.Scan(
		&companyProfile.ID,
		&companyProfile.Name,
		&companyProfile.Address,
		&companyProfile.BusinessNumber,
		&companyProfile.Pharmacist,
		&companyProfile.PharmacistLicenseNumber,
		&companyProfile.LastModified,
	)

	if err != nil {
		return nil, err
	}

	companyProfile.LastModified = companyProfile.LastModified.Local()

	return companyProfile, nil
}
