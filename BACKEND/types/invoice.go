package types

import (
	"time"
)

type InvoiceStore interface {
	GetInvoiceByID(id int) (*Invoice, error)
	GetInvoicesByNumber(int) ([]Invoice, error)
	GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]Invoice, error)

	GetInvoiceID(number int, userId int, customerId int, totalPrice float64, invoiceDate time.Time) (int, error)
	GetNumberOfInvoices() (int, error)

	CreateInvoice(Invoice) error
	CreateMedicineItems(MedicineItems) error
	GetMedicineItems(int) ([]MedicineItemReturnPayload, error)
	DeleteMedicineItems(*Invoice, int) error
	DeleteInvoice(*Invoice, int) error
	ModifyInvoice(int, Invoice) error
}

type ViewInvoiceDetailPayload struct {
	InvoiceID int `json:"invoiceId" validate:"required"`
}

type RegisterInvoicePayload struct {
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
	ID      int               `json:"id" validate:"required"`
	NewData RegisterInvoicePayload `json:"newData" validate:"required"`
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
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	Qty             float64 `json:"qty"`
	Unit            string  `json:"unit"`
	Price           float64 `json:"price"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal"`
}

type InvoiceDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	Subtotal               float64   `json:"subtotal"`
	Discount               float64   `json:"discount"`
	Tax                    float64   `json:"tax"`
	TotalPrice             float64   `json:"totalPrice"`
	PaidAmount             float64   `json:"paidAmount"`
	ChangeAmount           float64   `json:"changeAmount"`
	Description            string    `json:"description"`
	InvoiceDate            time.Time `json:"invoiceDate"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastModifiedByUserName"`

	// the one who creates the invoice
	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	Customer struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"customer"`

	PaymentMethod struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"paymentMethod"`

	MedicineLists []MedicineItemReturnPayload `json:"medicineLists"`
}

type DeleteInvoicePayload ViewInvoiceDetailPayload

type MedicineItems struct {
	ID         int     `json:"id"`
	InvoiceID  int     `json:"invoiceId"`
	MedicineID int     `json:"medicineId"`
	Qty        float64 `json:"qty"`
	UnitID     int     `json:"unitId"`
	Price      float64 `json:"price"`
	Discount   float64 `json:"discount"`
	Subtotal   float64 `json:"subtotal"`
}

type Invoice struct {
	ID                   int       `json:"id"`
	Number               int       `json:"number"`
	UserID               int       `json:"userId"`
	CustomerID           int       `json:"customerId"`
	Subtotal             float64   `json:"subtotal"`
	Discount             float64   `json:"discount"`
	Tax                  float64   `json:"tax"`
	TotalPrice           float64   `json:"totalPrice"`
	PaidAmount           float64   `json:"paidAmount"`
	ChangeAmount         float64   `json:"changeAmount"`
	PaymentMethodID      int       `json:"paymentMethodId"`
	Description          string    `json:"description"`
	InvoiceDate          time.Time `json:"invoiceDate"`
	CreatedAt            time.Time `json:"createdAt"`
	LastModified         time.Time `json:"lastModified"`
	LastModifiedByUserID int       `json:"lastModifiedByUserId"`
	DeletedAt            time.Time `json:"deletedAt"`
	DeletedByUserID      int       `json:"deletedByUserId"`
}
