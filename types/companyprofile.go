package types

import (
	"time"
)

type RegisterCompanyProfilePayload struct {
	Name                    string `json:"name" validate:"required"`
	Address                 string `json:"address" validate:"required"`
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
	GetCompanyProfile() (*CompanyProfileReturn, error)
	ModifyCompanyProfile(int, *User, CompanyProfile) error
}

type CompanyProfileReturn struct {
	ID                      int       `json:"id"`
	Name                    string    `json:"name"`
	Address                 string    `json:"address"`
	LastModified            time.Time `json:"lastModified"`
	LastModifiedByUserName  string    `json:"lastModifiedByUserName"`
}

type CompanyProfile struct {
	ID                      int           `json:"id"`
	Name                    string        `json:"name"`
	Address                 string        `json:"address"`
	CreatedAt               time.Time     `json:"createdAt"`
	LastModified            time.Time     `json:"lastModified"`
	LastModifiedByUserID    int           `json:"lastModifiedByUserId"`
}
