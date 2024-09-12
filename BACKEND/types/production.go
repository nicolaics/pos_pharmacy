package types

import (
	"database/sql"
	"time"
)

type ProductionStore interface {
	GetProductionByBatchNumber(int) (*Production, error)
	GetProductionByID(int) (*Production, error)
	GetProductionsByDate(startDate time.Time, endDate time.Time) ([]Production, error)
	GetProductionID(batchNumber int, producedMedId int, prodDate time.Time, totalCost float64, userId int) (int, error)
	GetNumberOfProductions() (int, error)

	CreateProduction(Production) error
	CreateProductionMedicineItems(ProductionMedicineItems) error

	GetProductionMedicineItems(prescriptionId int) ([]ProductionMedicineItemRow, error)
	DeleteProduction(*Production, int) error
	DeleteProductionMedicineItems(*Production, int) error
	ModifyProduction(int, Production, int) error
}

type RegisterProductionPayload struct {
	BatchNumber             int       `json:"batchNumber"`
	ProducedMedicineBarcode string    `json:"producedMedicineBarcode" validate:"required"`
	ProducedMedicineName    string    `json:"producedMedicineName" validate:"required"`
	ProducedQty             int       `json:"producedQty" validate:"required"`
	ProductionDate          time.Time `json:"productionDate" validate:"required"`
	Description             string    `json:"description"`
	UpdatedToStock          bool      `json:"updatedToStock" validate:"required"`
	UpdatedToAccount        bool      `json:"updatedToAccount" validate:"required"`
	TotalCost               float64   `json:"totalCost" validate:"required"`

	MedicineLists []ProductionMedicineListPayload `json:"productionMedicineList" validate:"required"`
}

type ProductionMedicineListPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Cost            float64 `json:"cost" validate:"required"`
}

// only view the production list
type ViewProductionsPayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the production
type ViewProductionMedicineItemsPayload struct {
	BatchNumber int `json:"batchNumber" validate:"required"`
}

type ModifyProductionPayload struct {
	ID      int                       `json:"id" validate:"required"`
	NewData RegisterProductionPayload `json:"newData" validate:"required"`
}

// data of the medicine per row in the prescription
type ProductionMedicineItemRow struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	Qty             float64 `json:"qty"`
	Unit            string  `json:"unit"`
	Cost            float64 `json:"cost"`
}

// data to be sent back to the client after clicking 1 prescription
type ProductionDetailPayload struct {
	ID          int `json:"id"`
	BatchNumber int `json:"batchNumber"`

	ProducedMedicine struct {
		Barcode string `json:"barcode"`
		Name    string `json:"name"`
	} `json:"producedMedicine"`

	ProducedQty      int       `json:"producedQty"`
	ProductionDate   time.Time `json:"productionDate"`
	Description      string    `json:"description"`
	UpdatedToStock   bool      `json:"updatedToStock"`
	UpdatedToAccount bool      `json:"updatedToAccount"`
	TotalCost        float64   `json:"totalCost"`

	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastLastModifiedByUserName"`

	MedicineLists []ProductionMedicineItemRow `json:"medicineLists"`
}

type DeleteProduction struct {
	ID int `json:"id" validate:"required"`
}

type ProductionMedicineItems struct {
	ID           int     `json:"id"`
	ProductionID int     `json:"prescriptionId"`
	MedicineID   int     `json:"medicineId"`
	Qty          float64 `json:"qty"`
	UnitID       int     `json:"unitId"`
	Cost         float64 `json:"cost"`
}

type Production struct {
	ID                   int           `json:"id"`
	BatchNumber          int           `json:"batchNumber"`
	ProducedMedicineID   int           `json:"producedMedicineId"`
	ProducedQty          int           `json:"producedQty"`
	ProductionDate       time.Time     `json:"productionDate"`
	Description          string        `json:"description"`
	UpdatedToStock       bool          `json:"updatedToStock"`
	UpdatedToAccount     bool          `json:"updatedToAccount"`
	TotalCost            float64       `json:"totalCost"`
	UserID               int           `json:"userId"`
	CreatedAt            time.Time     `json:"createdAt"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastLastModifiedByUserId"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}
