package invoice

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	store types.InvoiceStore
}

func NewHandler(store types.InvoiceStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/new-invoice", h.handleNewInvoice).Methods("POST")
	// router.HandleFunc("/get-customers", h.handleRegister).Methods("GET")
}

func (h *Handler) handleNewInvoice(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.NewInvoice

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// check if the invoice exists
	invoices, err := h.store.GetInvoicesByDate(payload.InvoiceDate)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	for _, invoice := range(invoices) {
		if payload.Number == invoice.Number {
			utils.WriteError(w, http.StatusBadRequest,
						fmt.Errorf("invoice with number %d and date %s already exists",
									payload.Number, payload.InvoiceDate))
			return
		}
	}

	err = h.store.CreateInvoice(types.Invoice{
		Number: payload.Number,
		CashierName: payload.CashierName,
		CustomerName: payload.CustomerName,
		Subtotal: payload.Subtotal,
		Discount: payload.Discount,
		TotalPrice: payload.TotalPrice,
		PaymentMethodName: payload.PaymentMethodName,
		PaidAmount: payload.PaidAmount,
		ChangeAmount: payload.ChangeAmount,
		Description: payload.Description,
		InvoiceDate: payload.InvoiceDate,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
