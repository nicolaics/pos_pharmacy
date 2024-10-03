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

type ConsumeWayStore interface {
	GetConsumeWayByName(string) (*ConsumeWay, error)
	GetConsumeWayByID(int) (*ConsumeWay, error)
	CreateConsumeWay(string) error
}
type ConsumeWay PaymentMethod

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

type PrescriptionSetUsageStore interface {
	GetPrescriptionSetUsageByName(string) (*PrescriptionSetUsage, error)
	GetPrescriptionSetUsageByID(int) (*PrescriptionSetUsage, error)
	CreatePrescriptionSetUsage(string) error
}
type PrescriptionSetUsage PaymentMethod