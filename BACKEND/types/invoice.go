package types

import (
	"time"
)

type InvoiceStore interface {
	GetInvoiceByID(id int) (*Invoice, error)
	GetInvoiceByAll(number int, cashierId int, customerId int, totalPrice float64, invoiceDate time.Time) (*Invoice, error)
	GetInvoicesByNumber(int) ([]Invoice, error)
	GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]*Invoice, error)
	CreateInvoice(Invoice) error
	CreateMedicineItems(MedicineItems) error
}

type NewInvoicePayload struct {
	Number            int       `json:"number" validate:"required"`
	CustomerID        int       `json:"customerId" validate:"required"`
	Subtotal          float64   `json:"subtotal" validate:"required"`
	Discount          float64   `json:"discount"`
	Tax               float64   `json:"tax"`
	TotalPrice        float64   `json:"totalPrice" validate:"required"`
	PaidAmount        float64   `json:"paidAmount" validate:"required"`
	ChangeAmount      float64   `json:"changeAmount" validate:"required"`
	PaymentMethodName string    `json:"paymentMethodString" validate:"required"`
	Description       string    `json:"description"`
	InvoiceDate       time.Time `json:"invoiceDate" validate:"required"`

	MedicineLists []MedicineListsPayload `json:"medicineLists" validate:"required"`
}

type ViewInvoicePayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

type MedicineListsPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
}

type MedicineItems struct {
	ID         int       `json:"id"`
	InvoiceID  int       `json:"invoiceId"`
	MedicineID int       `json:"medicineId"`
	Qty        float64   `json:"qty"`
	UnitID     int    `json:"unitId"`
	Price      float64   `json:"price"`
	Discount   float64   `json:"discount"`
	Subtotal   float64   `json:"subtotal"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Invoice struct {
	ID              int       `json:"id"`
	Number          int       `json:"number"`
	CashierID       int       `json:"cashierId"`
	CustomerID      int       `json:"customerId"`
	Subtotal        float64   `json:"subtotal"`
	Discount        float64   `json:"discount"`
	Tax             float64   `json:"tax"`
	TotalPrice      float64   `json:"totalPrice"`
	PaidAmount      float64   `json:"paidAmount"`
	ChangeAmount    float64   `json:"changeAmount"`
	PaymentMethodID int       `json:"paymentMethodId"`
	Description     string    `json:"description"`
	InvoiceDate     time.Time `json:"invoiceDate"`
	CreatedAt       time.Time `json:"createdAt"`
}
