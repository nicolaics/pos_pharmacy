package types

import (
	"time"
)

type RegisterSupplier struct {
	Name string `json:"name" validate:"required"`
}

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*Supplier, error)
	CreateSupplier(Supplier) error
}

type Supplier struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
