package types

import (
	"time"
)

type RegisterDoctorPayload struct {
	Name string `json:"name" validate:"required"`
}
type ModifyDoctorPayload struct {
	ID      int    `json:"id" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type DeleteDoctorPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type DoctorStore interface {
	GetDoctorByName(name string) (*Doctor, error)
	GetDoctorByID(id int) (*Doctor, error)
	CreateDoctor(Doctor) error
	GetAllDoctors() ([]Doctor, error)
	DeleteDoctor(*Doctor) error
	ModifyDoctor(int, string) error
}

type Doctor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"createdAt"`
}
