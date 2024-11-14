package types

import (
	"database/sql"
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

type GetOneMedicinePayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyMedicinePayload struct {
	ID      int                     `json:"id" validate:"required"`
	NewData RegisterMedicinePayload `json:"newData" validate:"required"`
}

type MedicineListsReturnPayload struct {
	ID                         int       `json:"id"`
	Barcode                    string    `json:"barcode"`
	Name                       string    `json:"name"`
	Qty                        float64   `json:"qty"`
	FirstUnitName              string    `json:"firstUnitName"`
	FirstDiscountPercentage    float64   `json:"firstDiscountPercentage"`
	FirstDiscountAmount        float64   `json:"firstDiscountAmount"`
	FirstPrice                 float64   `json:"firstPrice"`
	SecondUnitName             string    `json:"secondUnitName"`
	SecondUnitToFirstUnitRatio float64   `json:"secondUnitToFirstUnitRatio"`
	SecondDiscountPercentage   float64   `json:"secondDiscountPercentage"`
	SecondDiscountAmount       float64   `json:"secondDiscountAmount"`
	SecondPrice                float64   `json:"secondPrice"`
	ThirdUnitName              string    `json:"thirdUnitName"`
	ThirdUnitToFirstUnitRatio  float64   `json:"thirdUnitToFirstUnitRatio"`
	ThirdDiscountPercentage    float64   `json:"thirdDiscountPercentage"`
	ThirdDiscountAmount        float64   `json:"thirdDiscountAmount"`
	ThirdPrice                 float64   `json:"thirdPrice"`
	Description                string    `json:"description"`
	CreatedAt                  time.Time `json:"createdAt"`
	LastModified               time.Time `json:"lastModified"`
	LastModifiedByUserName     string    `json:"lastModifiedByUserName"`
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
