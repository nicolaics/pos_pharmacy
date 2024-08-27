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
	custStore    types.CustomerStore
	cashierStore types.CashierStore
}

func NewHandler(custStore types.CustomerStore, cashierStore types.CashierStore) *Handler {
	return &Handler{custStore: custStore, cashierStore: cashierStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/customer", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/customer", h.handleGetAllCustomer).Methods(http.MethodGet)
	router.HandleFunc("/customer", h.handleDeleteCustomer).Methods(http.MethodDelete)
	router.HandleFunc("/customer", h.handleModifyCustomer).Methods(http.MethodPatch)

	router.HandleFunc("/customer", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.CustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the customer exists
	_, err = h.custStore.GetCustomerByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s already exists", payload.Name))
		return
	}

	err = h.custStore.CreateCustomer(types.Customer{
		Name: payload.Name,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("customer %s successfully created by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleGetAllCustomer(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	customers, err := h.custStore.GetAllCustomers()
	if err == nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, customers)
}

func (h *Handler) handleDeleteCustomer(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.CustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the customer exists
	customer, err := h.custStore.GetCustomerByName(payload.Name)
	if customer == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s doesn't exist", payload.Name))
		return
	}

	err = h.custStore.DeleteCustomer(customer)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("customer %s deleted by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleModifyCustomer(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// validate token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the customer exists
	_, err = h.custStore.GetCustomerByName(payload.NewName)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s already exist", payload.NewName))
		return
	}

	customer, err := h.custStore.GetCustomerByName(payload.OldName)
	if customer == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s doesn't exist", payload.OldName))
		return
	}

	err = h.custStore.ModifyCustomer(customer, payload.NewName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("customer %s modified into %s by %s",
														payload.OldName, payload.NewName, cashier.Name))
}
