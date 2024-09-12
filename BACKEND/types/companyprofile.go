package types

import (
	"database/sql"
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
	ID      int                           `json:"id" validate:"required"`
	NewData RegisterCompanyProfilePayload `json:"newData" validate:"required"`
}

type CompanyProfileStore interface {
	GetCompanyProfileByName(string) (*CompanyProfile, error)
	GetCompanyProfileByID(int) (*CompanyProfile, error)
	CreateCompanyProfile(CompanyProfile) error
	GetAllCompanyProfiles() ([]CompanyProfile, error)
	DeleteCompanyProfile(int, int) error
	ModifyCompanyProfile(int, int, CompanyProfile) error
}

type CompanyProfile struct {
	ID                      int       `json:"id"`
	Name                    string    `json:"name"`
	Address                 string    `json:"address"`
	BusinessNumber          string    `json:"businessNumber"`
	Pharmacist              string    `json:"pharmacist"`
	PharmacistLicenseNumber string    `json:"pharmacistLicenseNumber"`
	CreatedAt               time.Time `json:"createdAt"`
	LastModified            time.Time `json:"lastModified"`
	LastModifiedByUserID    int       `json:"lastModifiedByUserId"`
	DeletedAt               sql.NullTime `json:"deletedAt"`
	DeletedByUserID         sql.NullInt64       `json:"deletedByUserId"`
}
