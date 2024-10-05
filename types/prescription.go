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

	GetPrescriptionID(invoiceId int, number int, date time.Time, patientId int, totalPrice float64, doctorId int) (int, error)

	CreatePrescription(Prescription) error
	DeletePrescription(*Prescription, *User) error
	ModifyPrescription(int, Prescription, *User) error

	// delete entirely from the db if there's error
	AbsoluteDeletePrescription(presc Prescription) error

	CreatePrescriptionMedicineItem(PrescriptionMedicineItem) error
	GetPrescriptionMedicineItems(setItemId int) ([]PrescriptionMedicineItemReturn, error)
	GetPrescriptionMedicineItemID(PrescriptionMedicineItem) (int, error)
	DeletePrescriptionMedicineItem(pres *Prescription, setItemId int, user *User) error

	GetSetItemByID(int) (*PrescriptionSetItem, error)
	GetSetItemsByPrescriptionID(int) ([]PrescriptionSetItem, error)
	GetSetItemID(PrescriptionSetItem) (int, error)
	GetPrescriptionSetAndMedicineItems(prescriptionId int) ([]PrescriptionSetItemReturn, error)
	CreateSetItem(PrescriptionSetItem) error
	DeleteSetItem(*Prescription, *User) error

	CreateEticket(Eticket) error
	DeleteEticket(int) error
	GetEticketID(Eticket) (int, error)

	// tabla nemae = prescription, eticket
	UpdatePDFUrl(tableName string, id int, fileName string) error
	IsPDFUrlExist(tableName string, fileName string) (bool, error)

	UpdateEticketID(eticketId int, prescSetItemId int) error
}

type RegisterPrescriptionPayload struct {
	Invoice struct {
		Number       int    `json:"number" validate:"required"`
		CustomerName string `json:"customerName" validate:"required"`
		InvoiceDate  string `json:"invoiceDate" validate:"required"`
	} `json:"invoice" validate:"required"`

	Number           int                          `json:"number" validate:"required"`
	PrescriptionDate string                       `json:"prescriptionDate" validate:"required"`
	PatientName      string                       `json:"patientName" validate:"required"`
	PatientAge       int                          `json:"patientAge"`
	DoctorName       string                       `json:"doctorName" validate:"required"`
	Qty              float64                      `json:"qty" validate:"required"`
	Price            float64                      `json:"price" validate:"required"`
	TotalPrice       float64                      `json:"totalPrice" validate:"required"`
	Description      string                       `json:"description"`
	SetItems         []PrescriptionSetItemPayload `json:"setItems" validate:"required"`
}

type PrescriptionSetItemPayload struct {
	MedicineLists []PrescriptionMedicineListPayload `json:"medicineLists" validate:"required"`
	Mf            string                            `json:"mf"`
	Dose          string                            `json:"dose"`
	SetUnit       string                            `json:"setUnit"`
	ConsumeTime   string                            `json:"consumeTime"`
	Det           string                            `json:"det"`
	Usage         string                            `json:"usage"`
	MustFinish    bool                              `json:"mustFinish"`
	PrintEticket  bool                              `json:"printEticket"`
	Eticket       struct {
		Number      int     `json:"number"`
		MedicineQty float64 `json:"medicineQty"`
	} `json:"eticket"`
}

type PrescriptionMedicineListPayload struct {
	MedicineBarcode string  `json:"medicineBarcode" validate:"required"`
	MedicineName    string  `json:"medicineName" validate:"required"`
	Qty             string  `json:"qty" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	Price           float64 `json:"price" validate:"required"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal" validate:"required"`
}

// prescription list payload returned to user after searching
type PrescriptionListsReturnPayload struct {
	ID               int       `json:"id"`
	Number           int       `json:"number"`
	PrescriptionDate time.Time `json:"prescriptionDate"`
	PatientName      string    `json:"patientName"`
	PatientAge       int       `json:"patientAge"`
	DoctorName       string    `json:"doctorName"`
	Qty              float64   `json:"qty"`
	Price            float64   `json:"price"`
	TotalPrice       float64   `json:"totalPrice"`
	Description      string    `json:"description"`
	UserName         string    `json:"userName"`

	Invoice struct {
		Number       int       `json:"number"`
		CustomerName string    `json:"customerName"`
		TotalPrice   float64   `json:"totalPrice"`
		InvoiceDate  time.Time `json:"invoiceDate"`
	} `json:"invoice"`
}

// only view the purchase invoice list
type ViewPrescriptionsPayload struct {
	StartDate string `json:"startDate" validate:"required"` // if empty, just give today's date from morning
	EndDate   string `json:"endDate" validate:"required"`   // if empty, just give today's date to current time
}

// view the detail of the prescription
type ViewPrescriptionDetailPayload struct {
	PrescriptionID int `json:"prescriptionId" validate:"required"`
}

type ModifyPrescriptionPayload struct {
	ID      int                         `json:"id" validate:"required"`
	NewData RegisterPrescriptionPayload `json:"newData" validate:"required"`
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
	PDFUrl                 string    `json:"prescPdfUrl"`

	Invoice struct {
		Number       int       `json:"number"`
		CustomerName string    `json:"customerName"`
		TotalPrice   float64   `json:"totalPrice"`
		InvoiceDate  time.Time `json:"invoiceDate"`
	} `json:"invoice"`

	Patient struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	} `json:"patient"`

	Doctor struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"doctor"`

	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`

	MedicineSets []PrescriptionSetItemReturn `json:"medicineSets"`
}

type PrescriptionSetItemReturn struct {
	ID            int                              `json:"id"`
	Mf            string                           `json:"mf"`
	Dose          string                           `json:"dose"`
	SetUnit       string                           `json:"setUnit"`
	ConsumeTime   string                           `json:"consumeTime"`
	Det           string                           `json:"det"`
	Usage         string                           `json:"usage"`
	MustFinish    bool                             `json:"mustFinish"`
	PrintEticket  bool                             `json:"printEticket"`
	EticketID     int                              `json:"eticketId"`
	MedicineItems []PrescriptionMedicineItemReturn `json:"medicineItems"`
}

// data of the medicine per row in the prescription
type PrescriptionMedicineItemReturn struct {
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	QtyString       string  `json:"qtyString"`
	QtyFloat        float64 `json:"qtyFloat"`
	Unit            string  `json:"unit"`
	Price           float64 `json:"price"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal"`
}

type PrescriptionMedicineItemTemp struct {
	MedicineBarcode string  `json:"medicineBarcode"`
	MedicineName    string  `json:"medicineName"`
	Qty             float64 `json:"qty"`
	Unit            string  `json:"unit"`
	Price           float64 `json:"price"`
	Discount        float64 `json:"discount"`
	Subtotal        float64 `json:"subtotal"`
}

type DeletePrescription struct {
	ID int `json:"id" validate:"required"`
}

type PrescriptionMedicineItem struct {
	ID                    int     `json:"id"`
	PrescriptionSetItemID int     `json:"prescriptionSetItemId"`
	MedicineID            int     `json:"medicineId"`
	Qty                   float64 `json:"qty"`
	UnitID                int     `json:"unitId"`
	Price                 float64 `json:"price"`
	Discount              float64 `json:"discount"`
	Subtotal              float64 `json:"subtotal"`
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
	PDFUrl               string        `json:"pdfUrl"`
	DeletedAt            sql.NullTime  `json:"deletedAt"`
	DeletedByUserID      sql.NullInt64 `json:"deletedByUserId"`
}

type PrescriptionSetItem struct {
	ID             int  `json:"id"`
	PrescriptionID int  `json:"prescriptionId"`
	MfID           int  `json:"mfId"`
	DoseID         int  `json:"doseId"`
	SetUnitID      int  `json:"setUnitId"`
	ConsumeTimeID  int  `json:"consumeTimeId"`
	DetID          int  `json:"detId"`
	UsageID        int  `json:"usageId"`
	MustFinish     bool `json:"mustFinish"`
	PrintEticket   bool `json:"printEticket"`
	EticketID      int  `json:"eticketId"`
}

type Eticket struct {
	ID                    int       `json:"id"`
	PrescriptionID        int       `json:"prescriptionId"`
	PrescriptionSetItemID int       `json:"prescriptionSetItemId"`
	Number                int       `json:"number"`
	MedicineQty           float64   `json:"medicineQty"`
	PDFUrl                string    `json:"pdfUrl"`
	CreatedAt             time.Time `json:"createdAt"`
}

type EticketReturnPayload struct {
	Number int `json:"number"`
}

type PrescriptionPDFReturn struct {
	Number       int
	Date         time.Time
	Patient      Patient
	Doctor       Doctor
	MedicineSets []PrescriptionSetItemReturn
}
