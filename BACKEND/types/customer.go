package types

import (
	"time"
)

type RegisterCustomerPayload struct {
	Name string `json:"name" validate:"required"`
}
type ModifyCustomerPayload struct {
	ID      int    `json:"id" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type DeleteCustomerPayload struct {
	ID      int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type CustomerStore interface {
	GetCustomerByName(name string) (*Customer, error)
	GetCustomerByID(id int) (*Customer, error)
	CreateCustomer(Customer) error
	GetAllCustomers() ([]Customer, error)
	DeleteCustomer(*Customer) error
	ModifyCustomer(int, string) error
}

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
