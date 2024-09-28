package types

import (
	"net/http"
	"time"
)

type UserStore interface {
	GetUserByName(string) (*User, error)
	GetUserByID(int) (*User, error)
	GetAllUsers() ([]User, error)

	GetUserBySearchName(string) ([]User, error)
	GetUserBySearchPhoneNumber(string) ([]User, error)

	CreateUser(User) error

	DeleteUser(*User, *User) error

	UpdateLastLoggedIn(int) error
	ModifyUser(int, User, *User) error

	SaveToken(int, *TokenDetails) error
	DeleteToken(int) error
	ValidateUserToken(http.ResponseWriter, *http.Request, bool) (*User, error)
}

// register new user
type RegisterUserPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required,min=3,max=130"`
	PhoneNumber   string `json:"phoneNumber" validate:"required"`
	Admin         bool   `json:"admin"`
}

// delete user account
type RemoveUserPayload struct {
	AdminPassword string `json:"adminPassword" validate:"required"`
	ID            int    `json:"id" validate:"required"`
}

// modify the data of the user, requires admin password and admin to do it
type ModifyUserPayload struct {
	ID      int                 `json:"id" validate:"required"`
	NewData RegisterUserPayload `json:"newData" validate:"required"`
}

// to change a user admin status
type ChangeAdminStatusPayload struct {
	ID            int    `json:"id" validate:"required"`
	AdminPassword string `json:"adminPassword" validate:"required"`
	Admin         bool   `json:"admin" validate:"required"`
}

// get one user data
type GetOneUserPayload struct {
	ID int `json:"id" validate:"required"`
}

// normal log-in
type LoginUserPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// validate token request from client
type VerifyTokenRequestFromClientPayload struct {
	NeedAdmin bool `json:"needAdmin" validate:"required"`
}

// basic user data info
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Password     string    `json:"password"`
	Admin        bool      `json:"admin"`
	PhoneNumber  string    `json:"phoneNumber"`
	LastLoggedIn time.Time `json:"lastLoggedIn"`
	CreatedAt    time.Time `json:"createdAt"`
}
