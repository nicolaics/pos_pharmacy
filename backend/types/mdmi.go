package types

import "time"

type MainDoctorMedItemStore interface {
	CreateMainDoctorMedItem(item MainDoctorMedItem) error
	GetMainDoctorMedItemByMedicineData(medId int) (*MainDoctorMedItemReturn, error)
	GetAllMainDoctorMedItemByMedicineData() ([]MainDoctorMedItemReturn, error)
	IsMedicineContentsExist(medId int) (bool, error)
	IsMedicineBarcodeExist(barcode string) (bool, error)

	DeleteMainDoctorMedItem(medId int, user *User) error
}

type RegisterMainDoctorMedItemPayload struct {
	MedicineName     string                      `json:"medicineName" validate:"required"`
	MedicineContents []MainDoctorPrescMedContent `json:"medicineContents" validate:"required"`
}

type ModifyMainDoctorMedItemPayload struct {
	MedicineID          int                         `json:"medicineId" validate:"required"`
	NewMedicineContents []MainDoctorPrescMedContent `json:"newMedicineContents" validate:"required"`
}

type ViewMainDoctorMedItemPayload struct {
	MedicineID          int                         `json:"medicineId" validate:"required"`
}

type MainDoctorMedItemReturn struct {
	MedicineName           string                      `json:"medicine"`
	MedicineContents       []MainDoctorPrescMedContent `json:"medicineContents"`
	LastModified           time.Time                   `json:"lastModified"`
	LastModifiedByUserName string                      `json:"lastModifiedByUserName"`
}

type MainDoctorPrescMedContent struct {
	Name string `json:"name"`
	Qty  string `json:"qty"`
	Unit string `json:"unit"`
}

type MainDoctorMedItem struct {
	ID                   int       `json:"id"`
	MedicineID           int       `json:"medicineId"`
	MedicineContentID    int       `json:"medicineContentId"`
	Qty                  float64   `json:"qty"`
	UnitID               int       `json:"unitId"`
	UserID               int       `json:"userId"`
	CreatedAt            time.Time `json:"createdAt"`
	LastModified         time.Time `json:"lastModified"`
	LastModifiedByUserID int       `json:"lastModifiedByUserId"`
}
