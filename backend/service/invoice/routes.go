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
	"github.com/nicolaics/pharmacon/logger"
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
		logFile, _ := logger.WriteServerErrorLog("register invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("register invoice", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again or check other devices!\nProbably you are logged-in in other device",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check customerID
	customer, err := h.custStore.GetCustomerByID(payload.CustomerID)
	if err != nil {
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "customer_id": payload.CustomerID}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get customer id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if customer == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Customer ID %d doesn't exist\nPlease create the customer first!", payload.CustomerID),
		}
		resp.WriteError(w)
		return
	}

	// check paymentMethodName
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.PaymentMethodName)
	if err != nil {
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "payment_method": payload.PaymentMethodName}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get payment method: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

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
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error create invoice: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
	}

	// get invoice id
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Number, payload.CustomerID, *invoiceDate)
	if err != nil {
		errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
		if errDel != nil {
			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
				fmt.Errorf("error get invoice id: %v\n\nerror absolute delete invoice: %v", err, errDel))

			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   (err.Error() + "\n" + errDel.Error()),
			}
			resp.WriteError(w)
			return
		}

		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get invoice id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error get medicine: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
					"medicine": medicine.MedicineName, "unit": medicine.Unit}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error get unit: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
				"medicine": medicine.MedicineName, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
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
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error create medicine item: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error create medicine item: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.CheckStock(medData, unit, medicine.Qty)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
					"medicine": medicine.MedicineName, "requested stock": medicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error check stock: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Stock for %s is not enough", medicine.MedicineName),
				Error:   err.Error(),
			}
			resp.WriteError(w)
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
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error create invoice pdf: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.invoiceStore.UpdatePdfUrl(invoiceId, invoiceFileName)
	if err != nil {
		data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error update invoice pdf url: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// reduce the stock
	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error get medicine: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
					"medicine": medicine.MedicineName, "unit": medicine.Unit}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error get unit: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
				"medicine": medicine.MedicineName, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
					"medicine": medicine.MedicineName, "requested stock": medicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error subtract stock: %v\n\nerror absolute delete invoice: %v", err, errDel))

				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate,
				"medicine": medicine.MedicineName, "requested stock": medicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error subtract stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = h.medStore.InsertIntoMedicineHistoryTable(medData.ID, invoiceId, constants.MEDICINE_HISTORY_OUT, medicine.Qty, unit.ID, *invoiceDate)
		if err != nil {
			errDel := h.invoiceStore.AbsoluteDeleteInvoice(newInvoice)
			if errDel != nil {
				data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
				logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data,
					fmt.Errorf("error insert medicine history: %v\n\nerror absolute delete invoice: %v", err, errDel))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + errDel.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"invoice_number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("register invoice", user.ID, data, fmt.Errorf("error insert medicine history: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	}

	resp := utils.Response{
		Code:    http.StatusCreated,
		Message: fmt.Sprintf("Invoice %d successfully created by %s", payload.Number, user.Name),
	}
	resp.WriteSuccess(w)
}

// beginning of invoice page, will request here
func (h *Handler) handleGetInvoiceNumberForToday(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoice number for today", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	startDate, err := utils.ParseStartDate(time.Now().Format("2006-01-02 -0700MST"))
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoice number for today", user.ID, nil, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	endDate, err := utils.ParseEndDate(time.Now().Format("2006-01-02 -0700MST"))
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoice number for today", user.ID, nil, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	numberOfInvoices, err := h.invoiceStore.GetNumberOfInvoices(*startDate, *endDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoice number for today", user.ID, nil, fmt.Errorf("error get number of invoices: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: (numberOfInvoices + 1),
	}
	resp.WriteSuccess(w)
}

// only view the purchase invoice list
func (h *Handler) handleGetInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoices", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("get invoices", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoices", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, nil, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, nil, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
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
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoice by date: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "invoice_id": val}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error parse id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		invoice, err := h.invoiceStore.GetInvoiceReturnDataByID(id)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "invoice_id": id}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoice return data by id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		invoices = append(invoices, *invoice)
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "invoice_number": val}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error parse number: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		invoices, err = h.invoiceStore.GetInvoicesByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "invoice_number": number}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoices by date and number: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "user" {
		invoices, err = h.invoiceStore.GetInvoicesByDateAndUser(*startDate, *endDate, val)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "searched_user": val}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoices by date and user: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "customer" {
		invoices, err = h.invoiceStore.GetInvoicesByDateAndCustomer(*startDate, *endDate, val)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "customer": val}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoices by date and customer: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "payment-method" {
		invoices, err = h.invoiceStore.GetInvoicesByDateAndPaymentMethod(*startDate, *endDate, val)
		if err != nil {
			data := map[string]interface{}{"start_date": *startDate, "end_date": *endDate, "payment_method": val}
			logFile, _ := logger.WriteServerErrorLog("get invoices", user.ID, data, fmt.Errorf("error get invoices by date and payment method: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Parameter undefined",
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusInternalServerError,
		Result: invoices,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get invoice detail", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("get invoice detail", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("get invoice detail", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get invoice data
	invoice, err := h.invoiceStore.GetInvoiceDetailByID(payload.InvoiceID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("get invoice detail", user.ID, data, fmt.Errorf("error get invoice by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if invoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invoice ID %d doesn't exist", payload.InvoiceID),
		}
		resp.WriteError(w)
		return
	}

	// get medicine item of the invoice
	medicineItem, err := h.invoiceStore.GetMedicineItem(invoice.ID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": invoice.ID}
		logFile, _ := logger.WriteServerErrorLog("get invoice detail", user.ID, data, fmt.Errorf("error get medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	invoice.MedicineLists = medicineItem

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: *invoice,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("delete invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("delete invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error get invoice by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if invoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invoice ID %d doesn't exist", payload.InvoiceID),
		}
		resp.WriteError(w)
		return
	}

	medicineItem, err := h.invoiceStore.GetMedicineItem(payload.InvoiceID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error get medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.invoiceStore.DeleteMedicineItem(invoice, user)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error delete medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.invoiceStore.DeleteInvoice(invoice, user)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error delete invoice: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	for _, medicineItem := range medicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.InvoiceID, "medicine": medicineItem.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error get medicine by barcode: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if medData == nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Medicine %s doesn't exist", medicineItem.MedicineName),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.InvoiceID, "medicine": medData.Name, "unit": medicineItem.Unit}
			logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = h.medStore.DeleteMedicineHistory(medData.ID, payload.InvoiceID, constants.MEDICINE_HISTORY_OUT, medicineItem.Qty, user)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.InvoiceID, "medicine": medData.Name}
			logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error delete medicine history: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.Qty, user)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.InvoiceID, "medicine": medData.Name, "qty": medicineItem.Qty}
			logFile, _ := logger.WriteServerErrorLog("delete invoice", user.ID, data, fmt.Errorf("error add stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Invoice number %d deleted by %s", invoice.Number, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("modify invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get payment name
	paymentMethod, err := h.paymentMethodStore.GetPaymentMethodByName(payload.NewData.PaymentMethodName)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID, "payment_method": payload.NewData.PaymentMethodName}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get payment method by name: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get invoice by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if invoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID, "date": payload.NewData.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	oldMedicineItem, err := h.invoiceStore.GetMedicineItem(invoice.ID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get old medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
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
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error modify invoice: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.invoiceStore.DeleteMedicineItem(invoice, user)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error delete old medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// reset the stock
	for _, medicineItem := range oldMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medicineItem.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if medData == nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Medicine %s doesn't exist", medicineItem.MedicineName),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "unit": medicineItem.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.Qty, user)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "qty": medicineItem.Qty}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error reset stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	}

	// create new medicine items
	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if medData == nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Medicine %s doesn't exist", medicine.MedicineName),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
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
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error create medicine item: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.CheckStock(medData, unit, medicine.Qty)
		if err != nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
			resp.WriteError(w)
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
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error create invoice pdf: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// subtract the stock
	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if medData == nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Medicine %s doesn't exist", medicine.MedicineName),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name, "qty": medicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error update stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = h.medStore.ModifyMedicineHistoryTable(medData.ID, payload.ID, constants.MEDICINE_HISTORY_OUT, medicine.Qty, unit.ID, *invoiceDate)
		if err != nil {
			data := map[string]interface{}{"invoice_id": payload.ID, "medicine": medData.Name}
			logFile, _ := logger.WriteServerErrorLog("modify invoice", user.ID, data, fmt.Errorf("error update medicine history: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Invoice modified by %s", user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("print invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("print invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("print invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.InvoiceID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("print invoice", user.ID, data, fmt.Errorf("error get invoice: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if invoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invoice ID %d doesn't exist", payload.InvoiceID),
		}
		resp.WriteError(w)
		return
	}

	pdfFile := constants.INVOICE_PDF_DIR_PATH + invoice.PdfUrl

	file, err := os.Open(pdfFile)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.InvoiceID}
		logFile, _ := logger.WriteServerErrorLog("print invoice", user.ID, data, fmt.Errorf("error open pdf file: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
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
		logFile, _ := logger.WriteServerErrorLog("print invoice receipt", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("print invoice receipt", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid Payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print invoice receipt", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the invoice exists
	invoice, err := h.invoiceStore.GetInvoiceByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print invoice receipt", user.ID, data, fmt.Errorf("error get invoice: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if invoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	if invoice.ReceiptPdfUrl.Valid {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Receipt for invoice number %d has been issued", invoice.Number),
		}
		resp.WriteError(w)
		return
	}

	// TODO: create the receipt pdf files here
	// TODO: update the database with the pdf file

	// TODO: add with the file name generated
	pdfFile := constants.INVOICE_RECEIPT_PDF_DIR_PATH

	file, err := os.Open(pdfFile)
	if err != nil {
		data := map[string]interface{}{"invoice_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print invoice receipt", user.ID, data, fmt.Errorf("error open receipt pdf file: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	defer file.Close()

	attachment := fmt.Sprintf("attachment; filename=%s", pdfFile)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, pdfFile)
}
