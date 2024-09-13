package types

import (
	"database/sql"
	"time"
)

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*Supplier, error)

	GetSupplierBySearchName(name string) ([]Supplier, error)
	GetSupplierBySearchContactPersonName(name string) ([]Supplier, error)
	
	CreateSupplier(Supplier) error

	GetAllSuppliers() ([]Supplier, error)

	DeleteSupplier(*Supplier, int) error

	ModifySupplier(id int, newSupplierData Supplier, userId int) error
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
	ID      int                     `json:"id" validate:"required"`
	NewData RegisterSupplierPayload `json:"newData" validate:"required"`
}

type DeleteSupplierPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetOneSupplierPayload struct {
	ID int `json:"id" validate:"required"`
}

type Supplier struct {
	ID                   int           `json:"id"`
	Name                 string        `json:"name"`
	Address              string        `json:"address"`
	CompanyPhoneNumber   string        `json:"companyPhoneNumber"`
	ContactPersonName    string        `json:"contactPersonName"`
	ContactPersonNumber  string        `json:"contactPersonNumber"`
	Terms                string        `json:"terms"`
	VendorIsTaxable      bool          `json:"vendorIsTaxable"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastModifiedByUserId"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}
