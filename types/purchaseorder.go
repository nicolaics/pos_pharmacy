package types

import (
	"database/sql"
	"time"
)

type PurchaseOrderStore interface {
	GetPurchaseOrderByNumber(int) (*PurchaseOrder, error)
	GetPurchaseOrderByID(int) (*PurchaseOrder, error)
	GetPurchaseOrderID(number int, supplierId int, totalItem int, invoiceDate time.Time) (int, error)
	GetNumberOfPurchaseOrders() (int, error)

	CreatePurchaseOrder(PurchaseOrder) error
	CreatePurchaseOrderItem(PurchaseOrderItem) error

	GetPurchaseOrdersByDate(startDate time.Time, endDate time.Time) ([]PurchaseOrderListsReturnPayload, error)
	GetPurchaseOrdersByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]PurchaseOrderListsReturnPayload, error)
	GetPurchaseOrdersByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]PurchaseOrderListsReturnPayload, error)
	GetPurchaseOrdersByDateAndSupplierID(startDate time.Time, endDate time.Time, sid int) ([]PurchaseOrderListsReturnPayload, error)

	GetPurchaseOrderItem(purchaseOrderId int) ([]PurchaseOrderItemReturn, error)

	DeletePurchaseOrder(*PurchaseOrder, *User) error
	DeletePurchaseOrderItem(*PurchaseOrder, *User) error

	ModifyPurchaseOrder(int, PurchaseOrder, *User) error

	UpdtaeReceivedQty(poinid int, newQty float64, user *User, mid int) error

	// delete entirely from the db if there's error
	AbsoluteDeletePurchaseOrder(poi PurchaseOrder) error

	UpdatePDFUrl(poId int, pdfUrl string) error
	IsPDFUrlExist(pdfUrl string) (bool, error)
}

// SHOW COMPANY ID AND SUPPLIER ID AS WELL IN THE FRONT-END
type RegisterPurchaseOrderPayload struct {
	Number      int    `json:"number" validate:"required"`
	SupplierID  int    `json:"supplierId" validate:"required"`
	TotalItem   int    `json:"totalItem" validate:"required"`
	InvoiceDate string `json:"invoiceDate" validate:"required"`

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
type ViewPurchaseOrderPayload struct {
	StartDate string `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   string `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type ViewPurchaseOrderItemPayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyPurchaseOrderPayload struct {
	ID      int                          `json:"id" validate:"required"`
	NewData RegisterPurchaseOrderPayload `json:"newData" validate:"required"`
}

type PurchaseOrderItemReturn struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	OrderQty        float64 `json:"orderQty"`
	ReceivedQty     float64 `json:"receivedQty"`
	Unit            string  `json:"unit"`
	Remarks         string  `json:"remarks"`
}

type PurchaseOrderListsReturnPayload struct {
	ID           int       `json:"id"`
	Number       int       `json:"number"`
	SupplierName string    `json:"supplierName"`
	UserName     string    `json:"userName"`
	TotalItem    int       `json:"totalItem"`
	InvoiceDate  time.Time `json:"invoiceDate"`
	PdfURL              string    `json:"pdfUrl"`
}

type PurchaseOrderDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	TotalItem              int       `json:"totalItem"`
	InvoiceDate            time.Time `json:"invoiceDate"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastModifiedByUserName"`
	PdfURL              string    `json:"pdfUrl"`

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

	MedicineLists []PurchaseOrderItemReturn `json:"medicineLists"`
}

type DeletePurchaseOrder struct {
	ID int `json:"id" validate:"required"`
}

type PurchaseOrder struct {
	ID                   int           `json:"id"`
	Number               int           `json:"number"`
	SupplierID           int           `json:"supplierId"`
	UserID               int           `json:"userId"`
	TotalItem            int           `json:"totalItem"`
	InvoiceDate          time.Time     `json:"invoiceDate"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastModifiedByUserId"`
	PdfURL               string        `json:"pdfUrl"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}

type PurchaseOrderItem struct {
	ID              int     `json:"id"`
	PurchaseOrderID int     `json:"purchaseOrderId"`
	MedicineID      int     `json:"medicineId"`
	OrderQty        float64 `json:"orderQty"`
	ReceivedQty     float64 `json:"receivedQty"`
	UnitID          int     `json:"unitId"`
	Remarks         string  `json:"remarks"`
}
