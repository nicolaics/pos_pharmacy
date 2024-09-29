package types

import (
	"database/sql"
	"time"
)

type ProductionStore interface {
	GetProductionByNumber(int) (*Production, error)
	GetProductionByID(int) (*Production, error)
	GetNumberOfProductions() (int, error)

	GetProductionsByDate(startDate time.Time, endDate time.Time) ([]ProductionListsReturnPayload, error)
	GetProductionsByDateAndNumber(startDate time.Time, endDate time.Time, bn int) ([]ProductionListsReturnPayload, error)
	GetProductionsByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]ProductionListsReturnPayload, error)
	GetProductionsByDateAndMedicineID(startDate time.Time, endDate time.Time, mid int) ([]ProductionListsReturnPayload, error)
	GetProductionsByDateAndUpdatedToStock(startDate time.Time, endDate time.Time, uts bool) ([]ProductionListsReturnPayload, error)
	GetProductionsByDateAndUpdatedToAccount(startDate time.Time, endDate time.Time, uta bool) ([]ProductionListsReturnPayload, error)

	CreateProduction(Production) error
	CreateProductionMedicineItems(ProductionMedicineItems) error

	GetProductionMedicineItems(prescriptionId int) ([]ProductionMedicineItemRow, error)
	DeleteProduction(*Production, *User) error
	DeleteProductionMedicineItems(*Production, *User) error
	ModifyProduction(int, Production, *User) error
}

type RegisterProductionPayload struct {
	Number                  int     `json:"number"`
	ProducedMedicineBarcode string  `json:"producedMedicineBarcode" validate:"required"`
	ProducedMedicineName    string  `json:"producedMedicineName" validate:"required"`
	ProducedQty             int     `json:"producedQty" validate:"required"`
	ProducedUnit            string  `json:"producedUnit" validate:"required"`
	ProductionDate          string  `json:"productionDate" validate:"required"`
	Description             string  `json:"description"`
	UpdatedToStock          bool    `json:"updatedToStock"`
	UpdatedToAccount        bool    `json:"updatedToAccount"`
	TotalCost               float64 `json:"totalCost" validate:"required"`

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
	StartDate string `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   string `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the production
type ViewProductionMedicineItemsPayload struct {
	Number int `json:"number" validate:"required"`
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
	ID     int `json:"id"`
	Number int `json:"number"`

	ProducedMedicine struct {
		Barcode string `json:"barcode"`
		Name    string `json:"name"`
	} `json:"producedMedicine"`

	ProducedQty      int       `json:"producedQty"`
	ProducedUnit     string    `json:"producedUnit"`
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

// return to the user when viewing the lists of production
type ProductionListsReturnPayload struct {
	ID                   int       `json:"id"`
	Number               int       `json:"number"`
	ProducedMedicineName string    `json:"producedMedicineName"`
	ProducedQty          int       `json:"producedQty"`
	ProducedUnit         string    `json:"producedUnit"`
	ProductionDate       time.Time `json:"productionDate"`
	Description          string    `json:"description"`
	UpdatedToStock       bool      `json:"updatedToStock"`
	UpdatedToAccount     bool      `json:"updatedToAccount"`
	TotalCost            float64   `json:"totalCost"`
	UserName             string    `json:"userName"`
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
	Number               int           `json:"number"`
	ProducedMedicineID   int           `json:"producedMedicineId"`
	ProducedQty          int           `json:"producedQty"`
	ProducedUnitID       int           `json:"producedUnitId"`
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
