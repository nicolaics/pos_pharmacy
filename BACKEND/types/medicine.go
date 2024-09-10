package types

import (
	"time"
)

type RegisterMedicinePayload struct {
	Barcode        string  `json:"barcode" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Qty            float64 `json:"qty" validate:"required"`
	FirstUnit      string  `json:"firstUnit" validate:"required"`
	FirstSubtotal  float64 `json:"firstSubtotal" validate:"required"`
	FirstDiscount  float64 `json:"firstDiscount"`
	FirstPrice     float64 `json:"firstPrice" validate:"required"`
	SecondUnit     string  `json:"secondUnit"`
	SecondSubtotal float64 `json:"secondSubtotal"`
	SecondDiscount float64 `json:"secondDiscount"`
	SecondPrice    float64 `json:"secondPrice"`
	ThirdUnit      string  `json:"thirdUnit"`
	ThirdSubtotal  float64 `json:"thirdSubtotal"`
	ThirdDiscount  float64 `json:"thirdDiscount"`
	ThirdPrice     float64 `json:"thirdPrice"`
	Description    string  `json:"description"`
}

type DeleteMedicinePayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetOneMedicinePayload struct {
	ID   int    `json:"id" validate:"required"`
}

type ModifyMedicinePayload struct {
	ID      int                     `json:"id" validate:"required"`
	NewData RegisterMedicinePayload `json:"newData" validate:"required"`
}

type MedicineStore interface {
	GetMedicineByName(string) (*Medicine, error)
	GetMedicineByID(int) (*Medicine, error)
	GetMedicineByBarcode(string) (*Medicine, error)
	CreateMedicine(Medicine, int) error
	GetAllMedicines() ([]Medicine, error)
	DeleteMedicine(*Medicine, int) error
	ModifyMedicine(int, Medicine, int) error
}

type Medicine struct {
	ID                   int       `json:"id"`
	Barcode              string    `json:"barcode"`
	Name                 string    `json:"name"`
	Qty                  float64   `json:"qty"`
	FirstUnitID          int       `json:"firstUnitId"`
	FirstSubtotal        float64   `json:"firstSubtotal"`
	FirstDiscount        float64   `json:"firstDiscount"`
	FirstPrice           float64   `json:"firstPrice"`
	SecondUnitID         int       `json:"secondUnitId"`
	SecondSubtotal       float64   `json:"secondSubtotal"`
	SecondDiscount       float64   `json:"secondDiscount"`
	SecondPrice          float64   `json:"secondPrice"`
	ThirdUnitID          int       `json:"thirdUnitId"`
	ThirdSubtotal        float64   `json:"thirdSubtotal"`
	ThirdDiscount        float64   `json:"thirdDiscount"`
	ThirdPrice           float64   `json:"thirdPrice"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"createdAt"`
	LastModified         time.Time `json:"lastModified"`
	LastModifiedByUserID int       `json:"lastModifiedByUserId"`
	DeletedAt            time.Time `json:"deletedAt"`
	DeletedByUserID      int       `json:"deletedByUserId"`
}
