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
	custStore types.CustomerStore
	cashierStore types.CashierStore
}

func NewHandler(custStore types.CustomerStore, cashierStore types.CashierStore) *Handler {
	return &Handler{custStore: custStore, cashierStore: cashierStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/customer/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/customer/register", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/customer", h.handleGetAllCustomer).Methods(http.MethodGet)
	router.HandleFunc("/customer", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterCustomerPayload

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
	_, err := h.cashierStore.ValidateCashierToken(w, r, false)
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

	utils.WriteJSON(w, http.StatusCreated, nil)
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
