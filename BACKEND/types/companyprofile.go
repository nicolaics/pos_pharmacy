package types

import (
	"time"
)

type RegisterCompanyProfilePayload struct {
	Name                    string `json:"name" validate:"required"`
	Address                 string `json:"address" validate:"required"`
	BusinessNumber          string `json:"businessNumber" validate:"required"`
	Pharmacist              string `json:"pharmacist" validate:"required"`
	PharmacistLicenseNumber string `json:"pharmacistLicenseNumber" validate:"required"`
}

type DeleteCompanyProfilePayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type ModifyCompanyProfilePayload struct {
	ID                         int    `json:"id" validate:"required"`
	NewName                    string `json:"newName" validate:"required"`
	NewAddress                 string `json:"newAddress" validate:"required"`
	NewBuinessNumber           string `json:"newBusinessNumber" validate:"required"`
	NewPharmacist              string `json:"newPharmacist" validate:"required"`
	NewPharmacistLicenseNumber string `json:"newPharmacistLicenseNumber" validate:"required"`
}

type CompanyProfileStore interface {
	GetCompanyProfileByName(string) (*CompanyProfile, error)
	GetCompanyProfileByID(int) (*CompanyProfile, error)
	CreateCompanyProfile(CompanyProfile) error
	GetAllCompanyProfiles() ([]CompanyProfile, error)
	DeleteCompanyProfile(*CompanyProfile) error
	ModifyCompanyProfile(int, CompanyProfile) error
}

type CompanyProfile struct {
	ID                      int       `json:"id"`
	Name                    string    `json:"name"`
	Address                 string    `json:"address"`
	BusinessNumber          string    `json:"businessNumber"`
	Pharmacist              string    `json:"pharmacist"`
	PharmacistLicenseNumber string    `json:"pharmacistLicenseNumber"`
	LastModified            time.Time `json:"lastModified"`
}
