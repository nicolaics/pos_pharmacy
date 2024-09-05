package types

import (
	"time"
)

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*Supplier, error)
	CreateSupplier(Supplier) error
	GetAllSuppliers() ([]Supplier, error)
	DeleteSupplier(*Supplier, int) error
	ModifySupplier(id int, newSupplierData Supplier) error
}

type RegisterSupplierPayload struct {
	Name                string `json:"name" validate:"required"`
	Address             string `json:"address" validate:"required"`
	CompanyPhoneNumber  string `json:"companyPhoneNumber" validate:"required"`
	ContactPersonName   string `json:"contactPersonName"`
	ContactPersonNumber string `json:"contactPersonNumber"`
	Terms               string `json:"terms" validate:"required"`
	VendorIsTaxable     bool   `json:"vendorIsTaxable" validate:"required"`
}

type ModifySupplierPayload struct {
	ID                     int    `json:"id" validate:"required"`
	NewName                string `json:"newName" validate:"required"`
	NewAddress             string `json:"newAddress" validate:"required"`
	NewCompanyPhoneNumber  string `json:"newCompanyPhoneNumber" validate:"required"`
	NewContactPersonName   string `json:"newContactPersonName"`
	NewContactPersonNumber string `json:"newContactPersonNumber"`
	NewTerms               string `json:"newTerms" validate:"required"`
	NewVendorIsTaxable     bool   `json:"newVendorIsTaxable" validate:"required"`
}

type DeleteSupplierPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type Supplier struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Address             string    `json:"address"`
	CompanyPhoneNumber  string    `json:"companyPhoneNumber"`
	ContactPersonName   string    `json:"contactPersonName"`
	ContactPersonNumber string    `json:"contactPersonNumber"`
	Terms               string    `json:"terms"`
	VendorIsTaxable     bool      `json:"vendorIsTaxable"`
	CreatedAt           time.Time `json:"createdAt"`
	LastModified     time.Time `json:"lastModified"`
	ModifiedByUserID int       `json:"modifiedByUserId"`
	DeletedAt        time.Time `json:"deletedAt"`
	DeletedByUserID  int       `json:"deletedByUserId"`
}
