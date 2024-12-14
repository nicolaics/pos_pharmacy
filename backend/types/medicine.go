package types

import (
	"database/sql"
	"encoding/xml"
	"time"
)

type MedicineStore interface {
	GetMedicineByName(string) (*Medicine, error)
	GetMedicineByID(int) (*MedicineListsReturnPayload, error)
	GetMedicineByBarcode(string) (*Medicine, error)

	GetMedicinesBySearchName(string) ([]MedicineListsReturnPayload, error)
	GetMedicinesBySearchBarcode(string) ([]MedicineListsReturnPayload, error)
	GetMedicinesByDescription(string) ([]MedicineListsReturnPayload, error)

	CreateMedicine(Medicine, int) error

	GetAllMedicines() ([]MedicineListsReturnPayload, error)

	DeleteMedicine(*Medicine, *User) error

	ModifyMedicine(int, Medicine, *User) error

	UpdateMedicineStock(mid int, newStock float64, user *User) error

	InsertIntoMedicineHistoryTable(mid, invoiceId, historyType int, qty float64, unitId int, invoiceDate time.Time) error
	ModifyMedicineHistoryTable(mid, invoiceId, historyType int, qty float64, unitId int, invoiceDate time.Time) error
	DeleteMedicineHistory(mid, invoiceId, historyType int, qty float64, user *User) error

	GetMedicineHistoryByInvoiceIdAndQty(mid, invoiceId, historyType int, qty float64) (*MedicineHistory, error)
	GetMedicineHistoryByDate(mid int, startDate time.Time, endDate time.Time) ([]MedicineHistoryReturn, error)
}

type RegisterMedicinePayload struct {
	Barcode                    string  `json:"barcode" validate:"required"`
	Name                       string  `json:"name" validate:"required"`
	Qty                        float64 `json:"qty" validate:"required"`
	FirstUnit                  string  `json:"firstUnit" validate:"required"`
	FirstSubtotal              float64 `json:"firstSubtotal" validate:"required"`
	FirstDiscountPercentage    float64 `json:"firstDiscountPercentage"`
	FirstDiscountAmount        float64 `json:"firstDiscountAmount"`
	FirstPrice                 float64 `json:"firstPrice" validate:"required"`
	SecondUnit                 string  `json:"secondUnit"`
	SecondUnitToFirstUnitRatio float64 `json:"secondUnitToFirstUnitRatio"`
	SecondSubtotal             float64 `json:"secondSubtotal"`
	SecondDiscountPercentage   float64 `json:"secondDiscountPercentage"`
	SecondDiscountAmount       float64 `json:"secondDiscountAmount"`
	SecondPrice                float64 `json:"secondPrice"`
	ThirdUnit                  string  `json:"thirdUnit"`
	ThirdUnitToFirstUnitRatio  float64 `json:"thirdUnitToFirstUnitRatio"`
	ThirdSubtotal              float64 `json:"thirdSubtotal"`
	ThirdDiscountPercentage    float64 `json:"thirdDiscountPercentage"`
	ThirdDiscountAmount        float64 `json:"thirdDiscountAmount"`
	ThirdPrice                 float64 `json:"thirdPrice"`
	Description                string  `json:"description"`
}

type DeleteMedicinePayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetDetailMedicinePayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyMedicinePayload struct {
	ID      int                     `json:"id" validate:"required"`
	NewData RegisterMedicinePayload `json:"newData" validate:"required"`
}

type RequestExportMedicinePayload struct {
	// Fields                     []string `json:"fields"`
	ID                         bool `json:"id" csv:"id" xml:"id,attr"`
	Barcode                    bool `json:"barcode" csv:"barcode" xml:"barcode"`
	Name                       bool `json:"name" csv:"name" xml:"name"`
	Qty                        bool `json:"qty" csv:"qty" xml:"qty"`
	FirstUnitName              bool `json:"firstUnitName" csv:"first_unit_name" xml:"first_unit>name"`
	FirstDiscountPercentage    bool `json:"firstDiscountPercentage" csv:"first_discount_percentage" xml:"first_unit>discount>percentage"`
	FirstDiscountAmount        bool `json:"firstDiscountAmount" csv:"first_discount_amount" xml:"first_unit>discount>amount"`
	FirstPrice                 bool `json:"firstPrice" csv:"first_price" xml:"first_unit>price"`
	SecondUnitName             bool `json:"secondUnitName" csv:"second_unit_name" xml:"second_unit>name,omitempty"`
	SecondUnitToFirstUnitRatio bool `json:"secondUnitToFirstUnitRatio" csv:"second_unit_to_first_unit_ratio" xml:"second_unit>first_unit_ratio,omitempty"`
	SecondDiscountPercentage   bool `json:"secondDiscountPercentage" csv:"second_discount_percentage" xml:"second_unit>discount>percentage,omitempty"`
	SecondDiscountAmount       bool `json:"secondDiscountAmount" csv:"second_discount_amount" xml:"second_unit>discount>amount,omitempty"`
	SecondPrice                bool `json:"secondPrice" csv:"second_price" xml:"second_unit>price,omitempty"`
	ThirdUnitName              bool `json:"thirdUnitName" csv:"third_unit_name" xml:"third_unit>name,omitempty"`
	ThirdUnitToFirstUnitRatio  bool `json:"thirdUnitToFirstUnitRatio" csv:"third_unit_to_first_unit_ratio" xml:"third_unit>first_unit_ratio,omitempty"`
	ThirdDiscountPercentage    bool `json:"thirdDiscountPercentage" csv:"third_discount_percentage" xml:"third_unit>discount>percentage,omitempty"`
	ThirdDiscountAmount        bool `json:"thirdDiscountAmount" csv:"third_discount_amount" xml:"third_unit>discount>amount,omitempty"`
	ThirdPrice                 bool `json:"thirdPrice" csv:"third_price" xml:"third_unit>price,omitempty"`
	Description                bool `json:"description" csv:"description" xml:"description"`
	CreatedAt                  bool `json:"createdAt" csv:"created_at" xml:"created_at"`
	LastModified               bool `json:"lastModified" csv:"last_modified" xml:"last_modified>at"`
	LastModifiedByUserName     bool `json:"lastModifiedByUserName" csv:"last_modified_by_user_name" xml:"last_modified>by_user_name"`
}

type MedicineListsReturnPayload struct {
	ID                         int       `json:"id" csv:"id" xml:"id,attr"`
	Barcode                    string    `json:"barcode" csv:"barcode" xml:"barcode"`
	Name                       string    `json:"name" csv:"name" xml:"name"`
	Qty                        float64   `json:"qty" csv:"qty" xml:"qty"`
	FirstUnitName              string    `json:"firstUnitName" csv:"first_unit_name" xml:"first_unit>name"`
	FirstDiscountPercentage    float64   `json:"firstDiscountPercentage" csv:"first_discount_percentage" xml:"first_unit>discount>percentage"`
	FirstDiscountAmount        float64   `json:"firstDiscountAmount" csv:"first_discount_amount" xml:"first_unit>discount>amount"`
	FirstPrice                 float64   `json:"firstPrice" csv:"first_price" xml:"first_unit>price"`
	SecondUnitName             string    `json:"secondUnitName" csv:"second_unit_name" xml:"second_unit>name,omitempty"`
	SecondUnitToFirstUnitRatio float64   `json:"secondUnitToFirstUnitRatio" csv:"second_unit_to_first_unit_ratio" xml:"second_unit>first_unit_ratio,omitempty"`
	SecondDiscountPercentage   float64   `json:"secondDiscountPercentage" csv:"second_discount_percentage" xml:"second_unit>discount>percentage,omitempty"`
	SecondDiscountAmount       float64   `json:"secondDiscountAmount" csv:"second_discount_amount" xml:"second_unit>discount>amount,omitempty"`
	SecondPrice                float64   `json:"secondPrice" csv:"second_price" xml:"second_unit>price,omitempty"`
	ThirdUnitName              string    `json:"thirdUnitName" csv:"third_unit_name" xml:"third_unit>name,omitempty"`
	ThirdUnitToFirstUnitRatio  float64   `json:"thirdUnitToFirstUnitRatio" csv:"third_unit_to_first_unit_ratio" xml:"third_unit>first_unit_ratio,omitempty"`
	ThirdDiscountPercentage    float64   `json:"thirdDiscountPercentage" csv:"third_discount_percentage" xml:"third_unit>discount>percentage,omitempty"`
	ThirdDiscountAmount        float64   `json:"thirdDiscountAmount" csv:"third_discount_amount" xml:"third_unit>discount>amount,omitempty"`
	ThirdPrice                 float64   `json:"thirdPrice" csv:"third_price" xml:"third_unit>price,omitempty"`
	Description                string    `json:"description" csv:"description" xml:"description"`
	CreatedAt                  time.Time `json:"createdAt" csv:"created_at" xml:"created_at"`
	LastModified               time.Time `json:"lastModified" csv:"last_modified" xml:"last_modified>at"`
	LastModifiedByUserName     string    `json:"lastModifiedByUserName" csv:"last_modified_by_user_name" xml:"last_modified>by_user_name"`
}

type MedicineXmlPayload struct {
	XMLName                    xml.Name  `xml:"medicine"`
	ID                         int       `json:"id" xml:"id,attr"`
	Barcode                    string    `json:"barcode" xml:"barcode"`
	Name                       string    `json:"name" xml:"name"`
	Qty                        float64   `json:"qty" xml:"qty"`
	FirstUnitName              string    `json:"firstUnitName" xml:"first_unit>name"`
	FirstDiscountPercentage    float64   `json:"firstDiscountPercentage" xml:"first_unit>discount>percentage"`
	FirstDiscountAmount        float64   `json:"firstDiscountAmount" xml:"first_unit>discount>amount"`
	FirstPrice                 float64   `json:"firstPrice" xml:"first_unit>price"`
	SecondUnitName             string    `json:"secondUnitName" xml:"second_unit>name,omitempty"`
	SecondUnitToFirstUnitRatio float64   `json:"secondUnitToFirstUnitRatio" xml:"second_unit>first_unit_ratio,omitempty"`
	SecondDiscountPercentage   float64   `json:"secondDiscountPercentage" xml:"second_unit>discount>percentage,omitempty"`
	SecondDiscountAmount       float64   `json:"secondDiscountAmount" csv:"second_discount_amount" xml:"second_unit_>discount>amount,omitempty"`
	SecondPrice                float64   `json:"secondPrice" xml:"second_unit>price,omitempty"`
	ThirdUnitName              string    `json:"thirdUnitName" xml:"third_unit>name,omitempty"`
	ThirdUnitToFirstUnitRatio  float64   `json:"thirdUnitToFirstUnitRatio" xml:"third_unit>first_unit_ratio,omitempty"`
	ThirdDiscountPercentage    float64   `json:"thirdDiscountPercentage" xml:"third_unit>discount>percentage,omitempty"`
	ThirdDiscountAmount        float64   `json:"thirdDiscountAmount" xml:"third_unit>discount>amount,omitempty"`
	ThirdPrice                 float64   `json:"thirdPrice" xml:"third_unit>price,omitempty"`
	Description                string    `json:"description" xml:"description"`
	CreatedAt                  time.Time `json:"createdAt" xml:"created_at"`
	LastModified               time.Time `json:"lastModified" xml:"last_modified>at"`
	LastModifiedByUserName     string    `json:"lastModifiedByUserName" xml:"last_modified>by_user_name"`
}

type Medicine struct {
	ID                         int           `json:"id"`
	Barcode                    string        `json:"barcode"`
	Name                       string        `json:"name"`
	Qty                        float64       `json:"qty"`
	FirstUnitID                int           `json:"firstUnitId"`
	FirstSubtotal              float64       `json:"firstSubtotal"`
	FirstDiscountPercentage    float64       `json:"firstDiscountPercentage"`
	FirstDiscountAmount        float64       `json:"firstDiscountAmount"`
	FirstPrice                 float64       `json:"firstPrice"`
	SecondUnitID               int           `json:"secondUnitId"`
	SecondUnitToFirstUnitRatio float64       `json:"secondUnitToFirstUnitRatio"`
	SecondSubtotal             float64       `json:"secondSubtotal"`
	SecondDiscountPercentage   float64       `json:"secondDiscountPercentage"`
	SecondDiscountAmount       float64       `json:"secondDiscountAmount"`
	SecondPrice                float64       `json:"secondPrice"`
	ThirdUnitID                int           `json:"thirdUnitId"`
	ThirdUnitToFirstUnitRatio  float64       `json:"thirdUnitToFirstUnitRatio"`
	ThirdSubtotal              float64       `json:"thirdSubtotal"`
	ThirdDiscountPercentage    float64       `json:"thirdDiscountPercentage"`
	ThirdDiscountAmount        float64       `json:"thirdDiscountAmount"`
	ThirdPrice                 float64       `json:"thirdPrice"`
	Description                string        `json:"description"`
	CreatedAt                  time.Time     `json:"createdAt"`
	LastModified               time.Time     `json:"lastModified"`
	LastModifiedByUserID       int           `json:"lastModifiedByUserId"`
	DeletedAt                  sql.NullTime  `json:"deletedAt"`
	DeletedByUserID            sql.NullInt64 `json:"deletedByUserId"`
}

type GetMedicineHistoryPayload struct {
	ID        int       `json:"id" validate:"required"`
	StartDate time.Time `json:"startDate" validate:"required"`
	EndDate   time.Time `json:"endDate" validate:"required"`
}

type MedicineHistoryReturn struct {
	ID                int       `json:"id"`
	MedicineID        int       `json:"medicineId"`
	Qty               float64   `json:"qty"`
	UnitID            int       `json:"unitId"`
	InvoiceID         int       `json:"invoiceId"`
	PurchaseInvoiceID int       `json:"purchaseInvoiceId"`
	HistoryType       int       `json:"historyType"`
	InvoiceDate       time.Time `json:"invoiceDate"`
	LastModified      time.Time `json:"lastModified"`
}

type MedicineHistory struct {
	ID                int           `json:"id"`
	MedicineID        int           `json:"medicineId"`
	Qty               float64       `json:"qty"`
	UnitID            int           `json:"unitId"`
	InvoiceID         sql.NullInt64 `json:"invoiceId"`
	PurchaseInvoiceID sql.NullInt64 `json:"purchaseInvoiceId"`
	HistoryType       int           `json:"historyType"`
	InvoiceDate       time.Time     `json:"invoiceDate"`
	LastModified      time.Time     `json:"lastModified"`
}
