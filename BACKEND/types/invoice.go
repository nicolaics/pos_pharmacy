package types

import (
	"time"
)

type InvoiceStore interface {
	GetInvoiceByID(id int) (*Invoice, error)
	GetInvoiceByAll(number int, userId int, customerId int, totalPrice float64, invoiceDate time.Time) (*Invoice, error)
	GetInvoicesByNumber(int) ([]Invoice, error)
	GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]Invoice, error)
	CreateInvoice(Invoice) error
	CreateMedicineItems(MedicineItems) error
	GetMedicineItems(int) ([]MedicineItemReturnPayload, error)
	DeleteMedicineItems(int) error
	DeleteInvoice(*Invoice) error
	ModifyInvoice(int, Invoice) error
}

type ViewInvoiceDetailPayload struct {
	InvoiceID int `json:"invoiceId" validate:"required"`
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

type ModifyInvoicePayload struct {
	ID                   int       `json:"id" validate:"required"`
	NewNumber            int       `json:"newNumber" validate:"required"`
	NewCustomerID        int       `json:"newcustomerId" validate:"required"`
	NewSubtotal          float64   `json:"newsubtotal" validate:"required"`
	NewDiscount          float64   `json:"newdiscount"`
	NewTax               float64   `json:"newTax"`
	NewTotalPrice        float64   `json:"newTotalPrice" validate:"required"`
	NewPaidAmount        float64   `json:"newPaidAmount" validate:"required"`
	NewChangeAmount      float64   `json:"newChangeAmount" validate:"required"`
	NewPaymentMethodName string    `json:"newPaymentMethodString" validate:"required"`
	NewDescription       string    `json:"newDescription"`
	NewInvoiceDate       time.Time `json:"newInvoiceDate" validate:"required"`

	NewMedicineLists []MedicineListsPayload `json:"newMedicineLists" validate:"required"`
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

type MedicineItemReturnPayload struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
}

type InvoiceDetailPayload struct {
	InvoiceID          int       `json:"invoiceId"`
	InvoiceNumber      int       `json:"invoiceNumber"`
	InvoiceSubtotal    float64   `json:"invoiceSubtotal"`
	InvoiceDiscount    float64   `json:"invoiceDiscount"`
	InvoiceTax         float64   `json:"invoiceTax"`
	InvoiceTotalPrice  float64   `json:"invoiceTotalPrice"`
	PaidAmount         float64   `json:"paidAmount"`
	ChangeAmount       float64   `json:"changeAmount"`
	InvoiceDescription string    `json:"invoiceDescription"`
	InvoiceDate        time.Time `json:"invoiceDate"`

	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	// UserID   int    `json:"userId"`
	// UserName string `json:"userName"`

	Customer struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"customer"`

	// CustomerID   int    `json:"customerId"`
	// CustomerName string `json:"customerName"`

	PaymentMethod struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"paymentMethod"`

	// PaymentMethodID   int    `json:"paymentMethodId"`
	// PaymentMethodName string `json:"paymentMethodName"`

	MedicineLists []MedicineItemReturnPayload `json:"medicineLists"`
}

type DeleteInvoicePayload ViewInvoiceDetailPayload

type MedicineItems struct {
	ID         int       `json:"id"`
	InvoiceID  int       `json:"invoiceId"`
	MedicineID int       `json:"medicineId"`
	Qty        float64   `json:"qty"`
	UnitID     int       `json:"unitId"`
	Price      float64   `json:"price"`
	Discount   float64   `json:"discount"`
	Subtotal   float64   `json:"subtotal"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TODO: made some changes with the DB, check the store.go as well
type Invoice struct {
	ID              int       `json:"id"`
	Number          int       `json:"number"`
	UserID          int       `json:"userId"`
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
