package types

import (
	"time"
)

type RegisterCashierPayload struct {
	AdminName     string `json:"adminName" validate:"required"`
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required,min=3,max=130"`
}

type RemoveCashierPayload struct {
	// AdminName     string `json:"adminName" validate:"required"`
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
}

type UpdateCashierAdminPayload RemoveCashierPayload

type RegisterCustomerPayload struct {
	Name string `json:"name" validate:"required"`
}

type RegisterSupplier struct {
	Name string `json:"name" validate:"required"`
}

type NewInvoice struct {
	Number            int       `json:"number" validate:"required"`
	CashierName       string    `json:"cashierName" validate:"required"`
	CustomerName      string    `json:"customerName" validate:"required"`
	Subtotal          float64   `json:"subtotal" validate:"required"`
	Discount          float64   `json:"discount"`
	TotalPrice        float64   `json:"totalPrice" validate:"required"`
	PaymentMethodName string    `json:"paymentMethodName" validate:"required"`
	PaidAmount        float64   `json:"paidAmount" validate:"required"`
	ChangeAmount      float64   `json:"changeAmount" validate:"required"`
	Description       string    `json:"description"`
	InvoiceDate       time.Time `json:"invoiceDate" validate:"required"`
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
	UpdateLastLoggedIn(*Cashier) error
	UpdateAdmin(*Cashier) error
	SaveAuth(int, *TokenDetails) error
	GetAuthentification(*AccessDetails) (int, error)
	DeleteAuth(givenUuid string) (int, error)
}

type Cashier struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Password     string    `json:"password"`
	Admin        bool      `json:"admin"`
	CreatedAt    time.Time `json:"createdAt"`
	LastLoggedIn time.Time `json:"lastLoggedIn"`
}

type PaymentMethod struct {
	ID        int       `json:"id"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"createdAt"`
}

type CustomerStore interface {
	GetCustomerByName(name string) (*Customer, error)
	GetCustomerByID(id int) (*Customer, error)
	CreateCustomer(Customer) error
}

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type SupplierStore interface {
	GetSupplierByName(name string) (*Supplier, error)
	GetSupplierByID(id int) (*Supplier, error)
	CreateSupplier(Supplier) error
}

type Supplier struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Unit struct {
	ID        int       `json:"id"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"createdAt"`
}

type PurchaseInvoice struct {
	ID            int       `json:"id"`
	Number        int       `json:"number"`
	SupplierID    string    `json:"supplierId"`
	Subtotal      float64   `json:"subtotal"`
	Discount      float64   `json:"discount"`
	TotalPrice    float64   `json:"totalPrice"`
	PaymentMethod string    `json:"paymentMethod"`
	PurchaseDate  time.Time `json:"purchaseDate"`
	PaidDate      time.Time `json:"paidDate"`
	CreatedAt     time.Time `json:"createdAt"`
}

type InvoiceStore interface {
	GetInvoiceByID(id int) (*Invoice, error)
	GetInvoicesByDate(date time.Time) ([]*Invoice, error)
	CreateInvoice(Invoice) error
}

type Invoice struct {
	ID                int       `json:"id"`
	Number            int       `json:"number"`
	CashierName       string    `json:"cashierName"`
	CustomerName      string    `json:"customerName"`
	Subtotal          float64   `json:"subtotal"`
	Discount          float64   `json:"discount"`
	TotalPrice        float64   `json:"totalPrice"`
	PaidAmount        float64   `json:"paidAmount"`
	ChangeAmount      float64   `json:"changeAmount"`
	PaymentMethodName string    `json:"paymentMethodName"`
	Description       string    `json:"description"`
	InvoiceDate       time.Time `json:"invoiceDate"`
	CreatedAt         time.Time `json:"createdAt"`
}

type Medicine struct {
	Barcode           string    `json:"barcode"`
	Name              string    `json:"name"`
	UnitID            string    `json:"unitId"`
	Stock             float64   `json:"stock"`
	PurchaseInvoiceID int       `json:"purchaseInvoiceId"`
	Price             float64   `json:"price"`
	CreatedAt         time.Time `json:"createdAt"`
}

type PurchaseMedicineItems struct {
	ID                int       `json:"id"`
	PurchaseInvoiceID int       `json:"purchaseInvoiceId"`
	MedicineBarcode   string    `json:"medicineBarcode"`
	Qty               float64   `json:"qty"`
	UnitID            string    `json:"unitId"`
	PurchasePrice     float64   `json:"purchasePrice"`
	PurchaseDiscount  float64   `json:"purchaseDiscount"`
	Subtotal          float64   `json:"subtotal"`
	CreatedAt         time.Time `json:"createdAt"`
}

type MedicineItems struct {
	ID              int       `json:"id"`
	InvoiceID       int       `json:"invoiceId"`
	MedicineBarcode string    `json:"medicineBarcode"`
	Qty             float64   `json:"qty"`
	UnitID          string    `json:"unitId"`
	Price           float64   `json:"price"`
	Discount        float64   `json:"discount"`
	Subtotal        float64   `json:"subtotal"`
	CreatedAt       time.Time `json:"createdAt"`
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
