package types

import (
	"database/sql"
	"time"
)

type PurchaseInvoiceStore interface {
	GetPurchaseInvoicesByNumber(number int) ([]PurchaseInvoice, error)
	GetPurchaseInvoiceByID(int) (*PurchaseInvoice, error)
	GetPurchaseInvoiceID(number int, companyId int, supplierId int, subtotal float64, totalPrice float64, invoiceDate time.Time) (int, error)
	GetPurchaseMedicineItem(purchaseInvoiceId int) ([]PurchaseMedicineItemReturn, error)

	GetPurchaseInvoicesByDate(startDate time.Time, endDate time.Time) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndPOINumber(startDate time.Time, endDate time.Time, poiNumber int) ([]PurchaseInvoiceListsReturnPayload, error)

	CreatePurchaseInvoice(PurchaseInvoice) error
	CreatePurchaseMedicineItem(PurchaseMedicineItem) error

	DeletePurchaseInvoice(*PurchaseInvoice, *User) error
	DeletePurchaseMedicineItem(*PurchaseInvoice, *User) error

	ModifyPurchaseInvoice(int, PurchaseInvoice, *User) error

	// delete entirely from the db if there's error
	AbsoluteDeletePurchaseInvoice(pi PurchaseInvoice) error
}

type PurchaseInvoicePayload struct {
	Number                     int                           `json:"number" validate:"required"`
	CompanyID                  int                           `json:"companyId" validate:"required"` // can get the ID from the text box
	SupplierID                 int                           `json:"supplierId" validate:"required"`
	PurchaseOrderInvoiceNumber int                           `json:"purchaseOrderInvoiceNumber"`
	Subtotal                   float64                       `json:"subtotal" validate:"required"`
	Discount                   float64                       `json:"discount"`
	Tax                        float64                       `json:"tax" validate:"required"`
	TotalPrice                 float64                       `json:"totalPrice" validate:"required"`
	Description                string                        `json:"description"`
	InvoiceDate                string                        `json:"invoiceDate" validate:"required"`
	MedicineLists              []PurchaseMedicineListPayload `json:"purchaseMedicineList" validate:"required"`
}

type PurchaseMedicineListPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Tax             float64 `json:"tax"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
	BatchNumber     string  `json:"batchNumber" validate:"required"`
	ExpDate         string  `json:"expDate" validate:"required"`
}

// only view the purchase invoice list
type ViewPurchaseInvoicePayload struct {
	StartDate string `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   string `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type ViewPurchaseMedicineItemPayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyPurchaseInvoicePayload struct {
	ID      int                    `json:"id" validate:"required"`
	NewData PurchaseInvoicePayload `json:"newData" validate:"required"`
}

type PurchaseMedicineItemReturn struct {
	ID              int       `json:"id"`
	MedicineBarcode string    `json:"medicineBarcode"`
	MedicineName    string    `json:"medicineName"`
	Qty             float64   `json:"qty"`
	Unit            string    `json:"unit"`
	Price           float64   `json:"price"`
	Discount        float64   `json:"discount"`
	Tax             float64   `json:"tax"`
	Subtotal        float64   `json:"subtotal"`
	BatchNumber     string    `json:"batchNumber"`
	ExpDate         time.Time `json:"expDate"`
}

type PurchaseInvoiceDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	Subtotal               float64   `json:"subtotal"`
	Discount               float64   `json:"discount"`
	Tax                    float64   `json:"tax"`
	TotalPrice             float64   `json:"totalPrice"`
	Description            string    `json:"description"`
	InvoiceDate            time.Time `json:"invoiceDate"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastModifiedByUserName"`

	CompanyProfile struct {
		ID                      int    `json:"id"`
		Name                    string `json:"name"`
		Address                 string `json:"address"`
		BusinessNumber          string `json:"businessNumber"`
		Pharmacist              string `json:"pharmacist"`
		PharmacistLicenseNumber string `json:"pharmacistLicenseNumber"`
	} `json:"companyProfile"`

	Supplier struct {
		ID                  int    `json:"id"`
		Name                string `json:"name"`
		Address             string `json:"address"`
		CompanyPhoneNumber  string `json:"companyPhoneNumber"`
		ContactPersonName   string `json:"contactPersonName"`
		ContactPersonNumber string `json:"contactPersonNumber"`
		Terms               string `json:"terms"`
		VendorIsTaxable     bool   `json:"vendorIsTaxable"`
	} `json:"supplier"`

	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	PurchaseOrderInvoiceNumber int `json:"purchaseOrderInvoiceNumber"`

	MedicineLists []PurchaseMedicineItemReturn `json:"medicineLists"`
}

// view the lists of the purchase invoice
type PurchaseInvoiceListsReturnPayload struct {
	ID                         int       `json:"id"`
	Number                     int       `json:"number"`
	SupplierName               string    `json:"supplierName"`
	PurchaseOrderInvoiceNumber int       `json:"purchaseOrderInvoiceNumber"`
	TotalPrice                 float64   `json:"totalPrice"`
	Description                string    `json:"description"`
	UserName                   string    `json:"userName"`
	InvoiceDate                time.Time `json:"invoiceDate"`
}

type DeletePurchaseInvoice struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseInvoice struct {
	ID                         int           `json:"id"`
	Number                     int           `json:"number"`
	CompanyID                  int           `json:"companyId"`
	SupplierID                 int           `json:"supplierId"`
	PurchaseOrderInvoiceNumber int           `json:"purchaseOrderInvoiceNumber"`
	Subtotal                   float64       `json:"subtotal"`
	Discount                   float64       `json:"discount"`
	Tax                        float64       `json:"tax"`
	TotalPrice                 float64       `json:"totalPrice"`
	Description                string        `json:"description"`
	UserID                     int           `json:"userId"`
	InvoiceDate                time.Time     `json:"invoiceDate"`
	CreatedAt                  time.Time     `json:"createdAt"`
	LastModified               time.Time     `json:"lastModified"`
	LastModifiedByUserID       int           `json:"lastModifiedByUserId"`
	DeletedAt                  sql.NullTime  `json:"deletedAt"`
	DeletedByUserID            sql.NullInt64 `json:"deletedByUserId"`
}

type PurchaseMedicineItem struct {
	ID                int       `json:"id"`
	PurchaseInvoiceID int       `json:"purchaseInvoiceId"`
	MedicineID        int       `json:"medicineId"`
	Qty               float64   `json:"qty"`
	UnitID            int       `json:"unitId"`
	PurchasePrice     float64   `json:"purchasePrice"`
	PurchaseDiscount  float64   `json:"purchaseDiscount"`
	PurchaseTax       float64   `json:"purchaseTax"`
	Subtotal          float64   `json:"subtotal"`
	BatchNumber       string    `json:"batchNumber"`
	ExpDate           time.Time `json:"expDate"`
}
