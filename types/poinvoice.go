package types

import (
	"database/sql"
	"time"
)

type PurchaseOrderInvoiceStore interface {
	GetPurchaseOrderInvoicesByNumber(int) ([]PurchaseOrderInvoice, error)
	GetPurchaseOrderInvoiceByID(int) (*PurchaseOrderInvoice, error)
	GetPurchaseOrderInvoiceID(number int, companyId int, supplierId int, userId int, totalItems int, invoiceDate time.Time) (int, error)
	GetNumberOfPurchaseOrderInvoices() (int, error)

	CreatePurchaseOrderInvoice(PurchaseOrderInvoice) error
	CreatePurchaseOrderItems(PurchaseOrderItem) error

	GetPurchaseOrderInvoicesByDate(startDate time.Time, endDate time.Time) ([]PurchaseOrderInvoiceListsReturnPayload, error)
	GetPurchaseOrderInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]PurchaseOrderInvoiceListsReturnPayload, error)
	GetPurchaseOrderInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]PurchaseOrderInvoiceListsReturnPayload, error)
	GetPurchaseOrderInvoicesByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]PurchaseOrderInvoiceListsReturnPayload, error)

	GetPurchaseOrderItems(purchaseOrderInvoiceId int) ([]PurchaseOrderItemsReturn, error)

	DeletePurchaseOrderInvoice(*PurchaseOrderInvoice, *User) error
	DeletePurchaseOrderItems(*PurchaseOrderInvoice, *User) error

	ModifyPurchaseOrderInvoice(int, PurchaseOrderInvoice, *User) error
}

// SHOW COMPANY ID AND SUPPLIER ID AS WELL IN THE FRONT-END
type RegisterPurchaseOrderInvoicePayload struct {
	Number      int       `json:"number" validate:"required"`
	CompanyID   int       `json:"companyId" validate:"required"`
	SupplierID  int       `json:"supplierId" validate:"required"`
	TotalItems  int       `json:"totalItems" validate:"required"`
	InvoiceDate time.Time `json:"invoiceDate" validate:"required"`

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
type ViewPurchaseOrderInvoicePayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type ViewPurchaseOrderItemsPayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyPurchaseOrderInvoicePayload struct {
	ID      int                                 `json:"id" validate:"required"`
	NewData RegisterPurchaseOrderInvoicePayload `json:"newData" validate:"required"`
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

type PurchaseOrderInvoiceListsReturnPayload struct {
	ID           int       `json:"id"`
	Number       int       `json:"number"`
	SupplierName string    `json:"supplierName"`
	UserName     string    `json:"userName"`
	TotalItems   int       `json:"totalItems"`
	InvoiceDate  time.Time `json:"invoiceDate"`
}

type PurchaseOrderInvoiceDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	TotalItems             int       `json:"totalItems"`
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

	MedicineLists []PurchaseOrderItemsReturn `json:"medicineLists"`
}

type DeletePurchaseOrderInvoice struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseOrderInvoice struct {
	ID                   int           `json:"id"`
	Number               int           `json:"number"`
	CompanyID            int           `json:"companyId"`
	SupplierID           int           `json:"supplierId"`
	UserID               int           `json:"userId"`
	TotalItems           int           `json:"totalItems"`
	InvoiceDate          time.Time     `json:"invoiceDate"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastModifiedByUserId"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
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
