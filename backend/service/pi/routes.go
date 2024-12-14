package pi

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pharmacon/constants"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"github.com/nicolaics/pharmacon/utils/pdf"
)

type Handler struct {
	purchaseInvoiceStore types.PurchaseInvoiceStore
	userStore            types.UserStore
	supplierStore        types.SupplierStore
	medStore             types.MedicineStore
	unitStore            types.UnitStore
	poInvoiceStore       types.PurchaseOrderStore
}

func NewHandler(purchaseInvoiceStore types.PurchaseInvoiceStore, userStore types.UserStore,
	supplierStore types.SupplierStore,
	medStore types.MedicineStore, unitStore types.UnitStore, poInvoiceStore types.PurchaseOrderStore) *Handler {
	return &Handler{
		purchaseInvoiceStore: purchaseInvoiceStore,
		userStore:            userStore,
		supplierStore:        supplierStore,
		medStore:             medStore,
		unitStore:            unitStore,
		poInvoiceStore:       poInvoiceStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice/purchase", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase/{params}/{val}", h.handleGetPurchaseInvoices).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase/detail", h.handleGetPurchaseInvoiceDetail).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice/purchase", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/invoice/purchase/print", h.handlePrint).Methods(http.MethodPost)

	router.HandleFunc("/invoice/purchase", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase/print", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterPurchaseInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check supplierID
	supplier, err := h.supplierStore.GetSupplierByID(payload.SupplierID)
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "supplier_id": payload.SupplierID}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get supplier by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
	}
	if supplier == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Supplier ID %d doesn't exist", payload.SupplierID),
		}
		resp.WriteError(w)
		return
	}

	// get purchase order
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByNumber(payload.PurchaseOrderNumber)
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "po_number": payload.PurchaseOrderNumber}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get po by number: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseOrder == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase order %d doesn't exist", payload.PurchaseOrderNumber),
		}
		resp.WriteError(w)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check duplicate
	purchaseInvoiceId, err := h.purchaseInvoiceStore.GetPurchaseInvoiceID(payload.Number, payload.SupplierID, payload.Subtotal, payload.TotalPrice, *invoiceDate)
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get pi id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseInvoiceId != 0 {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase invoice number %d already exists", payload.Number),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.CreatePurchaseInvoice(types.PurchaseInvoice{
		Number:               payload.Number,
		SupplierID:           payload.SupplierID,
		PurchaseOrderNumber:  payload.PurchaseOrderNumber,
		Subtotal:             payload.Subtotal,
		DiscountPercentage:   payload.DiscountPercentage,
		DiscountAmount:       payload.DiscountAmount,
		TaxPercentage:        payload.TaxPercentage,
		TaxAmount:            payload.TaxAmount,
		TotalPrice:           payload.TotalPrice,
		Description:          payload.Description,
		UserID:               user.ID,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error create pi: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get purchaseInvoiceID
	purchaseInvoiceId, err = h.purchaseInvoiceStore.GetPurchaseInvoiceID(payload.Number, payload.SupplierID, payload.Subtotal, payload.TotalPrice, *invoiceDate)
	if err != nil {
		delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
			Number:              payload.Number,
			SupplierID:          payload.SupplierID,
			PurchaseOrderNumber: payload.PurchaseOrderNumber,
			Subtotal:            payload.Subtotal,
			DiscountPercentage:  payload.DiscountPercentage,
			DiscountAmount:      payload.DiscountAmount,
			TaxPercentage:       payload.TaxPercentage,
			TaxAmount:           payload.TaxAmount,
			TotalPrice:          payload.TotalPrice,
			Description:         payload.Description,
			InvoiceDate:         *invoiceDate,
		})
		if delErr != nil {
			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get pi id: %v\nerror absolute delete: %v", err, delErr))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   (err.Error() + "\n" + delErr.Error()),
			}
			resp.WriteError(w)
			return
		}

		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get pi id: %v", err))
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
			delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
				Number:              payload.Number,
				SupplierID:          payload.SupplierID,
				PurchaseOrderNumber: payload.PurchaseOrderNumber,
				Subtotal:            payload.Subtotal,
				DiscountPercentage:  payload.DiscountPercentage,
				DiscountAmount:      payload.DiscountAmount,
				TaxPercentage:       payload.TaxPercentage,
				TaxAmount:           payload.TaxAmount,
				TotalPrice:          payload.TotalPrice,
				Description:         payload.Description,
				InvoiceDate:         *invoiceDate,
			})
			if delErr != nil {
				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get medicine: %v\nerror absolute delete: %v", err, delErr))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + delErr.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
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
			delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
				Number:              payload.Number,
				SupplierID:          payload.SupplierID,
				PurchaseOrderNumber: payload.PurchaseOrderNumber,
				Subtotal:            payload.Subtotal,
				DiscountPercentage:  payload.DiscountPercentage,
				DiscountAmount:      payload.DiscountAmount,
				TaxPercentage:       payload.TaxPercentage,
				TaxAmount:           payload.TaxAmount,
				TotalPrice:          payload.TotalPrice,
				Description:         payload.Description,
				InvoiceDate:         *invoiceDate,
			})
			if delErr != nil {
				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "unit": medicine.Unit}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get unit: %v\nerror absolute delete: %v", err, delErr))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + delErr.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		expDate, err := utils.ParseDate(medicine.ExpDate)
		if err != nil {
			delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
				Number:              payload.Number,
				SupplierID:          payload.SupplierID,
				PurchaseOrderNumber: payload.PurchaseOrderNumber,
				Subtotal:            payload.Subtotal,
				DiscountPercentage:  payload.DiscountPercentage,
				DiscountAmount:      payload.DiscountAmount,
				TaxPercentage:       payload.TaxPercentage,
				TaxAmount:           payload.TaxAmount,
				TotalPrice:          payload.TotalPrice,
				Description:         payload.Description,
				InvoiceDate:         *invoiceDate,
			})
			if delErr != nil {
				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "exp_date": medicine.ExpDate}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error parse exp date: %v\nerror absolute delete: %v", err, delErr))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + delErr.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "exp_date": medicine.ExpDate}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error parse exp date: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = h.purchaseInvoiceStore.CreatePurchaseMedicineItem(types.PurchaseMedicineItem{
			PurchaseInvoiceID:  purchaseInvoiceId,
			MedicineID:         medData.ID,
			Qty:                medicine.Qty,
			UnitID:             unit.ID,
			Price:              medicine.Price,
			DiscountPercentage: payload.DiscountPercentage,
			DiscountAmount:     payload.DiscountAmount,
			TaxPercentage:      payload.TaxPercentage,
			TaxAmount:          payload.TaxAmount,
			Subtotal:           medicine.Subtotal,
			BatchNumber:        medicine.BatchNumber,
			ExpDate:            *expDate,
		})
		if err != nil {
			delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
				Number:              payload.Number,
				SupplierID:          payload.SupplierID,
				PurchaseOrderNumber: payload.PurchaseOrderNumber,
				Subtotal:            payload.Subtotal,
				DiscountPercentage:  payload.DiscountPercentage,
				DiscountAmount:      payload.DiscountAmount,
				TaxPercentage:       payload.TaxPercentage,
				TaxAmount:           payload.TaxAmount,
				TotalPrice:          payload.TotalPrice,
				Description:         payload.Description,
				InvoiceDate:         *invoiceDate,
			})
			if delErr != nil {
				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data,
					fmt.Errorf("error create medicine item: %v\nerror absolute delete: %v", err, delErr))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + delErr.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error create medicine item: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// update stock
		err = utils.AddStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
				Number:              payload.Number,
				SupplierID:          payload.SupplierID,
				PurchaseOrderNumber: payload.PurchaseOrderNumber,
				Subtotal:            payload.Subtotal,
				DiscountPercentage:  payload.DiscountPercentage,
				DiscountAmount:      payload.DiscountAmount,
				TaxPercentage:       payload.TaxPercentage,
				TaxAmount:           payload.TaxAmount,
				TotalPrice:          payload.TotalPrice,
				Description:         payload.Description,
				InvoiceDate:         *invoiceDate,
			})
			if delErr != nil {
				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "qty": medicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data,
					fmt.Errorf("error add stock: %v\nerror absolute delete: %v", err, delErr))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   (err.Error() + "\n" + delErr.Error()),
				}
				resp.WriteError(w)
				return
			}

			data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "qty": medicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error add stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// update received qty
		if payload.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, payload.PurchaseOrderNumber, medData, medicine.Qty, unit, user, 1)
			if err != nil {
				delErr := h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
					Number:              payload.Number,
					SupplierID:          payload.SupplierID,
					PurchaseOrderNumber: payload.PurchaseOrderNumber,
					Subtotal:            payload.Subtotal,
					DiscountPercentage:  payload.DiscountPercentage,
					DiscountAmount:      payload.DiscountAmount,
					TaxPercentage:       payload.TaxPercentage,
					TaxAmount:           payload.TaxAmount,
					TotalPrice:          payload.TotalPrice,
					Description:         payload.Description,
					InvoiceDate:         *invoiceDate,
				})
				if delErr != nil {
					data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "qty": medicine.Qty}
					logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data,
						fmt.Errorf("error update received qty: %v\nerror absolute delete: %v", err, delErr))
					resp := utils.Response{
						Code:    http.StatusInternalServerError,
						Message: "Internal server error",
						Log:     logFile,
						Error:   (err.Error() + "\n" + delErr.Error()),
					}
					resp.WriteError(w)
					return
				}

				data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate, "medicine": medData.Name, "qty": medicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error update received qty: %v", err))
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
	}

	purchaseInvoicePdf := types.PurchaseInvoicePdfPayload{
		Number:             payload.Number,
		Subtotal:           payload.Subtotal,
		DiscountPercentage: payload.DiscountPercentage,
		DiscountAmount:     payload.DiscountAmount,
		TaxPercentage:      payload.TaxPercentage,
		TaxAmount:          payload.TaxAmount,
		TotalPrice:         payload.TotalPrice,
		Description:        payload.Description,
		InvoiceDate:        *invoiceDate,

		Supplier: struct {
			Name                string "json:\"name\""
			Address             string "json:\"address\""
			CompanyPhoneNumber  string "json:\"companyPhoneNumber\""
			ContactPersonName   string "json:\"contactPersonName\""
			ContactPersonNumber string "json:\"contactPersonNumber\""
			Terms               string "json:\"terms\""
			VendorIsTaxable     bool   "json:\"vendorIsTaxable\""
		}{
			Name:                supplier.Name,
			Address:             supplier.Address,
			CompanyPhoneNumber:  supplier.CompanyPhoneNumber,
			ContactPersonName:   supplier.ContactPersonName,
			ContactPersonNumber: supplier.ContactPersonNumber,
			Terms:               supplier.Terms,
			VendorIsTaxable:     supplier.VendorIsTaxable,
		},

		UserName: user.Name,

		PurchaseOrderNumber: purchaseOrder.Number,
		PurchaseOrderDate:   purchaseOrder.InvoiceDate,

		MedicineLists: payload.MedicineLists,
	}

	// create pdf
	fileName, err := pdf.CreatePurchaseInvoicePdf(h.purchaseInvoiceStore, purchaseInvoicePdf, "")
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error create pi pdf: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.UpdatePdfUrl(purchaseInvoiceId, fileName)
	if err != nil {
		data := map[string]interface{}{"number": payload.Number, "date": payload.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("register purchase invoice", user.ID, data, fmt.Errorf("error update pi pdf url: %v", err))
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
		Code:    http.StatusCreated,
		Message: fmt.Sprintf("Purchase invoice %d successfully created by %s", payload.Number, user.Name),
	}
	resp.WriteSuccess(w)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get purchase invoices", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get purchase invoices", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoices", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error parse date: %v", err))
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
		data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	var purchaseInvoices []types.PurchaseInvoiceListsReturnPayload

	if val == "all" {
		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDate(*startDate, *endDate)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get pi by date: %v", err))
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
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "id": val}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error parse pi id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceListByID(id)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "id": id}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get pi by id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		purchaseInvoices = append(purchaseInvoices, *purchaseInvoice)
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "number": val}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error parse pi number: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "number": number}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get pi by date and number: %v", err))
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
		users, err := h.userStore.GetUserBySearchName(val)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "searched_user": val}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get user by search name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		for _, user := range users {
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndUserID(*startDate, *endDate, user.ID)
			if err != nil {
				data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "searched_user": val}
				logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get user by search name: %v", err))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   err.Error(),
				}
				resp.WriteError(w)
				return
			}

			purchaseInvoices = append(purchaseInvoices, temp...)
		}
	} else if params == "supplier" {
		suppliers, err := h.supplierStore.GetSupplierBySearchName(val)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "supplier": val}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get supplier by search name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		for _, supplier := range suppliers {
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndSupplierID(*startDate, *endDate, supplier.ID)
			if err != nil {
				data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "supplier": val}
				logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get supplier by search name: %v", err))
				resp := utils.Response{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Log:     logFile,
					Error:   err.Error(),
				}
				resp.WriteError(w)
				return
			}

			purchaseInvoices = append(purchaseInvoices, temp...)
		}
	} else if params == "purchase-order" {
		poiNumber, err := strconv.Atoi(val)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "purchase_order_number": val}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error parse po number: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndPONumber(*startDate, *endDate, poiNumber)
		if err != nil {
			data := map[string]interface{}{"start_date": payload.StartDate, "end_date": payload.EndDate, "purchase_order_number": poiNumber}
			logFile, _ := logger.WriteServerErrorLog("get purchase invoices", user.ID, data, fmt.Errorf("error get pi by date and po number: %v", err))
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
		Code:   http.StatusOK,
		Result: purchaseInvoices,
	}
	resp.WriteSuccess(w)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get purchase invoice detail", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get purchase invoice detail", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoice detail", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get purchase invoice data
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceDetailByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoice detail", user.ID, data, fmt.Errorf("get pi by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseInvoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	// get medicine item of the purchase invoice
	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get purchase invoice detail", user.ID, data, fmt.Errorf("get medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	purchaseInvoice.MedicineLists = purchaseMedicineItem

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: *purchaseInvoice,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePurchaseInvoice

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error get pi by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseInvoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error get medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItem(purchaseInvoice, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error delete medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// subtract stock and received qty
	for _, purchaseMedicine := range purchaseMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(purchaseMedicine.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": purchaseMedicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, purchaseMedicine.Qty, user)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit, "qty": purchaseMedicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error subtract stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, purchaseMedicine.Qty, unit, user, 0)
			if err != nil {
				data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit, "qty": purchaseMedicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error update received qty: %v", err))
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
	}

	err = h.purchaseInvoiceStore.DeletePurchaseInvoice(purchaseInvoice, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete purchase invoice", user.ID, data, fmt.Errorf("error delete pi: %v", err))
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
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Purchase invoice number %d deleted by %s", purchaseInvoice.Number, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPurchaseInvoicePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get pi by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseInvoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	// check supplier
	supplier, err := h.supplierStore.GetSupplierByID(payload.NewData.SupplierID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "supplier": payload.NewData.SupplierID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get supplier by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if supplier == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Supplier ID %d doesn't exist", payload.NewData.SupplierID),
		}
		resp.WriteError(w)
		return
	}

	// check purchase order
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByNumber(payload.NewData.PurchaseOrderNumber)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "po_number": purchaseInvoice.PurchaseOrderNumber}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get po by number: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseOrder == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase order number %d doesn't exist", payload.NewData.PurchaseOrderNumber),
		}
		resp.WriteError(w)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "date": payload.NewData.InvoiceDate}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.ModifyPurchaseInvoice(payload.ID, types.PurchaseInvoice{
		Number:               payload.NewData.Number,
		SupplierID:           payload.NewData.SupplierID,
		Subtotal:             payload.NewData.Subtotal,
		DiscountPercentage:   payload.NewData.DiscountPercentage,
		DiscountAmount:       payload.NewData.DiscountAmount,
		TaxPercentage:        payload.NewData.TaxPercentage,
		TaxAmount:            payload.NewData.TaxAmount,
		TotalPrice:           payload.NewData.TotalPrice,
		Description:          payload.NewData.Description,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error modify pi: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItem(purchaseInvoice, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error delete medicine item: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// subtract the stock and received qty
	for _, purchaseMedicine := range purchaseMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(purchaseMedicine.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": purchaseMedicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
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
				Message: fmt.Sprintf("Medicine %s doesn't exist", purchaseMedicine.MedicineName),
			}
			resp.WriteError(w)
			return
		}

		unit, err := h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, purchaseMedicine.Qty, user)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit, "qty": purchaseMedicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error subtract stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, purchaseMedicine.Qty, unit, user, 0)
			if err != nil {
				data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": purchaseMedicine.Unit, "qty": purchaseMedicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error update received qty: %v", err))
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
	}

	purchaseInvoicePdf := types.PurchaseInvoicePdfPayload{
		Number:             payload.NewData.Number,
		Subtotal:           payload.NewData.Subtotal,
		DiscountPercentage: payload.NewData.DiscountPercentage,
		DiscountAmount:     payload.NewData.DiscountAmount,
		TaxPercentage:      payload.NewData.TaxPercentage,
		TaxAmount:          payload.NewData.TaxAmount,
		TotalPrice:         payload.NewData.TotalPrice,
		Description:        payload.NewData.Description,
		InvoiceDate:        *invoiceDate,

		Supplier: struct {
			Name                string "json:\"name\""
			Address             string "json:\"address\""
			CompanyPhoneNumber  string "json:\"companyPhoneNumber\""
			ContactPersonName   string "json:\"contactPersonName\""
			ContactPersonNumber string "json:\"contactPersonNumber\""
			Terms               string "json:\"terms\""
			VendorIsTaxable     bool   "json:\"vendorIsTaxable\""
		}{
			Name:                supplier.Name,
			Address:             supplier.Address,
			CompanyPhoneNumber:  supplier.CompanyPhoneNumber,
			ContactPersonName:   supplier.ContactPersonName,
			ContactPersonNumber: supplier.ContactPersonNumber,
			Terms:               supplier.Terms,
			VendorIsTaxable:     supplier.VendorIsTaxable,
		},

		UserName: user.Name,

		PurchaseOrderNumber: purchaseOrder.Number,
		PurchaseOrderDate:   purchaseOrder.InvoiceDate,

		MedicineLists: payload.NewData.MedicineLists,
	}

	// create pdf
	fileName, err := pdf.CreatePurchaseInvoicePdf(h.purchaseInvoiceStore, purchaseInvoicePdf, purchaseInvoice.PdfURL)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error create pdf: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.purchaseInvoiceStore.UpdatePdfUrl(purchaseInvoice.ID, fileName)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error update pdf url: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medicine.MedicineName}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get medicine: %v", err))
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
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error get unit: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		expDate, err := utils.ParseDate(medicine.ExpDate)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "exp_date": medicine.ExpDate}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error parse exp date: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		err = h.purchaseInvoiceStore.CreatePurchaseMedicineItem(types.PurchaseMedicineItem{
			PurchaseInvoiceID:  payload.ID,
			MedicineID:         medData.ID,
			Qty:                medicine.Qty,
			UnitID:             unit.ID,
			Price:              medicine.Price,
			DiscountPercentage: medicine.DiscountPercentage,
			DiscountAmount:     medicine.DiscountAmount,
			TaxPercentage:      medicine.TaxPercentage,
			TaxAmount:          medicine.TaxAmount,
			Subtotal:           medicine.Subtotal,
			BatchNumber:        medicine.BatchNumber,
			ExpDate:            *expDate,
		})
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error create medicine item: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// add the stock with the new value
		err = utils.AddStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit, "qty": medicine.Qty}
			logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error add stock: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, medicine.Qty, unit, user, 1)
			if err != nil {
				data := map[string]interface{}{"id": payload.ID, "medicine": medData.Name, "unit": medicine.Unit, "qty": medicine.Qty}
				logFile, _ := logger.WriteServerErrorLog("modify purchase invoice", user.ID, data, fmt.Errorf("error update received qty: %v", err))
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
	}

	resp := utils.Response{
		Code:    http.StatusOK,
		Message: fmt.Sprintf("Purchase invoice ID %d modified by %s", purchaseInvoice.ID, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoiceDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("print purchase invoice", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("print purchase invoice", 0, nil, fmt.Errorf("invalid payload: %v", err))
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
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print purchase invoice", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print purchase invoice", user.ID, data, fmt.Errorf("error get pi by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if purchaseInvoice == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Purchase invoice ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	pdfFile := constants.PI_PDF_DIR_PATH + purchaseInvoice.PdfURL

	file, err := os.Open(pdfFile)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("print purchase invoice", user.ID, data, fmt.Errorf("error open pdf file: %v", err))
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

/*
 * req_type = 0: subtract
 * req_type = 1: add
 */
func updateReceivedQty(h *Handler, poinn int, medData *types.Medicine, addQty float64, receivedPurchasedUnit *types.Unit, user *types.User, req_type int) error {
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByNumber(poinn)
	if err != nil {
		return fmt.Errorf("purchase order invoice %d not found: %v", poinn, err)
	}

	purchaseOrderMeds, err := h.poInvoiceStore.GetPurchaseOrderItem(purchaseOrder.ID)
	if err != nil {
		return fmt.Errorf("purchase order item not found: %v", err)
	}

	for _, purchaseOrderMed := range purchaseOrderMeds {
		medPurchaseData, err := h.medStore.GetMedicineByBarcode(purchaseOrderMed.MedicineBarcode)
		if err != nil {
			return fmt.Errorf("medicine %s doesn't exists", purchaseOrderMed.MedicineName)
		}

		if medPurchaseData.ID == medData.ID {
			poUnit, err := h.unitStore.GetUnitByName(purchaseOrderMed.Unit)
			if err != nil {
				return fmt.Errorf("po unit error: %v", err)
			}

			if req_type == 0 {
				if purchaseOrderMed.ReceivedQty == 0 {
					return nil
				}

				err = subtractReceivedQty(h, medData, &purchaseOrderMed, addQty, poUnit, receivedPurchasedUnit, purchaseOrder.ID, user)
			} else {
				// update received qty
				err = addReceivedQty(h, medData, &purchaseOrderMed, addQty, poUnit, receivedPurchasedUnit, purchaseOrder.ID, user)
			}
			if err != nil {
				return fmt.Errorf("update received qty error: %v", err)
			}

			return nil
		}
	}
	return nil
	// return fmt.Errorf("medicine not found in purchase order invoice")
}

func addReceivedQty(h *Handler, medData *types.Medicine, purchaseOrderMed *types.PurchaseOrderItemReturn, additionalReceivedQty float64,
	poUnit *types.Unit, purchasedUnit *types.Unit, poiid int, user *types.User) error {
	var purchasedOrderQty float64
	var purchasedReceivedQty float64
	var updatedQty float64

	if medData.FirstUnitID == poUnit.ID {
		purchasedOrderQty = purchaseOrderMed.OrderQty
		purchasedReceivedQty = purchaseOrderMed.ReceivedQty
	} else if medData.SecondUnitID == poUnit.ID {
		purchasedOrderQty = (purchaseOrderMed.OrderQty * medData.SecondUnitToFirstUnitRatio)
		purchasedReceivedQty = (purchaseOrderMed.ReceivedQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == poUnit.ID {
		purchasedOrderQty = (purchaseOrderMed.OrderQty * medData.ThirdUnitToFirstUnitRatio)
		purchasedReceivedQty = (purchaseOrderMed.ReceivedQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	if medData.FirstUnitID == purchasedUnit.ID {
		updatedQty = (additionalReceivedQty + purchasedReceivedQty)
	} else if medData.SecondUnitID == purchasedUnit.ID {
		updatedQty = (additionalReceivedQty * medData.SecondUnitToFirstUnitRatio) + purchasedReceivedQty
	} else if medData.ThirdUnitID == purchasedUnit.ID {
		updatedQty = (additionalReceivedQty * medData.ThirdUnitToFirstUnitRatio) + purchasedReceivedQty
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	if updatedQty > purchasedOrderQty {
		return fmt.Errorf("received medicine is larger than ordered")
	}

	err := h.poInvoiceStore.UpdtaeReceivedQty(poiid, updatedQty, user, medData.ID)
	if err != nil {
		return fmt.Errorf("update error: %v", err)
	}

	return nil
}

func subtractReceivedQty(h *Handler, medData *types.Medicine, purchaseOrderMed *types.PurchaseOrderItemReturn, additionalReceivedQty float64,
	poUnit *types.Unit, purchasedUnit *types.Unit, poiid int, user *types.User) error {
	var purchasedReceivedQty float64
	var updatedQty float64

	if medData.FirstUnitID == poUnit.ID {
		purchasedReceivedQty = purchaseOrderMed.ReceivedQty
	} else if medData.SecondUnitID == poUnit.ID {
		purchasedReceivedQty = (purchaseOrderMed.ReceivedQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == poUnit.ID {
		purchasedReceivedQty = (purchaseOrderMed.ReceivedQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	if medData.FirstUnitID == purchasedUnit.ID {
		updatedQty = (purchasedReceivedQty - additionalReceivedQty)
	} else if medData.SecondUnitID == purchasedUnit.ID {
		updatedQty = purchasedReceivedQty - (additionalReceivedQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == purchasedUnit.ID {
		updatedQty = purchasedReceivedQty - (additionalReceivedQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	if updatedQty < 0 {
		return fmt.Errorf("received medicine is smaller than 0")
	}

	err := h.poInvoiceStore.UpdtaeReceivedQty(poiid, updatedQty, user, medData.ID)
	if err != nil {
		return err
	}

	return nil
}
