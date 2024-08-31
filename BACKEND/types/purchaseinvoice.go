package types

import (
	"time"
)

type PurchaseInvoicePayload struct {
	Number        int                           `json:"number" validate:"required"`
	CompanyID     int                           `json:"companyId" validate:"required"` // can get the ID from the text box
	SupplierID    int                           `json:"supplierId" validate:"required"`
	Subtotal      float64                       `json:"subtotal" validate:"required"`
	Discount      float64                       `json:"discount"`
	Tax           float64                       `json:"tax" validate:"required"`
	TotalPrice    float64                       `json:"totalPrice" validate:"required"`
	Description   string                        `json:"description"`
	InvoiceDate   time.Time                     `json:"invoiceDate" validate:"required"`
	MedicineLists []PurchaseMedicineListPayload `json:"purchaseMedicineList" validate:"required"`
	CreatedAt     time.Time                     `json:"createdAt" validate:"required"`
}

type PurchaseMedicineListPayload struct {
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

// only view the purchase invoice list
type PurchaseInvoiceSummaryPayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type PurchaseMedicineItemsPayload struct {
	PurchaseInvoiceID int `json:"purchaseInvoiceId" validate:"required"`
}

type ModifyPurchaseInvoicePayload struct {
	PurchaseInvoiceID int                           `json:"purchaseInvoiceId" validate:"required"`
	NewNumber         int                           `json:"newNumber" validate:"required"`
	NewCompanyID      int                           `json:"newCompanyId" validate:"required"`
	NewSupplierID     int                           `json:"newSupplierId" validate:"required"`
	NewSubtotal       float64                       `json:"newSubtotal" validate:"required"`
	NewDiscount       float64                       `json:"newDiscount"`
	NewTax            float64                       `json:"newTax" validate:"required"`
	NewTotalPrice     float64                       `json:"newTotalPrice" validate:"required"`
	NewDescription    string                        `json:"newDescription"`
	NewInvoiceDate    time.Time                     `json:"newInvoiceDate" validate:"required"`
	NewMedicineLists  []PurchaseMedicineListPayload `json:"newPurchaseMedicineList" validate:"required"`
}

type PurchaseMedicineItemsReturn struct {
	ID           int       `json:"id"`
	MedicineName string    `json:"medicineName"`
	Qty          float64   `json:"qty"`
	Unit         string    `json:"unit"`
	Price        float64   `json:"price"`
	Discount     float64   `json:"discount"`
	Tax          float64   `json:"tax"`
	Subtotal     float64   `json:"subtotal"`
	BatchNumber  string    `json:"batchNumber"`
	ExpDate      time.Time `json:"expDate"`
}

type PurchaseInvoiceReturnJSONPayload struct {
	PurchaseInvoiceID          int       `json:"purchaseInvoiceId"`
	PurchaseInvoiceNumber      int       `json:"purchaseInvoiceNumber"`
	PurchaseInvoiceSubtotal    float64   `json:"purchaseInvoiceSubtotal"`
	PurchaseInvoiceDiscount    float64   `json:"purchaseInvoiceDiscount"`
	PurchaseInvoiceTax         float64   `json:"purchaseInvoiceTax"`
	PurchaseInvoiceTotalPrice  float64   `json:"purchaseInvoiceTotalPrice"`
	PurchaseInvoiceDescription string    `json:"purchaseInvoiceDescription"`
	PurchaseInvoiceInvoiceDate time.Time `json:"purchaseInvoiceInvoiceDate"`

	CompanyID               int    `json:"companyId"`
	CompanyName             string `json:"companyName"`
	CompanyAddress          string `json:"companyAddress"`
	CompanyBusinessNumber   string `json:"companyBusinessNumber"`
	Pharmacist              string `json:"pharmacist"`
	PharmacistLicenseNumber string `json:"pharmacistLicenseNumber"`

	SupplierID                  int    `json:"supplierId"`
	SupplierName                string `json:"supplierName"`
	SupplierAddress             string `json:"supplierAddress"`
	SupplierPhoneNumber         string `json:"supplierPhoneNumber"`
	SupplierContactPersonName   string `json:"supplierContactPersonName"`
	SupplierContactPersonNumber string `json:"supplierContactPersonNumber"`
	SupplierTerms               string `json:"supplierTerms"`
	SupplierVendorIsTaxable     bool   `json:"supplierVendorIsTaxable"`

	CashierID   int    `json:"cashierId"`
	CashierName string `json:"cashierName"`

	MedicineLists []PurchaseMedicineItemsReturn `json:"medicineLists"`
}

type DeletePurchaseInvoice struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseInvoiceStore interface {
	GetPurchaseInvoiceByNumber(int) (*PurchaseInvoice, error)
	GetPurchaseInvoiceByID(int) (*PurchaseInvoice, error)
	CreatePurchaseInvoice(PurchaseInvoice) error
	CreatePurchaseMedicineItems(PurchaseMedicineItem) error
	GetPurhcaseInvoices(startDate time.Time, endDate time.Time) ([]PurchaseInvoice, error)
	GetPurhcaseMedicineItems(purchaseInvoiceId int) ([]PurchaseMedicineItemsReturn, error)
	DeletePurchaseInvoice(*PurchaseInvoice) error
	DeletePurchaseMedicineItems(int) error
	ModifyPurchaseInvoice(int, PurchaseInvoice) error
}

type PurchaseInvoice struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	CompanyID   int       `json:"companyId"`
	SupplierID  int       `json:"supplierId"`
	Subtotal    float64   `json:"subtotal"`
	Discount    float64   `json:"discount"`
	Tax         float64   `json:"tax"`
	TotalPrice  float64   `json:"totalPrice"`
	Description string    `json:"description"`
	CashierID   int       `json:"cashierId"`
	InvoiceDate time.Time `json:"invoiceDate"`
	CreatedAt   time.Time `json:"createdAt"`
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
