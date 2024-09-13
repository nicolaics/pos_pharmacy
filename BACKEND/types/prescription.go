package types

import (
	"database/sql"
	"time"
)

type PrescriptionStore interface {
	GetPrescriptionsByNumber(int) ([]Prescription, error)
	GetPrescriptionByID(int) (*Prescription, error)

	GetPrescriptionsByDate(startDate time.Time, endDate time.Time) ([]PrescriptionListsReturnPayload, error)
	GetPrescriptionsByDateAndNumber(startDate time.Time, endDate time.Time, number int) ([]PrescriptionListsReturnPayload, error)
	GetPrescriptionsByDateAndUserID(startDate time.Time, endDate time.Time, uid int) ([]PrescriptionListsReturnPayload, error)
	GetPrescriptionsByDateAndPatientID(startDate time.Time, endDate time.Time, pid int) ([]PrescriptionListsReturnPayload, error)
	GetPrescriptionsByDateAndDoctorID(startDate time.Time, endDate time.Time, did int) ([]PrescriptionListsReturnPayload, error)
	GetPrescriptionsByDateAndInvoiceID(startDate time.Time, endDate time.Time, iid int) ([]PrescriptionListsReturnPayload, error)

	GetPrescriptionID(invoiceId int, number int, date time.Time, patientName string, totalPrice float64) (int, error)

	CreatePrescription(Prescription) error
	CreatePrescriptionMedicineItems(PrescriptionMedicineItems) error

	GetPrescriptionMedicineItems(prescriptionId int) ([]PrescriptionMedicineItemRow, error)
	DeletePrescription(*Prescription, int) error
	DeletePrescriptionMedicineItems(*Prescription, int) error
	ModifyPrescription(int, Prescription, int) error
}

type RegisterPrescriptionPayload struct {
	Invoice struct {
		Number       int       `json:"number" validate:"required"`
		UserName     string    `json:"userName" validate:"required"`
		CustomerName string    `json:"customerName" validate:"required"`
		TotalPrice   float64   `json:"totalPrice" validate:"required"`
		InvoiceDate  time.Time `json:"invoiceDate" validate:"required"`
	} `json:"invoice" validate:"required"`

	Number           int                               `json:"number" validate:"required"`
	PrescriptionDate time.Time                         `json:"prescriptionDate" validate:"required"`
	PatientName      string                            `json:"patientName" validate:"required"`
	DoctorName       string                            `json:"doctorName" validate:"required"`
	Qty              float64                           `json:"qty" validate:"required"`
	Price            float64                           `json:"price" validate:"required"`
	TotalPrice       float64                           `json:"totalPrice" validate:"required"`
	Description      string                            `json:"description"`
	MedicineLists    []PrescriptionMedicineListPayload `json:"prescriptionMedicineList" validate:"required"`
}

type PrescriptionMedicineListPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             float64 `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
}

// prescription list payload returned to user after searching
type PrescriptionListsReturnPayload struct {
	ID                   int           `json:"id"`
	Number               int           `json:"number"`
	PrescriptionDate     time.Time     `json:"prescriptionDate"`
	PatientName          string        `json:"patientName"`
	DoctorName           string        `json:"doctorName"`
	Qty                  float64       `json:"qty"`
	Price                float64       `json:"price"`
	TotalPrice           float64       `json:"totalPrice"`
	Description          string        `json:"description"`
	UserName             string        `json:"userName"`

	Invoice struct {
		Number       int       `json:"number"`
		CustomerName string    `json:"customerName"`
		TotalPrice   float64   `json:"totalPrice"`
		InvoiceDate  time.Time `json:"invoiceDate"`
	} `json:"invoice"`
}

// only view the purchase invoice list
type ViewPrescriptionsPayload struct {
	StartDate time.Time `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   time.Time `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the purchase invoice
type ViewPrescriptionMedicineItemsPayload struct {
	PrescriptionID int `json:"prescriptionId" validate:"required"`
}

type ModifyPrescriptionPayload struct {
	ID      int                         `json:"id" validate:"required"`
	NewData RegisterPrescriptionPayload `json:"newData" validate:"required"`
}

// data of the medicine per row in the prescription
type PrescriptionMedicineItemRow struct {
	ID              int     `json:"id"`
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	Qty             float64 `json:"qty"`
	Unit            string  `json:"unit"`
	Price           float64 `json:"price"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal"`
}

// data to be sent back to the client after clicking 1 prescription
type PrescriptionDetailPayload struct {
	ID                     int       `json:"id"`
	Number                 int       `json:"number"`
	PrescriptionDate       time.Time `json:"prescriptionDate"`
	Qty                    float64   `json:"qty"`
	Price                  float64   `json:"price"`
	TotalPrice             float64   `json:"totalPrice"`
	Description            string    `json:"description"`
	CreatedAt              time.Time `json:"createdAt"`
	LastModified           time.Time `json:"lastModified"`
	LastModifiedByUserName string    `json:"lastLastModifiedByUserName"`

	Invoice struct {
		Number       int       `json:"number"`
		CustomerName string    `json:"customerName"`
		TotalPrice   float64   `json:"totalPrice"`
		InvoiceDate  time.Time `json:"invoiceDate"`
	} `json:"invoice"`

	Patient struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"patient"`

	Doctor struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"doctor"`

	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	MedicineLists []PrescriptionMedicineItemRow `json:"medicineLists"`
}

type DeletePrescription struct {
	ID int `json:"id" validate:"required"`
}

type PrescriptionMedicineItems struct {
	ID             int     `json:"id"`
	PrescriptionID int     `json:"prescriptionId"`
	MedicineID     int     `json:"medicineId"`
	Qty            float64 `json:"qty"`
	UnitID         int     `json:"unitId"`
	Price          float64 `json:"price"`
	Discount       float64 `json:"discount"`
	Subtotal       float64 `json:"subtotal"`
}

type Prescription struct {
	ID                   int           `json:"id"`
	InvoiceID            int           `json:"invoiceId"`
	Number               int           `json:"number"`
	PrescriptionDate     time.Time     `json:"prescriptionDate"`
	PatientID            int           `json:"patientId"`
	DoctorID             int           `json:"doctorId"`
	Qty                  float64       `json:"qty"`
	Price                float64       `json:"price"`
	TotalPrice           float64       `json:"totalPrice"`
	Description          string        `json:"description"`
	CreatedAt            time.Time     `json:"createdAt"`
	UserID               int           `json:"userId"`
	LastModified         time.Time     `json:"lastModified"`
	LastModifiedByUserID int           `json:"lastLastModifiedByUserId"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}
