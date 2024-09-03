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
	invoiceStore       types.InvoiceStore
	cashierStore       types.CashierStore
	custStore          types.CustomerStore
	paymentMethodStore types.PaymentMethodStore
	medStore           types.MedicineStore
	unitStore          types.UnitStore
}

func NewHandler(invoiceStore types.InvoiceStore, cashierStore types.CashierStore,
	custStore types.CustomerStore, paymentMethodStore types.PaymentMethodStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		invoiceStore:       invoiceStore,
		cashierStore:       cashierStore,
		custStore:          custStore,
		paymentMethodStore: paymentMethodStore,
		medStore:           medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice", h.handleNew).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleGetInvoices).Methods(http.MethodGet)
	router.HandleFunc("/invoice/detail", h.handleGetInvoiceDetail).Methods(http.MethodGet)
	router.HandleFunc("/invoice", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleNew(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.NewInvoicePayload

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
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	// check customerID
	_, err = h.custStore.GetCustomerByID(payload.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d not found", payload.CustomerID))
		return
	}

	// check paymentMethodName
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.PaymentMethodName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found", payload.PaymentMethodName))
		return
	}

	err = h.invoiceStore.CreateInvoice(types.Invoice{
		Number:          payload.Number,
		CashierID:       cashier.ID,
		CustomerID:      payload.CustomerID,
		Subtotal:        payload.Subtotal,
		Discount:        payload.Discount,
		Tax:             payload.Tax,
		TotalPrice:      payload.TotalPrice,
		PaidAmount:      payload.PaidAmount,
		ChangeAmount:    payload.ChangeAmount,
		PaymentMethodID: paymentMethod.ID,
		Description:     payload.Description,
		InvoiceDate:     payload.InvoiceDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get invoice id
	invoice, err := h.invoiceStore.GetInvoiceByAll(payload.Number, cashier.ID, payload.CustomerID, payload.TotalPrice, payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d doesn't exists", payload.Number))
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = h.invoiceStore.CreateMedicineItems(types.MedicineItems{
			InvoiceID:  invoice.ID,
			MedicineID: medData.ID,
			Qty:        medicine.Qty,
			UnitID:     unit.ID,
			Price:      medicine.Price,
			Discount:   medicine.Discount,
			Subtotal:   medicine.Subtotal,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("invoice %d successfully created by %s", payload.Number, cashier.Name))
}

// only view the purchase invoice list
func (h *Handler) handleGetInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoicePayload

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
	_, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	invoices, err := h.invoiceStore.GetInvoicesByDate(payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, invoices)
}

func (h *Handler) handleGetInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoiceDetailPayload

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
	_, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	// get invoice data
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d doesn't exists", payload.InvoiceID))
		return
	}

	// get customer data
	customer, err := h.custStore.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d doesn't exists", invoice.CustomerID))
		return
	}

	// get payment method data
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByID(invoice.PaymentMethodID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method id %d doesn't exists", invoice.PaymentMethodID))
		return
	}

	// get medicine items of the invoice
	medicineItems, err := h.invoiceStore.GetMedicineItems(invoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get cashier data, the one who inputs the invoice
	inputter, err := h.cashierStore.GetCashierByID(invoice.CashierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier id %d doesn't exists", invoice.CashierID))
		return
	}

	returnPayload := types.InvoiceDetailPayload{
		InvoiceID:          invoice.ID,
		InvoiceNumber:      invoice.Number,
		InvoiceSubtotal:    invoice.Subtotal,
		InvoiceDiscount:    invoice.Discount,
		InvoiceTax:         invoice.Tax,
		InvoiceTotalPrice:  invoice.TotalPrice,
		PaidAmount:         invoice.PaidAmount,
		ChangeAmount:       invoice.ChangeAmount,
		InvoiceDescription: invoice.Description,
		InvoiceDate:        invoice.InvoiceDate,

		CashierID:   inputter.ID,
		CashierName: inputter.Name,

		CustomerID:   customer.ID,
		CustomerName: customer.Name,

		PaymentMethodID:   paymentMethod.ID,
		PaymentMethodName: paymentMethod.Method,

		MedicineLists: medicineItems,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteInvoicePayload

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
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid or not admin: %v", err))
		return
	}

	// check if the purchase invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if invoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice id %d doesn't exist", payload.InvoiceID))
		return
	}

	err = h.invoiceStore.DeleteMedicineItems(invoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteInvoice(invoice)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("invoice number %d deleted by %s", invoice.Number, cashier.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyInvoicePayload

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
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	// get payment method
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.NewPaymentMethodName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found", payload.NewPaymentMethodName))
		return
	}

	// check if the invoice exists
	_, err = h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice with id %d doesn't exists", payload.ID))
		return
	}

	err = h.invoiceStore.ModifyInvoice(payload.ID, types.Invoice{
		Number:          payload.NewNumber,
		CashierID:       cashier.ID,
		CustomerID:      payload.NewCustomerID,
		Subtotal:        payload.NewSubtotal,
		Discount:        payload.NewDiscount,
		Tax:             payload.NewTax,
		TotalPrice:      payload.NewTotalPrice,
		PaidAmount:      payload.NewPaidAmount,
		ChangeAmount:    payload.NewChangeAmount,
		PaymentMethodID: paymentMethod.ID,
		Description:     payload.NewDescription,
		InvoiceDate:     payload.NewInvoiceDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteMedicineItems(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	for _, medicine := range payload.NewMedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = h.invoiceStore.CreateMedicineItems(types.MedicineItems{
			InvoiceID:  payload.ID,
			MedicineID: medData.ID,
			Qty:        medicine.Qty,
			UnitID:     unit.ID,
			Price:      medicine.Price,
			Discount:   medicine.Discount,
			Subtotal:   medicine.Subtotal,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("invoice %d, med %s: %v", payload.NewNumber, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("invoice modified by %s", cashier.Name))
}
