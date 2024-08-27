package supplier

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	supplierStore types.SupplierStore
	cashierStore types.CashierStore
}

func NewHandler(supplierStore types.SupplierStore, cashierStore types.CashierStore) *Handler {
	return &Handler{supplierStore: supplierStore, cashierStore: cashierStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/supplier", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/supplier", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/supplier", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/supplier", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/supplier", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterSupplierPayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	// check if the supplier exists
	_, err = h.supplierStore.GetSupplierByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("supplier with name %s already exists", payload.Name))
		return
	}

	err = h.supplierStore.CreateSupplier(types.Supplier{
		Name: payload.Name,
		Address: payload.Address,
		CompanyPhoneNumber: payload.CompanyPhoneNumber,
		ContactPersonName: payload.ContactPersonName,
		ContactPersonNumber: payload.ContactPersonNumber,
		Terms: payload.Terms,
		VendorIsTaxable: payload.VendorIsTaxable,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("supplier %s created by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate cashier token
	_, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	// check if the supplier exists
	suppliers, err := h.supplierStore.GetAllSuppliers()
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, suppliers)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteSupplierPayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	// check if the supplier exists
	supplier, err := h.supplierStore.GetSupplierByName(payload.Name)
	if err != nil || supplier == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("supplier with name %s doesn't exists", payload.Name))
		return
	}

	err = h.supplierStore.DeleteSupplier(supplier)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("supplier %s deleted by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifySupplierPayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	// check if the supplier exists
	supplier, err := h.supplierStore.GetSupplierByName(payload.Name)
	if err != nil || supplier == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("supplier with name %s doesn't exists", payload.Name))
		return
	}

	err = h.supplierStore.ModifySupplier(supplier.ID, types.Supplier{
														Name: payload.Name,
														Address: payload.Address,
														CompanyPhoneNumber: payload.CompanyPhoneNumber,
														ContactPersonName: payload.ContactPersonName,
														ContactPersonNumber: payload.ContactPersonNumber,
														Terms: payload.Terms,
														VendorIsTaxable: payload.VendorIsTaxable,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("supplier %s modified by %s", payload.Name, cashier.Name))
}
