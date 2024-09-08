package types

import (
	"net/http"
	"time"
)

type UserStore interface {
	GetUserByName(string) (*User, error)
	GetUserByID(int) (*User, error)
	GetAllUsers() ([]User, error)

	CreateUser(User) error
	
	DeleteUser(*User, int) error

	UpdateLastLoggedIn(int) error
	ModifyUser(int, User, int) error
	
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
	Name          string `json:"name" validate:"required"`
}

// modify the data of the user, requires admin password and admin to do it
type ModifyUserPayload struct {
	ID      int                 `json:"id" validate:"required"`
	NewData RegisterUserPayload `json:"newData" validate:"required"`
}

// normal log-in
type LoginUserPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
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
