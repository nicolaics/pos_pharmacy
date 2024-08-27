package customer

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	store types.CustomerStore
}

func NewHandler(store types.CustomerStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register-customer", h.handleRegister).Methods("POST")
	// router.HandleFunc("/get-customers", h.handleRegister).Methods("GET")
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterCustomer

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// check if the cashier exists
	_, err := h.store.GetCustomerByName(payload.Name)

	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s already exists", payload.Name))
		return
	}

	err = h.store.CreateCustomer(types.Customer{
		Name: payload.Name,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
