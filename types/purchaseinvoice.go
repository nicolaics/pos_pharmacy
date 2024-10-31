package types

import (
	"database/sql"
	"time"
)

type PurchaseInvoiceStore interface {
	GetPurchaseInvoicesByNumber(number int) ([]PurchaseInvoice, error)
	GetPurchaseInvoiceByID(int) (*PurchaseInvoice, error)
	GetPurchaseInvoiceID(number int, supplierId int, subtotal float64, totalPrice float64, invoiceDate time.Time) (int, error)
	GetPurchaseMedicineItem(purchaseInvoiceId int) ([]PurchaseMedicineItemReturn, error)

	GetPurchaseInvoicesByDate(startDate time.Time, endDate time.Time) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]PurchaseInvoiceListsReturnPayload, error)
	GetPurchaseInvoicesByDateAndPONumber(startDate time.Time, endDate time.Time, poiNumber int) ([]PurchaseInvoiceListsReturnPayload, error)

	CreatePurchaseInvoice(PurchaseInvoice) error
	CreatePurchaseMedicineItem(PurchaseMedicineItem) error

	DeletePurchaseInvoice(*PurchaseInvoice, *User) error
	DeletePurchaseMedicineItem(*PurchaseInvoice, *User) error

	ModifyPurchaseInvoice(int, PurchaseInvoice, *User) error

	// delete entirely from the db if there's error
	AbsoluteDeletePurchaseInvoice(pi PurchaseInvoice) error

	UpdatePDFUrl(piId int, pdfUrl string) error
	IsPDFUrlExist(pdfUrl string) (bool, error)
}

type PurchaseInvoicePayload struct {
	Number              int                           `json:"number" validate:"required"`
	SupplierID          int                           `json:"supplierId" validate:"required"`
	PurchaseOrderNumber int                           `json:"purchaseOrderNumber"`
	Subtotal            float64                       `json:"subtotal" validate:"required"`
	DiscountPercentage  float64                       `json:"discountPercentage"`
	DiscountAmount      float64                       `json:"discountAmount"`
	TaxPercentage       float64                       `json:"taxPercentage" validate:"required"`
	TaxAmount           float64                       `json:"taxAmount" validate:"required"`
	TotalPrice          float64                       `json:"totalPrice" validate:"required"`
	Description         string                        `json:"description"`
	InvoiceDate         string                        `json:"invoiceDate" validate:"required"`
	MedicineLists       []PurchaseMedicineListPayload `json:"purchaseMedicineList" validate:"required"`
}

type PurchaseMedicineListPayload struct {
	MedicineBarcode    string  `json:"medicineBarcode" validate:"required"`
	MedicineName       string  `json:"medicineName" validate:"required"`
	Qty                float64 `json:"qty" validate:"required"`
	Unit               string  `json:"unit" validate:"required"`
	Price              float64 `json:"price" validate:"required"`
	DiscountPercentage float64 `json:"discountPercentage"`
	DiscountAmount     float64 `json:"discountAmount"`
	TaxPercentage      float64 `json:"taxPercentage"`
	TaxAmount          float64 `json:"taxAmount"`
	Subtotal           float64 `json:"subtotal" validate:"required"`
	BatchNumber        string  `json:"batchNumber" validate:"required"`
	ExpDate            string  `json:"expDate" validate:"required"`
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
	ID                 int       `json:"id"`
	MedicineBarcode    string    `json:"medicineBarcode"`
	MedicineName       string    `json:"medicineName"`
	Qty                float64   `json:"qty"`
	Unit               string    `json:"unit"`
	Price              float64   `json:"price"`
	DiscountPercentage float64   `json:"discountPercentage"`
	DiscountAmount     float64   `json:"discountAmount"`
	TaxPercentage      float64   `json:"taxPercentage"`
	TaxAmount          float64   `json:"taxAmount"`
	Subtotal           float64   `json:"subtotal"`
	BatchNumber        string    `json:"batchNumber"`
	ExpDate            time.Time `json:"expDate"`
}

type PurchaseInvoiceDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	Subtotal               float64   `json:"subtotal"`
	DiscountPercentage     float64   `json:"discountPercentage"`
	DiscountAmount         float64   `json:"discountAmount"`
	TaxPercentage          float64   `json:"taxPercentage"`
	TaxAmount              float64   `json:"taxAmount"`
	TotalPrice             float64   `json:"totalPrice"`
	Description            string    `json:"description"`
	InvoiceDate            time.Time `json:"invoiceDate"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastModifiedByUserName"`
	PdfURL                 string    `json:"pdfUrl"`

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

	PurchaseOrderNumber int `json:"purchaseOrderNumber"`

	MedicineLists []PurchaseMedicineItemReturn `json:"medicineLists"`
}

// view the lists of the purchase invoice
type PurchaseInvoiceListsReturnPayload struct {
	ID                  int       `json:"id"`
	Number              int       `json:"number"`
	SupplierName        string    `json:"supplierName"`
	PurchaseOrderNumber int       `json:"purchaseOrderNumber"`
	TotalPrice          float64   `json:"totalPrice"`
	Description         string    `json:"description"`
	UserName            string    `json:"userName"`
	InvoiceDate         time.Time `json:"invoiceDate"`
	PdfURL              string    `json:"pdfUrl"`
}

type DeletePurchaseInvoice struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseInvoicePDFPayload struct {
	Number             int       `json:"number"`
	Subtotal           float64   `json:"subtotal"`
	DiscountPercentage float64   `json:"discountPercentage"`
	DiscountAmount     float64   `json:"discountAmount"`
	TaxPercentage      float64   `json:"taxPercentage"`
	TaxAmount          float64   `json:"taxAmount"`
	TotalPrice         float64   `json:"totalPrice"`
	Description        string    `json:"description"`
	InvoiceDate        time.Time `json:"invoiceDate"`

	Supplier struct {
		Name                string `json:"name"`
		Address             string `json:"address"`
		CompanyPhoneNumber  string `json:"companyPhoneNumber"`
		ContactPersonName   string `json:"contactPersonName"`
		ContactPersonNumber string `json:"contactPersonNumber"`
		Terms               string `json:"terms"`
		VendorIsTaxable     bool   `json:"vendorIsTaxable"`
	} `json:"supplier"`

	UserName string `json:"name"`

	PurchaseOrderNumber int       `json:"purchaseOrderNumber"`
	PurchaseOrderDate   time.Time `json:"purchaseOrderDate"`

	MedicineLists []PurchaseMedicineListPayload `json:"medicineLists"`
}

type PurchaseInvoice struct {
	ID                   int           `json:"id"`
	Number               int           `json:"number"`
	SupplierID           int           `json:"supplierId"`
	PurchaseOrderNumber  int           `json:"purchaseOrderNumber"`
	Subtotal             float64       `json:"subtotal"`
	DiscountPercentage   float64       `json:"discountPercentage"`
	DiscountAmount       float64       `json:"dicsountAmount"`
	TaxPercentage        float64       `json:"taxPercentage"`
	TaxAmount            float64       `json:"taxAmount"`
	TotalPrice           float64       `json:"totalPrice"`
	Description          string        `json:"description"`
	UserID               int           `json:"userId"`
	InvoiceDate          time.Time     `json:"invoiceDate"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastModifiedByUserId"`
	PdfURL               string        `json:"pdfUrl"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}

type PurchaseMedicineItem struct {
	ID                 int       `json:"id"`
	PurchaseInvoiceID  int       `json:"purchaseInvoiceId"`
	MedicineID         int       `json:"medicineId"`
	Qty                float64   `json:"qty"`
	UnitID             int       `json:"unitId"`
	Price              float64   `json:"price"`
	DiscountPercentage float64   `json:"discountPercentage"`
	DiscountAmount     float64   `json:"dicsountAmount"`
	TaxPercentage      float64   `json:"taxPercentage"`
	TaxAmount          float64   `json:"taxAmount"`
	Subtotal           float64   `json:"subtotal"`
	BatchNumber        string    `json:"batchNumber"`
	ExpDate            time.Time `json:"expDate"`
}
