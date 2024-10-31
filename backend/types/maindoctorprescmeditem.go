package types

import "time"

type MainDoctorPrescMedItemStore interface {
	CreateMainDoctorPrescMedItem(item MainDoctorPrescMedItem) error
	GetMainDoctorPrescMedItemByMedicineData(medId int) (*MainDoctorPrescMedItemReturn, error)
	GetAllMainDoctorPrescMedItemByMedicineData() ([]MainDoctorPrescMedItemReturn, error)
	IsMedicineContentsExist(medId int) (bool, error)
	IsMedicineBarcodeExist(barcode string) (bool, error)

	DeleteMainDoctorPrescMedItem(medId int, user *User) error
}

type RegisterMainDoctorPrescMedItemPayload struct {
	MedicineName     string                      `json:"medicineName" validate:"required"`
	MedicineContents []MainDoctorPrescMedContent `json:"medicineContents" validate:"required"`
}

type ModifyMainDoctorPrescMedItemPayload struct {
	MedicineID          int                         `json:"medicineId" validate:"required"`
	NewMedicineContents []MainDoctorPrescMedContent `json:"newMedicineContents" validate:"required"`
}

type ViewMainDoctorPrescMedItemPayload struct {
	MedicineID          int                         `json:"medicineId" validate:"required"`
}

type MainDoctorPrescMedItemReturn struct {
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

type MainDoctorPrescMedItem struct {
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
