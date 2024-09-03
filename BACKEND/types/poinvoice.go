package types

import (
	"time"
)

type PurchaseOrderInvoiceStore interface {
	GetPurchaseOrderInvoicesByNumber(int) ([]PurchaseOrderInvoice, error)
	GetPurchaseOrderInvoiceByID(int) (*PurchaseOrderInvoice, error)
	GetPurchaseOrderInvoiceByAll(number int, companyId int, supplierId int, cashierId int, totalItems int, invoiceDate time.Time) (*PurchaseOrderInvoice, error)
	CreatePurchaseOrderInvoice(PurchaseOrderInvoice) error
	CreatePurchaseOrderItems(PurchaseOrderItem) error
	GetPurchaseOrderInvoices(startDate time.Time, endDate time.Time) ([]PurchaseOrderInvoice, error)
	GetPurchaseOrderItems(purchaseOrderInvoiceId int) ([]PurchaseOrderItemsReturn, error)
	DeletePurchaseOrderInvoice(*PurchaseOrderInvoice) error
	DeletePurchaseOrderItems(int) error
	ModifyPurchaseOrderInvoice(int, PurchaseOrderInvoice) error
}

type NewPurchaseOrderInvoicePayload struct {
	Number       int       `json:"number" validate:"required"`
	CompanyID    int       `json:"companyId" validate:"required"`
	SupplierID   int       `json:"supplierId" validate:"required"`
	TotalItems   int       `json:"totalItems" validate:"required"`
	InvoiceDate  time.Time `json:"invoiceDate" validate:"required"`
	LastModified time.Time `json:"lastModified" validate:"required"`
	CreatedAt    time.Time `json:"createdAt" validate:"required"`

	MedicineLists []PurchaseOrderMedicineListPayload `json:"purchaseOrderMedicineList" validate:"required"`
}

type PurchaseOrderMedicineListPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	OrderQty        float64 `json:"orderQty" validate:"required"`
	ReceivedQty     float64 `json:"receivedQty"`
	Unit            string  `json:"unit" validate:"required"`
	Remarks         string  `json:"remarks"`
}

// only view the purchase invoice list
type ViewOnePurchaseOrderInvoicePayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type ViewPurchaseOrderItemsPayload struct {
	PurchaseOrderInvoiceID int `json:"purchaseOrderInvoiceId" validate:"required"`
}

type ModifyPurchaseOrderInvoicePayload struct {
	PurchaseOrderInvoiceID int       `json:"purchaseOrderInvoiceId" validate:"required"`
	NewNumber              int       `json:"newNumber" validate:"required"`
	NewCompanyID           int       `json:"newCompanyId" validate:"required"`
	NewSupplierID          int       `json:"newSupplierId" validate:"required"`
	NewTotalItems          int       `json:"newTotalItems" validate:"required"`
	NewInvoiceDate         time.Time `json:"newInvoiceDate" validate:"required"`
	NewLastModified        time.Time `json:"newLastModified" validate:"required"`

	NewMedicineLists []PurchaseOrderMedicineListPayload `json:"purchaseOrderMedicineList" validate:"required"`
}

type PurchaseOrderItemsReturn struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	OrderQty        float64 `json:"orderQty"`
	ReceivedQty     float64 `json:"receivedQty"`
	Unit            string  `json:"unit"`
	Remarks         string  `json:"remarks"`
}

type PurchaseOrderInvoiceDetailPayload struct {
	PurchaseOrderInvoiceID           int       `json:"purchaseOrderInvoiceId"`
	PurchaseOrderInvoiceNumber       int       `json:"purchaseOrderInvoiceNumber"`
	PurchaseOrderInvoiceTotalItems   int       `json:"purchaseOrderInvoiceTotalItems"`
	PurchaseOrderInvoiceInvoiceDate  time.Time `json:"purchaseOrderInvoiceInvoiceDate"`
	PurchaseOrderInvoiceLastModified time.Time `json:"purchaseOrderInvoiceLastModified"`

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

	MedicineLists []PurchaseOrderItemsReturn `json:"medicineLists"`
}

type DeletePurchaseOrderInvoice struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseOrderInvoice struct {
	ID           int       `json:"id"`
	Number       int       `json:"number"`
	CompanyID    int       `json:"companyId"`
	SupplierID   int       `json:"supplierId"`
	CashierID    int       `json:"cashierId"`
	TotalItems   int       `json:"totalItems"`
	InvoiceDate  time.Time `json:"invoiceDate"`
	LastModified time.Time `json:"lastModified"`
	CreatedAt    time.Time `json:"createdAt"`
}

type PurchaseOrderItem struct {
	ID                     int     `json:"id"`
	PurchaseOrderInvoiceID int     `json:"purchaseOrderInvoiceId"`
	MedicineID             int     `json:"medicineId"`
	OrderQty               float64 `json:"orderQty"`
	ReceivedQty            float64 `json:"receivedQty"`
	UnitID                 int     `json:"unitId"`
	Remarks                string  `json:"remarks"`
}
