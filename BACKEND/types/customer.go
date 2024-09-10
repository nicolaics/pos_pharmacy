package types

import (
	"time"
)

type CustomerStore interface {
	GetCustomerByName(name string) (*Customer, error)
	GetCustomerByID(id int) (*Customer, error)
	CreateCustomer(Customer) error
	GetAllCustomers() ([]Customer, error)
	DeleteCustomer(int, *Customer) error
	ModifyCustomer(int, string, int) error
}

type RegisterCustomerPayload struct {
	Name string `json:"name" validate:"required"`
}
type ModifyCustomerPayload struct {
	ID      int                     `json:"id" validate:"required"`
	NewData RegisterCustomerPayload `json:"newData" validate:"required"`
}

type DeleteCustomerPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetOneCustomerPayload struct {
	ID   int    `json:"id" validate:"required"`
}

type Customer struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"createdAt"`
	DeletedAt       time.Time `json:"deletedAt"`
	DeletedByUserID int       `json:"deletedByUserId"`
}
