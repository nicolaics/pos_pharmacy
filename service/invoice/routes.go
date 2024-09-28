package invoice

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
		unitStore:          unitStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleGetInvoiceNumberForToday).Methods(http.MethodGet)
	router.HandleFunc("/invoice/{params}/{val}", h.handleGetInvoices).Methods(http.MethodPost)
	router.HandleFunc("/invoice/detail", h.handleGetInvoiceDetail).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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
	if paymentMethod == nil {
		err = h.paymentMethodStore.CreatePaymentMethod(payload.PaymentMethodName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create payment method %s", payload.PaymentMethodName))
			return
		}

		paymentMethod, err = h.paymentMethodStore.GetPaymentMethodByName(payload.PaymentMethodName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found", payload.PaymentMethodName))
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse date"))
		return
	}

	// check duplicate
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Number, payload.CustomerID, *invoiceDate)
	if err == nil || invoiceId != 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d with today's date exists", payload.Number))
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
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get invoice id
	invoiceId, err = h.invoiceStore.GetInvoiceID(payload.Number, payload.CustomerID, *invoiceDate)
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

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	log.Println("start date:", *startDate)
	log.Println("end date:", *endDate)

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	var invoices []types.InvoiceListsReturnPayload

	if val == "all" {
		invoices, err = h.invoiceStore.GetInvoicesByDate(*startDate, *endDate)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		invoice, err := h.invoiceStore.GetInvoiceByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d not exist", id))
			return
		}

		user, err := h.userStore.GetUserByID(invoice.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user %d not exist", invoice.UserID))
			return
		}

		customer, err := h.custStore.GetCustomerByID(invoice.CustomerID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("customer %d not exist", invoice.CustomerID))
			return
		}

		paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByID(invoice.PaymentMethodID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("payment method %d not exist", invoice.PaymentMethodID))
			return
		}

		invoices = append(invoices, types.InvoiceListsReturnPayload{
			ID:                invoice.ID,
			Number:            invoice.Number,
			UserName:          user.Name,
			CustomerName:      customer.Name,
			Subtotal:          invoice.Subtotal,
			Discount:          invoice.Discount,
			Tax:               invoice.Tax,
			TotalPrice:        invoice.TotalPrice,
			PaymentMethodName: paymentMethod.Name,
			Description:       invoice.Description,
			InvoiceDate:       invoice.InvoiceDate,
		})
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "user" {
		user, err := h.userStore.GetUserByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val))
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndUserID(*startDate, *endDate, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s doesn't create any invoice between %s and %s", val, payload.StartDate, payload.EndDate))
			return
		}
	} else if params == "customer" {
		customer, err := h.custStore.GetCustomerByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not exists", val))
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndCustomerID(*startDate, *endDate, customer.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s doesn't have any invoice between %s and %s", val, payload.StartDate, payload.EndDate))
			return
		}
	} else if params == "payment-method" {
		paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found between %s and %s", val, payload.StartDate, payload.EndDate))
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndPaymentMethodID(*startDate, *endDate, paymentMethod.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s doesn't have any invoice between %s and %s", val, payload.StartDate, payload.EndDate))
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"))
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

	err = h.invoiceStore.DeleteMedicineItems(invoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteInvoice(invoice, user)
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

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
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
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.invoiceStore.DeleteMedicineItems(invoice, user)
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
