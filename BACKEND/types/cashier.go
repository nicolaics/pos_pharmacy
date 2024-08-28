package types

import (
	"net/http"
	"time"
)

type InitAdminPayload struct {
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required,min=3,max=130"`
}

type RegisterCashierPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber   string `json:"phoneNumber" validate:"required"`
	MakeAdmin     bool   `json:"makeAdmin"`
}

type RemoveCashierPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
}

type UpdateCashierAdminPayload RemoveCashierPayload

type LoginCashierPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CashierStore interface {
	GetCashierByName(string) (*Cashier, error)
	GetCashierByID(int) (*Cashier, error)
	CreateCashier(Cashier) error
	DeleteCashier(*Cashier) error
	GetAllCashiers() ([]Cashier, error)
	UpdateLastLoggedIn(int) error
	UpdateAdmin(*Cashier) error
	SaveAuth(int, *TokenDetails) error
	GetCashierIDFromRedis(*AccessDetails) (int, error)
	DeleteAuth(string) (int, error)
	ValidateCashierToken(http.ResponseWriter, *http.Request, bool) (*Cashier, error)
}

type Cashier struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Password     string    `json:"password"`
	Admin        bool      `json:"admin"`
	PhoneNumber  string    `json:"phoneNumber"`
	LastLoggedIn time.Time `json:"lastLoggedIn"`
	CreatedAt    time.Time `json:"createdAt"`
}
