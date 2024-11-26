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
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type UnitStore interface {
	GetUnitByName(string) (*Unit, error)
	GetUnitByID(int) (*Unit, error)
	CreateUnit(string) error
}
type Unit PaymentMethod

type ConsumeTimeStore interface {
	GetConsumeTimeByName(string) (*ConsumeTime, error)
	GetConsumeTimeByID(int) (*ConsumeTime, error)
	CreateConsumeTime(string) error
}
type ConsumeTime PaymentMethod

type DoseStore interface {
	GetDoseByName(string) (*Dose, error)
	GetDoseByID(int) (*Dose, error)
	CreateDose(string) error
}
type Dose PaymentMethod

type DetStore interface {
	GetDetByName(string) (*Det, error)
	GetDetByID(int) (*Det, error)
	CreateDet(string) error
}
type Det PaymentMethod

type MfStore interface {
	GetMfByName(string) (*Mf, error)
	GetMfByID(int) (*Mf, error)
	CreateMf(string) error
}
type Mf PaymentMethod

type UsageStore interface {
	GetUsageByName(string) (*Usage, error)
	GetUsageByID(int) (*Usage, error)
	CreateUsage(string) error
}
type Usage PaymentMethod
