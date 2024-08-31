package types

import (
	"net/http"
	"time"
)

type InitAdminPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=3,max=130"`
}

type RegisterCashierPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber   string `json:"phoneNumber" validate:"required"`
	Admin         bool   `json:"admin"`
}

type RemoveCashierPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	ID            int    `json:"id" validate:"required"`
	Name          string `json:"name" validate:"required"`
}

type ModifyCashierPayload struct {
	AdminPassword  string `json:"adminPassword" validate:"required"`
	ID             int    `json:"id" validate:"required"`
	NewName        string `json:"newName" validate:"required"`
	NewPassword    string `json:"newPassword" validate:"required,min=3,max=130"`
	NewAdmin       bool   `json:"newAdmin" validate:"required"`
	NewPhoneNumber string `json:"newPhoneNumber" validate:"required"`
}

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
	ModifyCashier(int, Cashier) error
	SaveToken(int, *TokenDetails) error
	GetCashierIDFromRedis(*AccessDetails, *RefreshDetails) (int, error)
	DeleteToken(string) (int, error)
	ValidateCashierAccessToken(http.ResponseWriter, *http.Request, bool) (*Cashier, error)
	ValidateCashierRefreshToken(http.ResponseWriter, *http.Request) (*Cashier, error)
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
