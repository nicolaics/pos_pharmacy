package types

import (
	"database/sql"
	"time"
)

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*SupplierInformationReturnPayload, error)

	GetSupplierBySearchName(name string) ([]SupplierInformationReturnPayload, error)
	GetSupplierBySearchContactPersonName(name string) ([]SupplierInformationReturnPayload, error)

	CreateSupplier(Supplier) error

	GetAllSuppliers() ([]SupplierInformationReturnPayload, error)

	DeleteSupplier(*Supplier, *User) error

	ModifySupplier(id int, newSupplierData Supplier, user *User) error
}

type RegisterSupplierPayload struct {
	Name                string `json:"name" validate:"required"`
	Address             string `json:"address" validate:"required"`
	CompanyPhoneNumber  string `json:"companyPhoneNumber" validate:"required"`
	ContactPersonName   string `json:"contactPersonName"`
	ContactPersonNumber string `json:"contactPersonNumber"`
	Terms               string `json:"terms" validate:"required"`
	VendorIsTaxable     bool   `json:"vendorIsTaxable"`
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

type SupplierInformationReturnPayload struct {
	ID                     int       `json:"id"`
	Name                   string    `json:"name"`
	Address                string    `json:"address"`
	CompanyPhoneNumber     string    `json:"companyPhoneNumber"`
	ContactPersonName      string    `json:"contactPersonName"`
	ContactPersonNumber    string    `json:"contactPersonNumber"`
	Terms                  string    `json:"terms"`
	VendorIsTaxable        bool      `json:"vendorIsTaxable"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastModifiedByUserName"`
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
