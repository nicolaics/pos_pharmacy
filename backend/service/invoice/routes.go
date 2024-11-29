package invoice

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"github.com/nicolaics/pharmacon/utils/pdf"
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
	router.HandleFunc("/invoice/print", h.handlePrint).Methods(http.MethodPost)
	router.HandleFunc("/invoice/print-receipt", h.handlePrintReceipt).Methods(http.MethodPost)

	router.HandleFunc("/invoice", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/print", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/print-receipt", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check customerID
	_, err = h.custStore.GetCustomerByID(payload.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d not found", payload.CustomerID), nil)
		return
	}

	// check paymentMethodName
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.PaymentMethodName)
	if paymentMethod == nil {
		err = h.paymentMethodStore.CreatePaymentMethod(payload.PaymentMethodName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create payment method %s", payload.PaymentMethodName), nil)
			return
		}

		paymentMethod, err = h.paymentMethodStore.GetPaymentMethodByName(payload.PaymentMethodName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found", payload.PaymentMethodName), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse date"), nil)
		return
	}

	// no need to check for duplicates, because number will be given

	newInvoice := types.Invoice{
		Number:               payload.Number,
		UserID:               user.ID,
		CustomerID:           payload.CustomerID,
		Subtotal:             payload.Subtotal,
		DiscountPercentage:   payload.DiscountPercentage,
		DiscountAmount:       payload.DiscountAmount,
		TaxPercentage:        payload.TaxPercentage,
		TaxAmount:            payload.TaxAmount,
		TotalPrice:           payload.TotalPrice,
		PaidAmount:           payload.PaidAmount,
		ChangeAmount:         payload.ChangeAmount,
		PaymentMethodID:      paymentMethod.ID,
		Description:          payload.Description,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	}
	err = h.invoiceStore.CreateInvoice(newInvoice)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
	}

	// get invoice id
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Number, payload.CustomerID, *invoiceDate)
	if err != nil {
		errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
			return
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d doesn't exists", payload.Number), nil)
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.invoiceStore.CreateMedicineItem(types.InvoiceMedicineItem{
			InvoiceID:          invoiceId,
			MedicineID:         medData.ID,
			Qty:                medicine.Qty,
			UnitID:             unit.ID,
			Price:              medicine.Price,
			DiscountPercentage: medicine.DiscountPercentage,
			DiscountAmount:     medicine.DiscountAmount,
			Subtotal:           medicine.Subtotal,
		})
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err), nil)
			return
		}

		err = utils.CheckStock(medData, unit, medicine.Qty)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough", medicine.MedicineName), nil)
			return
		}
	}

	invoicePdf := types.InvoicePdfPayload{
		Number:             payload.Number,
		UserName:           user.Name,
		Subtotal:           payload.Subtotal,
		DiscountPercentage: payload.DiscountPercentage,
		DiscountAmount:     payload.DiscountAmount,
		TaxPercentage:      payload.TaxPercentage,
		TaxAmount:          payload.TaxAmount,
		TotalPrice:         payload.TotalPrice,
		PaidAmount:         payload.PaidAmount,
		ChangeAmount:       payload.ChangeAmount,
		Description:        payload.Description,
		InvoiceDate:        *invoiceDate,
		MedicineLists:      payload.MedicineLists,
	}
	invoiceFileName, err := pdf.CreateInvoicePdf(invoicePdf, h.invoiceStore, "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create invoice pdf: %v", err), nil)
		return
	}

	err = h.invoiceStore.UpdatePdfUrl(invoiceId, invoiceFileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update invoice pdf url: %v", err), nil)
		return
	}

	// reduce the stock
	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}

		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
			return
		}

		err = h.medStore.InsertIntoMedicineHistoryTable(medData.ID, invoiceId, constants.MEDICINE_HISTORY_OUT, medicine.Qty, unit.ID, *invoiceDate)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete invoice: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error insert medicine history: %v", err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("invoice %d successfully created by %s", payload.Number, user.Name), nil)
}

// beginning of invoice page, will request here
func (h *Handler) handleGetInvoiceNumberForToday(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	startDate, err := utils.ParseStartDate(time.Now().Format("2006-01-02 -0700MST"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parse start date: %v", err), nil)
		return
	}

	endDate, err := utils.ParseEndDate(time.Now().Format("2006-01-02 -0700MST"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parse end date: %v", err), nil)
		return
	}

	numberOfInvoices, err := h.invoiceStore.GetNumberOfInvoices(*startDate, *endDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, (numberOfInvoices + 1), nil)
}

// only view the purchase invoice list
func (h *Handler) handleGetInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		invoice, err := h.invoiceStore.GetInvoiceByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d not exist", id), nil)
			return
		}

		user, err := h.userStore.GetUserByID(invoice.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user %d not exist", invoice.UserID), nil)
			return
		}

		customer, err := h.custStore.GetCustomerByID(invoice.CustomerID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("customer %d not exist", invoice.CustomerID), nil)
			return
		}

		paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByID(invoice.PaymentMethodID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("payment method %d not exist", invoice.PaymentMethodID), nil)
			return
		}

		invoices = append(invoices, types.InvoiceListsReturnPayload{
			ID:                 invoice.ID,
			Number:             invoice.Number,
			UserName:           user.Name,
			CustomerName:       customer.Name,
			Subtotal:           invoice.Subtotal,
			DiscountPercentage: invoice.DiscountPercentage,
			DiscountAmount:     invoice.DiscountAmount,
			TaxPercentage:      invoice.TaxPercentage,
			TaxAmount:          invoice.TaxAmount,
			TotalPrice:         invoice.TotalPrice,
			PaymentMethodName:  paymentMethod.Name,
			Description:        invoice.Description,
			InvoiceDate:        invoice.InvoiceDate,
		})
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else if params == "user" {
		user, err := h.userStore.GetUserByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val), nil)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndUserID(*startDate, *endDate, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s doesn't create any invoice between %s and %s", val, payload.StartDate, payload.EndDate), nil)
			return
		}
	} else if params == "customer" {
		customer, err := h.custStore.GetCustomerByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not exists", val), nil)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndCustomerID(*startDate, *endDate, customer.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s doesn't have any invoice between %s and %s", val, payload.StartDate, payload.EndDate), nil)
			return
		}
	} else if params == "payment-method" {
		paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s not found between %s and %s", val, payload.StartDate, payload.EndDate), nil)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndPaymentMethodID(*startDate, *endDate, paymentMethod.ID)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment method %s doesn't have any invoice between %s and %s", val, payload.StartDate, payload.EndDate), nil)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, invoices, nil)
}

func (h *Handler) handleGetInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// get invoice data
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d doesn't exists", payload.InvoiceID), nil)
		return
	}

	// get customer data
	customer, err := h.custStore.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d doesn't exists", invoice.CustomerID), nil)
		return
	}

	// get payment method data
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByID(invoice.PaymentMethodID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment name id %d doesn't exists", invoice.PaymentMethodID), nil)
		return
	}

	// get medicine item of the invoice
	medicineItem, err := h.invoiceStore.GetMedicineItem(invoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get user data, the one who inputs the invoice
	inputter, err := h.userStore.GetUserByID(invoice.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", invoice.UserID), nil)
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(invoice.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", invoice.LastModifiedByUserID), nil)
		return
	}

	returnPayload := types.InvoiceDetailPayload{
		ID:                     invoice.ID,
		Number:                 invoice.Number,
		Subtotal:               invoice.Subtotal,
		DiscountPercentage:     invoice.DiscountPercentage,
		DiscountAmount:         invoice.DiscountAmount,
		TaxPercentage:          invoice.TaxPercentage,
		TaxAmount:              invoice.TaxAmount,
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

		MedicineLists: medicineItem,
	}

	utils.WriteSuccess(w, http.StatusOK, returnPayload, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err), nil)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if invoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice id %d doesn't exist", payload.InvoiceID), nil)
		return
	}

	medicineItem, err := h.invoiceStore.GetMedicineItem(payload.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding medicine item: %v", err), nil)
		return
	}

	err = h.invoiceStore.DeleteMedicineItem(invoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	err = h.invoiceStore.DeleteInvoice(invoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	for _, medicineItem := range medicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.medStore.DeleteMedicineHistory(medData.ID, payload.InvoiceID, constants.MEDICINE_HISTORY_OUT, medicineItem.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error delete medicine history: %v", err), nil)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("invoice number %d deleted by %s", invoice.Number, user.Name), nil)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// get payment name
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.NewData.PaymentMethodName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payment name %s not found", payload.NewData.PaymentMethodName), nil)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice with id %d doesn't exists", payload.ID), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	oldMedicineItem, err := h.invoiceStore.GetMedicineItem(invoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding medicine item: %v", err), nil)
		return
	}

	newInvoice := types.Invoice{
		Number:               payload.NewData.Number,
		CustomerID:           payload.NewData.CustomerID,
		Subtotal:             payload.NewData.Subtotal,
		DiscountPercentage:   payload.NewData.DiscountPercentage,
		DiscountAmount:       payload.NewData.DiscountAmount,
		TaxPercentage:        payload.NewData.TaxPercentage,
		TaxAmount:            payload.NewData.TaxAmount,
		TotalPrice:           payload.NewData.TotalPrice,
		PaidAmount:           payload.NewData.PaidAmount,
		ChangeAmount:         payload.NewData.ChangeAmount,
		PaymentMethodID:      paymentMethod.ID,
		Description:          payload.NewData.Description,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	}

	err = h.invoiceStore.ModifyInvoice(payload.ID, newInvoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	err = h.invoiceStore.DeleteMedicineItem(invoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// reset the stock
	for _, medicineItem := range oldMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
			return
		}
	}

	// create new medicine items
	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.invoiceStore.CreateMedicineItem(types.InvoiceMedicineItem{
			InvoiceID:          payload.ID,
			MedicineID:         medData.ID,
			Qty:                medicine.Qty,
			UnitID:             unit.ID,
			Price:              medicine.Price,
			DiscountPercentage: medicine.DiscountPercentage,
			DiscountAmount:     medicine.DiscountAmount,
			Subtotal:           medicine.Subtotal,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("invoice %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err), nil)
			return
		}

		err = utils.CheckStock(medData, unit, medicine.Qty)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough", medicine.MedicineName), nil)
			return
		}
	}

	invoicePdf := types.InvoicePdfPayload{
		Number:             invoice.Number,
		UserName:           user.Name,
		Subtotal:           payload.NewData.Subtotal,
		DiscountPercentage: payload.NewData.DiscountPercentage,
		DiscountAmount:     payload.NewData.DiscountAmount,
		TaxPercentage:      payload.NewData.TaxPercentage,
		TaxAmount:          payload.NewData.TaxAmount,
		TotalPrice:         payload.NewData.TotalPrice,
		PaidAmount:         payload.NewData.PaidAmount,
		ChangeAmount:       payload.NewData.ChangeAmount,
		Description:        payload.NewData.Description,
		InvoiceDate:        *invoiceDate,
		MedicineLists:      payload.NewData.MedicineLists,
	}
	_, err = pdf.CreateInvoicePdf(invoicePdf, h.invoiceStore, invoice.PdfUrl)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update invoice pdf: %v", err), nil)
		return
	}

	// subtract the stock
	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
			return
		}

		err = h.medStore.ModifyMedicineHistoryTable(medData.ID, payload.ID, constants.MEDICINE_HISTORY_OUT, medicine.Qty, unit.ID, *invoiceDate)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update medicine history: %v", err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("invoice modified by %s", user.Name), nil)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice with id %d doesn't exists", payload.InvoiceID), nil)
		return
	}

	pdfFile := "static/pdf/invoice/" + invoice.PdfUrl

	file, err := os.Open(pdfFile)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d file not found", payload.InvoiceID), nil)
		return
	}
	defer file.Close()

	attachment := fmt.Sprintf("attachment; filename=%s", pdfFile)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, pdfFile)
}

func (h *Handler) handlePrintReceipt(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.PrintReceiptPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("invoice with id %d doesn't exists", payload.ID), nil)
		return
	}

	if invoice.ReceiptPdfUrl.Valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("receipt for this invoice has been issued"), nil)
		return
	}

	// TODO: create the receipt pdf files here

	pdfFile := "static/pdf/invoice/receipt/" + invoice.ReceiptPdfUrl.String

	// TODO: update the database with the pdf file

	file, err := os.Open(pdfFile)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d file not found", payload.ID), nil)
		return
	}
	defer file.Close()

	attachment := fmt.Sprintf("attachment; filename=%s", pdfFile)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, pdfFile)
}
