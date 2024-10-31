package types

import (
	"time"
)

type PatientStore interface {
	GetPatientByName(name string, age int) (*Patient, error)
	GetPatientsBySearchName(name string) ([]Patient, error)
	GetPatientByID(id int) (*Patient, error)

	CreatePatient(Patient) error

	GetAllPatients() ([]Patient, error)

	DeletePatient(*Patient, *User) error

	ModifyPatient(int, RegisterPatientPayload, *User) error
}

type RegisterPatientPayload struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age"`
}
type ModifyPatientPayload struct {
	ID      int                    `json:"id" validate:"required"`
	NewData RegisterPatientPayload `json:"newData" validate:"required"`
}

type DeletePatientPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetOnePatientPayload struct {
	ID int `json:"id" validate:"required"`
}

type Patient struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"createdAt"`
}
