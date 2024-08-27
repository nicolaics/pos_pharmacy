package types

import (
	"time"
)

type RegisterSupplierPayload struct {
	Name                string `json:"name" validate:"required"`
	Address             string `json:"address" validate:"required"`
	CompanyPhoneNumber  string `json:"companyPhoneNumber" validate:"required"`
	ContactPersonName   string `json:"contactPersonName"`
	ContactPersonNumber string `json:"contactPersonNumber"`
	Terms               string `json:"terms" validate:"required"`
	VendorIsTaxable     bool   `json:"vendorIsTaxable" validate:"required"`
}

type ModifySupplierPayload RegisterSupplierPayload

type DeleteSupplierPayload struct {
	Name string `json:"name" validate:"required"`
}

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*Supplier, error)
	CreateSupplier(Supplier) error
	GetAllSuppliers() ([]Supplier, error)
	DeleteSupplier(*Supplier) error
	ModifySupplier(id int, newSupplierData Supplier) error
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
}
