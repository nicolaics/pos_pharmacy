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
	userStore          types.UserStore
	custStore          types.CustomerStore
	paymentMethodStore types.PaymentMethodStore
	medStore           types.MedicineStore
	unitStore          types.UnitStore
}

func NewHandler(invoiceStore types.InvoiceStore, userStore types.UserStore,
	custStore types.CustomerStore, paymentMethodStore types.PaymentMethodStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		invoiceStore:       invoiceStore,
		userStore:          userStore,
		custStore:          custStore,
		paymentMethodStore: paymentMethodStore,
		medStore:           medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleGetInvoiceNumberForToday).Methods(http.MethodGet)
	router.HandleFunc("/invoice/all", h.handleGetInvoices).Methods(http.MethodPost)
	router.HandleFunc("/invoice/detail", h.handleGetInvoiceDetail).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/all", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterInvoicePayload

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
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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
		Number:               payload.Number,
		UserID:               user.ID,
		CustomerID:           payload.CustomerID,
		Subtotal:             payload.Subtotal,
		Discount:             payload.Discount,
		Tax:                  payload.Tax,
		TotalPrice:           payload.TotalPrice,
		PaidAmount:           payload.PaidAmount,
		ChangeAmount:         payload.ChangeAmount,
		PaymentMethodID:      paymentMethod.ID,
		Description:          payload.Description,
		InvoiceDate:          payload.InvoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get invoice id
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Number, user.ID, payload.CustomerID, payload.TotalPrice, payload.InvoiceDate)
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
			InvoiceID:  invoiceId,
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

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("invoice %d successfully created by %s", payload.Number, user.Name))
}

// beginning of invoice page, will request here
func (h *Handler) handleGetInvoiceNumberForToday(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	numberOfInvoices, err := h.invoiceStore.GetNumberOfInvoices()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"nextNumber": (numberOfInvoices + 1)})
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
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment name id %d doesn't exists", invoice.PaymentMethodID))
		return
	}

	// get medicine items of the invoice
	medicineItems, err := h.invoiceStore.GetMedicineItems(invoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get user data, the one who inputs the invoice
	inputter, err := h.userStore.GetUserByID(invoice.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", invoice.UserID))
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(invoice.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", invoice.LastModifiedByUserID))
		return
	}

	returnPayload := types.InvoiceDetailPayload{
		ID:                     invoice.ID,
		Number:                 invoice.Number,
		Subtotal:               invoice.Subtotal,
		Discount:               invoice.Discount,
		Tax:                    invoice.Tax,
		TotalPrice:             invoice.TotalPrice,
		PaidAmount:             invoice.PaidAmount,
		ChangeAmount:           invoice.ChangeAmount,
		Description:            invoice.Description,
		InvoiceDate:            invoice.InvoiceDate,
		LastModified:           invoice.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,

		User: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   inputter.ID,
			Name: inputter.Name,
		},

		Customer: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   customer.ID,
			Name: customer.Name,
		},

		PaymentMethod: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   paymentMethod.ID,
			Name: paymentMethod.Name,
		},

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
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err))
		return
	}

	// check if the purchase invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if invoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice id %d doesn't exist", payload.InvoiceID))
		return
	}

	err = h.invoiceStore.DeleteMedicineItems(invoice, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteInvoice(invoice, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("invoice number %d deleted by %s", invoice.Number, user.Name))
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
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	// get payment name
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.NewData.PaymentMethodName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment name %s not found", payload.NewData.PaymentMethodName))
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice with id %d doesn't exists", payload.ID))
		return
	}

	err = h.invoiceStore.ModifyInvoice(payload.ID, types.Invoice{
		Number:               payload.NewData.Number,
		CustomerID:           payload.NewData.CustomerID,
		Subtotal:             payload.NewData.Subtotal,
		Discount:             payload.NewData.Discount,
		Tax:                  payload.NewData.Tax,
		TotalPrice:           payload.NewData.TotalPrice,
		PaidAmount:           payload.NewData.PaidAmount,
		ChangeAmount:         payload.NewData.ChangeAmount,
		PaymentMethodID:      paymentMethod.ID,
		Description:          payload.NewData.Description,
		InvoiceDate:          payload.NewData.InvoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteMedicineItems(invoice, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	for _, medicine := range payload.NewData.MedicineLists {
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
				fmt.Errorf("invoice %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("invoice modified by %s", user.Name))
}
