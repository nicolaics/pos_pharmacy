package types

import (
	"time"
)

type PaymentMethodStore interface {
	GetPaymentMethodByName(paymentMethodName string) (*PaymentMethod, error)
	GetPaymentMethodByID(int) (*PaymentMethod, error)
	CreatePaymentMethod(paymentMethodName string) error
}

type PaymentMethod struct {
	ID        int       `json:"id"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"createdAt"`
}

type UnitStore interface {
	GetUnitByName(string) (*Unit, error)
	CreateUnit(string) error
}
type Unit struct {
	ID        int       `json:"id"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"createdAt"`
}

type TokenDetails struct {
	AccessToken     string `json:"accessToken"`
	RefreshToken    string `json:"refreshToken"`
	AccessUUID      string `json:"accessUuid"`
	RefreshUUID     string `json:"refreshUuid"`
	AccessTokenExp  int64  `json:"accessTokenExp"`
	RefreshTokenExp int64  `json:"refreshTokenExp"`
}

type AccessDetails struct {
	AccessUUID string
	CashierID  int
}

type RefreshDetails struct {
	RefreshUUID string
	CashierID   int
}
