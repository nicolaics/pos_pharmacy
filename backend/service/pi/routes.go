package pi

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

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

	// check supplierID
	supplier, err := h.supplierStore.GetSupplierByID(payload.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d not found", payload.SupplierID))
		return
	}

	// get purchase order
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByNumber(payload.PurchaseOrderNumber)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("po number %d not found", payload.PurchaseOrderNumber))
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	// check duplicate
	purchaseInvoiceId, err := h.purchaseInvoiceStore.GetPurchaseInvoiceID(payload.Number, payload.SupplierID, payload.Subtotal, payload.TotalPrice, *invoiceDate)
	if err == nil || purchaseInvoiceId != 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice number %d exists", payload.Number))
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
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get purchaseInvoiceID
	purchaseInvoiceId, err = h.purchaseInvoiceStore.GetPurchaseInvoiceID(payload.Number, payload.SupplierID, payload.Subtotal, payload.TotalPrice, *invoiceDate)
	if err != nil {
		err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
			return
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice number %d doesn't exists", payload.Number))
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
				return
			}

			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
				if err != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		expDate, err := utils.ParseDate(medicine.ExpDate)
		if err != nil {
			err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
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
			err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}

		// update stock
		err = utils.AddStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}

		// update received qty
		if payload.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, payload.PurchaseOrderNumber, medData, medicine.Qty, unit, user, 1)
			if err != nil {
				err = h.purchaseInvoiceStore.AbsoluteDeletePurchaseInvoice(types.PurchaseInvoice{
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
				if err != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete purchase invoice: %v", err))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update received qty: %v", err))
				return
			}
		}
	}

	purchaseInvoicePdf := types.PurchaseInvoicePDFPayload{
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
	fileName, err := pdf.CreatePurchaseInvoicePDF(h.purchaseInvoiceStore, purchaseInvoicePdf, "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating pdf: %v", err))
		return
	}

	err = h.purchaseInvoiceStore.UpdatePDFUrl(purchaseInvoiceId, fileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update pdf url: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase invoice %d successfully created by %s", payload.Number, user.Name))
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoicePayload

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

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

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

	var purchaseInvoices []types.PurchaseInvoiceListsReturnPayload

	if val == "all" {
		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDate(*startDate, *endDate)
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

		purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice id %d not exist", id))
			return
		}

		supplier, err := h.supplierStore.GetSupplierByID(purchaseInvoice.SupplierID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("supplier id %d not found", purchaseInvoice.SupplierID))
			return
		}

		user, err := h.userStore.GetUserByID(purchaseInvoice.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user id %d not found", purchaseInvoice.UserID))
			return
		}

		purchaseInvoices = append(purchaseInvoices, types.PurchaseInvoiceListsReturnPayload{
			ID:           purchaseInvoice.ID,
			Number:       purchaseInvoice.Number,
			SupplierName: supplier.Name,
			TotalPrice:   purchaseInvoice.TotalPrice,
			Description:  purchaseInvoice.Description,
			UserName:     user.Name,
			InvoiceDate:  purchaseInvoice.InvoiceDate,
		})
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "user" {
		users, err := h.userStore.GetUserBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val))
			return
		}

		for _, user := range users {
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndUserID(*startDate, *endDate, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s doesn't create any purchase invoice between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			purchaseInvoices = append(purchaseInvoices, temp...)
		}
	} else if params == "supplier" {
		suppliers, err := h.supplierStore.GetSupplierBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier %s not exists", val))
			return
		}

		for _, supplier := range suppliers {
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndSupplierID(*startDate, *endDate, supplier.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier %s doesn't create any purchase invoice between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			purchaseInvoices = append(purchaseInvoices, temp...)
		}
	} else if params == "purchase-order" {
		poiNumber, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndPONumber(*startDate, *endDate, poiNumber)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, purchaseInvoices)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseInvoiceDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoiceDetailPayload

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

	// get purchase invoice data
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice id %d doesn't exists", payload.ID))
		return
	}

	// get medicine item of the purchase invoice
	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get supplier data
	supplier, err := h.supplierStore.GetSupplierByID(purchaseInvoice.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d doesn't exists", purchaseInvoice.SupplierID))
		return
	}

	// get user data, the one who inputs the purchase invoice
	inputter, err := h.userStore.GetUserByID(purchaseInvoice.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseInvoice.UserID))
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(purchaseInvoice.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseInvoice.LastModifiedByUserID))
		return
	}

	returnPayload := types.PurchaseInvoiceDetailPayload{
		ID:                     purchaseInvoice.ID,
		Number:                 purchaseInvoice.Number,
		Subtotal:               purchaseInvoice.Subtotal,
		DiscountPercentage:     purchaseInvoice.DiscountPercentage,
		DiscountAmount:         purchaseInvoice.DiscountAmount,
		TaxPercentage:          purchaseInvoice.TaxPercentage,
		TaxAmount:              purchaseInvoice.TaxAmount,
		TotalPrice:             purchaseInvoice.TotalPrice,
		Description:            purchaseInvoice.Description,
		InvoiceDate:            purchaseInvoice.InvoiceDate,
		CreatedAt:              purchaseInvoice.CreatedAt,
		LastModified:           purchaseInvoice.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,
		PdfURL:                 purchaseInvoice.PdfURL,

		Supplier: struct {
			ID                  int    "json:\"id\""
			Name                string "json:\"name\""
			Address             string "json:\"address\""
			CompanyPhoneNumber  string "json:\"companyPhoneNumber\""
			ContactPersonName   string "json:\"contactPersonName\""
			ContactPersonNumber string "json:\"contactPersonNumber\""
			Terms               string "json:\"terms\""
			VendorIsTaxable     bool   "json:\"vendorIsTaxable\""
		}{
			ID:                  supplier.ID,
			Name:                supplier.Name,
			Address:             supplier.Address,
			CompanyPhoneNumber:  supplier.CompanyPhoneNumber,
			ContactPersonName:   supplier.ContactPersonName,
			ContactPersonNumber: supplier.ContactPersonNumber,
			Terms:               supplier.Terms,
			VendorIsTaxable:     supplier.VendorIsTaxable,
		},

		User: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   inputter.ID,
			Name: inputter.Name,
		},

		MedicineLists: purchaseMedicineItem,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePurchaseInvoice

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
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if purchaseInvoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice id %d doesn't exist", payload.ID))
		return
	}

	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("purchase medicine item don't exist: %v", err))
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItem(purchaseInvoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// subtract stock and received qty
	for _, purchaseMedicine := range purchaseMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(purchaseMedicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", purchaseMedicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(purchaseMedicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, purchaseMedicine.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, purchaseMedicine.Qty, unit, user, 0)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update received qty: %v", err))
				return
			}
		}
	}

	err = h.purchaseInvoiceStore.DeletePurchaseInvoice(purchaseInvoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("purchase invoice number %d deleted by %s", purchaseInvoice.Number, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPurchaseInvoicePayload

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

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice with id %d doesn't exists", payload.ID))
		return
	}

	// check supplier
	supplier, err := h.supplierStore.GetSupplierByID(payload.NewData.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d not found", payload.NewData.SupplierID))
		return
	}

	// check purchase order
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByNumber(payload.NewData.PurchaseOrderNumber)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order number %d not exist", payload.NewData.PurchaseOrderNumber))
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
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
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	purchaseMedicineItem, err := h.purchaseInvoiceStore.GetPurchaseMedicineItem(purchaseInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("purchase medicine item don't exist: %v", err))
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItem(purchaseInvoice, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// subtract the stock and received qty
	for _, purchaseMedicine := range purchaseMedicineItem {
		medData, err := h.medStore.GetMedicineByBarcode(purchaseMedicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", purchaseMedicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(purchaseMedicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(purchaseMedicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = utils.SubtractStock(h.medStore, medData, unit, purchaseMedicine.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, purchaseMedicine.Qty, unit, user, 0)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update received qty: %v", err))
				return
			}
		}

		purchaseInvoicePdf := types.PurchaseInvoicePDFPayload{
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
		fileName, err := pdf.CreatePurchaseInvoicePDF(h.purchaseInvoiceStore, purchaseInvoicePdf, purchaseInvoice.PdfURL)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating pdf: %v", err))
			return
		}

		err = h.purchaseInvoiceStore.UpdatePDFUrl(purchaseInvoice.ID, fileName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update pdf url: %v", err))
			return
		}
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

		expDate, err := utils.ParseDate(medicine.ExpDate)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
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
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase invoice %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err))
			return
		}

		// add the stock with the new value
		err = utils.AddStock(h.medStore, medData, unit, medicine.Qty, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}

		// update received qty
		if purchaseInvoice.PurchaseOrderNumber != 0 {
			err = updateReceivedQty(h, purchaseInvoice.PurchaseOrderNumber, medData, medicine.Qty, unit, user, 1)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update received qty: %v", err))
				return
			}
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase invoice modified by %s", user.Name))
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseInvoiceDetailPayload

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

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice with id %d doesn't exists", payload.ID))
		return
	}

	pdfFile := "static/pdf/purchase-invoice/" + purchaseInvoice.PdfURL

	file, err := os.Open(pdfFile)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice id %d file not found", payload.ID))
		return
	}
	defer file.Close()

	attachment := fmt.Sprintf("attachment; filename=%s", pdfFile)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, pdfFile)
}

// req_type == 0, means subtract
// req_typ == 1, means add
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
