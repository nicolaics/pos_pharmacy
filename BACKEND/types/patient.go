package types

import (
	"time"
)

type RegisterPatientPayload struct {
	Name string `json:"name" validate:"required"`
}
type ModifyPatientPayload struct {
	ID      int    `json:"id" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type DeletePatientPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type PatientStore interface {
	GetPatientByName(name string) (*Patient, error)
	GetPatientByID(id int) (*Patient, error)
	CreatePatient(Patient) error
	GetAllPatients() ([]Patient, error)
	DeletePatient(*Patient) error
	ModifyPatient(int, string) error
}

type Patient struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"createdAt"`
}
