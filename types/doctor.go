package types

import (
	"time"
)

type DoctorStore interface {
	GetDoctorByName(name string) (*Doctor, error)
	GetDoctorsBySearchName(name string) ([]Doctor, error)
	GetDoctorByID(id int) (*Doctor, error)
	CreateDoctor(Doctor) error
	GetAllDoctors() ([]Doctor, error)
	DeleteDoctor(*Doctor, *User) error
	ModifyDoctor(int, string, *User) error
}

type RegisterDoctorPayload struct {
	Name string `json:"name" validate:"required"`
}
type ModifyDoctorPayload struct {
	ID      int                   `json:"id" validate:"required"`
	NewData RegisterDoctorPayload `json:"newData" validate:"required"`
}

type DeleteDoctorPayload struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GetOneDoctorPayload struct {
	ID int `json:"id" validate:"required"`
}

type Doctor struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
