package types

import (
	"database/sql"
	"time"
)

type InvoiceStore interface {
	GetInvoiceByID(id int) (*Invoice, error)

	GetInvoicesByNumber(int) ([]Invoice, error)

	GetInvoicesByDate(startDate time.Time, endDate time.Time) ([]InvoiceListsReturnPayload, error)
	GetInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]InvoiceListsReturnPayload, error)
	GetInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]InvoiceListsReturnPayload, error)
	GetInvoicesByDateAndCustomerID(startDate time.Time, endDate time.Time, cid int) ([]InvoiceListsReturnPayload, error)
	GetInvoicesByDateAndPaymentMethodID(startDate time.Time, endDate time.Time, pmid int) ([]InvoiceListsReturnPayload, error)

	GetInvoiceID(number int, customerId int, invoiceDate time.Time) (int, error)
	GetNumberOfInvoices() (int, error)

	CreateInvoice(Invoice) error
	CreateMedicineItem(InvoiceMedicineItem) error
	GetMedicineItem(int) ([]InvoiceMedicineItemReturnPayload, error)
	DeleteMedicineItem(*Invoice, *User) error
	DeleteInvoice(*Invoice, *User) error
	ModifyInvoice(int, Invoice, *User) error

	UpdatePdfURL(invoiceId int, pdfUrl string) error

	// delete entirely from the db if there's error
	AbsoluteDeleteInvoice(invoice Invoice) error
}

type ViewInvoiceDetailPayload struct {
	InvoiceID int `json:"invoiceId" validate:"required"`
}

type RegisterInvoicePayload struct {
	Number            int     `json:"number" validate:"required"`
	CustomerID        int     `json:"customerId" validate:"required"`
	Subtotal          float64 `json:"subtotal" validate:"required"`
	Discount          float64 `json:"discount"`
	Tax               float64 `json:"tax"`
	TotalPrice        float64 `json:"totalPrice" validate:"required"`
	PaidAmount        float64 `json:"paidAmount" validate:"required"`
	ChangeAmount      float64 `json:"changeAmount"`
	PaymentMethodName string  `json:"paymentMethodName" validate:"required"`
	Description       string  `json:"description"`
	InvoiceDate       string  `json:"invoiceDate" validate:"required"`

	MedicineLists []InvoiceMedicineListsPayload `json:"medicineLists" validate:"required"`
}

type ModifyInvoicePayload struct {
	ID      int                    `json:"id" validate:"required"`
	NewData RegisterInvoicePayload `json:"newData" validate:"required"`
}

type ViewInvoicePayload struct {
	StartDate string `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   string `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

type InvoiceMedicineListsPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
}

type InvoiceMedicineItemReturnPayload struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	Qty             float64 `json:"qty"`
	Unit            string  `json:"unit"`
	Price           float64 `json:"price"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal"`
}

// viewing the invoice lists only
type InvoiceListsReturnPayload struct {
	ID                int       `json:"id"`
	Number            int       `json:"number"`
	UserName          string    `json:"userName"`
	CustomerName      string    `json:"customerName"`
	Subtotal          float64   `json:"subtotal"`
	Discount          float64   `json:"discount"`
	Tax               float64   `json:"tax"`
	TotalPrice        float64   `json:"totalPrice"`
	PaymentMethodName string    `json:"paymentMethodName"`
	Description       string    `json:"description"`
	InvoiceDate       time.Time `json:"invoiceDate"`
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
	PdfURL                 string    `json:"pdfUrl"`

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

	MedicineLists []InvoiceMedicineItemReturnPayload `json:"medicineLists"`
}

type DeleteInvoicePayload ViewInvoiceDetailPayload

type InvoiceMedicineItem struct {
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
	ID                   int           `json:"id"`
	Number               int           `json:"number"`
	UserID               int           `json:"userId"`
	CustomerID           int           `json:"customerId"`
	Subtotal             float64       `json:"subtotal"`
	Discount             float64       `json:"discount"`
	Tax                  float64       `json:"tax"`
	TotalPrice           float64       `json:"totalPrice"`
	PaidAmount           float64       `json:"paidAmount"`
	ChangeAmount         float64       `json:"changeAmount"`
	PaymentMethodID      int           `json:"paymentMethodId"`
	Description          string        `json:"description"`
	InvoiceDate          time.Time     `json:"invoiceDate"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastModifiedByUserId"`
	PdfURL               string        `json:"pdfUrl"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}
