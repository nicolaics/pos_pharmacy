package poi

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"github.com/nicolaics/pharmacon/utils/pdf"
)

type Handler struct {
	poInvoiceStore types.PurchaseOrderStore
	userStore      types.UserStore
	supplierStore  types.SupplierStore
	medStore       types.MedicineStore
	unitStore      types.UnitStore
}

func NewHandler(poInvoiceStore types.PurchaseOrderStore, userStore types.UserStore,
	supplierStore types.SupplierStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		poInvoiceStore: poInvoiceStore,
		userStore:      userStore,
		supplierStore:  supplierStore,
		medStore:       medStore,
		unitStore:      unitStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice/purchase-order", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase-order", h.handleGetPOnvoiceNumberForToday).Methods(http.MethodGet)
	router.HandleFunc("/invoice/purchase-order/{params}/{val}", h.handleGetPurchaseOrders).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase-order/detail", h.handleGetPurchaseOrderDetail).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase-order", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice/purchase-order", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/invoice/purchase-order/print", h.handlePrint).Methods(http.MethodPost)

	router.HandleFunc("/invoice/purchase-order", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase-order/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase-order/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase-order/print", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterPurchaseOrderPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check supplierID
	supplier, err := h.supplierStore.GetSupplierByID(payload.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d not found", payload.SupplierID), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.InvoiceDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("error parse date: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	// check duplicate
	purchaseOrderId, err := h.poInvoiceStore.GetPurchaseOrderID(payload.Number, payload.SupplierID, payload.TotalItem, *invoiceDate)
	if err == nil || purchaseOrderId != 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order invoice number %d exists", payload.Number), nil)
		return
	}

	err = h.poInvoiceStore.CreatePurchaseOrder(types.PurchaseOrder{
		Number:               payload.Number,
		SupplierID:           payload.SupplierID,
		UserID:               user.ID,
		TotalItem:            payload.TotalItem,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get purchaseInvoice ID
	purchaseOrderId, err = h.poInvoiceStore.GetPurchaseOrderID(payload.Number, payload.SupplierID, payload.TotalItem, *invoiceDate)
	if err != nil {
		delErr := h.poInvoiceStore.AbsoluteDeletePurchaseOrder(types.PurchaseOrder{
			Number:      payload.Number,
			SupplierID:  payload.SupplierID,
			UserID:      user.ID,
			TotalItem:   payload.TotalItem,
			InvoiceDate: *invoiceDate,
		})
		if delErr != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete po invoice: %v", delErr), nil)
			return
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order invoice number %d doesn't exists: %v", payload.Number, err), nil)
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			delErr := h.poInvoiceStore.AbsoluteDeletePurchaseOrder(types.PurchaseOrder{
				Number:      payload.Number,
				SupplierID:  payload.SupplierID,
				UserID:      user.ID,
				TotalItem:   payload.TotalItem,
				InvoiceDate: *invoiceDate,
			})
			if delErr != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete po invoice: %v", delErr), nil)
				return
			}

			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			delErr := h.poInvoiceStore.AbsoluteDeletePurchaseOrder(types.PurchaseOrder{
				Number:      payload.Number,
				SupplierID:  payload.SupplierID,
				UserID:      user.ID,
				TotalItem:   payload.TotalItem,
				InvoiceDate: *invoiceDate,
			})
			if delErr != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete po invoice: %v", delErr), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.poInvoiceStore.CreatePurchaseOrderItem(types.PurchaseOrderItem{
			PurchaseOrderID: purchaseOrderId,
			MedicineID:      medData.ID,
			OrderQty:        medicine.OrderQty,
			ReceivedQty:     medicine.ReceivedQty,
			UnitID:          unit.ID,
			Remarks:         medicine.Remarks,
		})
		if err != nil {
			delErr := h.poInvoiceStore.AbsoluteDeletePurchaseOrder(types.PurchaseOrder{
				Number:      payload.Number,
				SupplierID:  payload.SupplierID,
				UserID:      user.ID,
				TotalItem:   payload.TotalItem,
				InvoiceDate: *invoiceDate,
			})
			if delErr != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete po invoice: %v", delErr), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase order invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err), nil)
			return
		}
	}

	poiPdf := types.PurchaseOrderPdfPayload{
		Number:        payload.Number,
		InvoiceDate:   *invoiceDate,
		UserName:      user.Name,
		MedicineLists: payload.MedicineLists,
		Supplier:      *supplier,
	}
	fileName, err := pdf.CreatePurchaseOrderInvoicePdf(h.poInvoiceStore, poiPdf, "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("saved in database but failed to create pdf: %v", err), nil)
		return
	}

	err = h.poInvoiceStore.UpdatePdfUrl(purchaseOrderId, fileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update pdf in database: %v", err), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("purchase order invoice %d successfully created by %s", payload.Number, user.Name), nil)
}

// beginning of po invoice page, will request here
func (h *Handler) handleGetPOnvoiceNumberForToday(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	numberOfInvoices, err := h.poInvoiceStore.GetNumberOfPurchaseOrders()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, (numberOfInvoices + 1), nil)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseOrderPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("error parse date: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("error parse date: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	var purchaseOrders []types.PurchaseOrderListsReturnPayload

	if val == "all" {
		purchaseOrders, err = h.poInvoiceStore.GetPurchaseOrdersByDate(*startDate, *endDate)
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

		purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order id %d not exist", id), nil)
			return
		}

		supplier, err := h.supplierStore.GetSupplierByID(purchaseOrder.SupplierID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("supplier id %d not found", purchaseOrder.SupplierID), nil)
			return
		}

		user, err := h.userStore.GetUserByID(purchaseOrder.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user id %d not found", purchaseOrder.UserID), nil)
			return
		}

		purchaseOrders = append(purchaseOrders, types.PurchaseOrderListsReturnPayload{
			ID:           purchaseOrder.ID,
			Number:       purchaseOrder.Number,
			SupplierName: supplier.Name,
			UserName:     user.Name,
			TotalItem:    purchaseOrder.TotalItem,
			InvoiceDate:  purchaseOrder.InvoiceDate,
		})
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		purchaseOrders, err = h.poInvoiceStore.GetPurchaseOrdersByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else if params == "user" {
		users, err := h.userStore.GetUserBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val), nil)
			return
		}

		for _, user := range users {
			temp, err := h.poInvoiceStore.GetPurchaseOrdersByDateAndUserID(*startDate, *endDate, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest,
					fmt.Errorf("user %s doesn't create any po invoice between %s and %s", val, payload.StartDate, payload.EndDate), nil)
				return
			}

			purchaseOrders = append(purchaseOrders, temp...)
		}
	} else if params == "supplier" {
		suppliers, err := h.supplierStore.GetSupplierBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier %s not exists", val), nil)
			return
		}

		for _, supplier := range suppliers {
			temp, err := h.poInvoiceStore.GetPurchaseOrdersByDateAndSupplierID(*startDate, *endDate, supplier.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest,
					fmt.Errorf("supplier %s doesn't create any po invoice between %s and %s", val, payload.StartDate, payload.EndDate), nil)
				return
			}

			purchaseOrders = append(purchaseOrders, temp...)
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, purchaseOrders, nil)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseOrderDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseOrderDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get purchase order invoice data
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order invoice id %d doesn't exists", payload.ID), nil)
		return
	}

	// get medicine item of the purchase invoice
	purchaseOrderItem, err := h.poInvoiceStore.GetPurchaseOrderItem(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get supplier data
	supplier, err := h.supplierStore.GetSupplierByID(purchaseOrder.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d doesn't exists", purchaseOrder.SupplierID), nil)
		return
	}

	// get user data, the one who inputs the purchase invoice
	inputter, err := h.userStore.GetUserByID(purchaseOrder.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseOrder.UserID), nil)
		return
	}

	// get last modified user
	lastModifiedUser, err := h.userStore.GetUserByID(purchaseOrder.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseOrder.LastModifiedByUserID), nil)
		return
	}

	returnPayload := types.PurchaseOrderDetailPayload{
		ID:                     purchaseOrder.ID,
		Number:                 purchaseOrder.Number,
		TotalItem:              purchaseOrder.TotalItem,
		InvoiceDate:            purchaseOrder.InvoiceDate,
		CreatedAt:              purchaseOrder.CreatedAt,
		LastModified:           purchaseOrder.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,

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

		MedicineLists: purchaseOrderItem,
	}

	utils.WriteSuccess(w, http.StatusOK, returnPayload, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePurchaseOrder

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err), nil)
		return
	}

	// check if the purchase invoice exists
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByID(payload.ID)
	if purchaseOrder == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice id %d doesn't exist", payload.ID), nil)
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrderItem(purchaseOrder, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrder(purchaseOrder, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("purchase order invoice number %d deleted by %s", purchaseOrder.Number, user.Name), nil)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPurchaseOrderPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check supplier
	supplier, err := h.supplierStore.GetSupplierByID(payload.NewData.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d not found: %v", payload.NewData.SupplierID, err), nil)
		return
	}

	// check if the purchase order invoice exists
	purchaseOrder, err := h.poInvoiceStore.GetPurchaseOrderByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase order invoice with id %d doesn't exists", payload.ID), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.InvoiceDate)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("error parse date: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	err = h.poInvoiceStore.ModifyPurchaseOrder(payload.ID, types.PurchaseOrder{
		Number:               payload.NewData.Number,
		SupplierID:           payload.NewData.SupplierID,
		TotalItem:            payload.NewData.TotalItem,
		InvoiceDate:          *invoiceDate,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrderItem(purchaseOrder, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.poInvoiceStore.CreatePurchaseOrderItem(types.PurchaseOrderItem{
			PurchaseOrderID: payload.ID,
			MedicineID:      medData.ID,
			OrderQty:        medicine.OrderQty,
			ReceivedQty:     medicine.ReceivedQty,
			UnitID:          unit.ID,
			Remarks:         medicine.Remarks,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase order invoice %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err), nil)
			return
		}
	}

	poiPdf := types.PurchaseOrderPdfPayload{
		Number:        payload.NewData.Number,
		InvoiceDate:   *invoiceDate,
		UserName:      user.Name,
		MedicineLists: payload.NewData.MedicineLists,
		Supplier:      *supplier,
	}
	fileName, err := pdf.CreatePurchaseOrderInvoicePdf(h.poInvoiceStore, poiPdf, purchaseOrder.PdfURL)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("saved in database but failed to create pdf: %v", err), nil)
		return
	}

	err = h.poInvoiceStore.UpdatePdfUrl(purchaseOrder.ID, fileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update pdf in database: %v", err), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("purchase order invoice modified by %s", user.Name), nil)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseOrderDetailPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check if the purchase order exists
	purchaseInvoice, err := h.poInvoiceStore.GetPurchaseOrderByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase order with id %d doesn't exists", payload.ID), nil)
		return
	}

	pdfFile := "static/pdf/purchase-order/" + purchaseInvoice.PdfURL

	file, err := os.Open(pdfFile)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order id %d file not found", payload.ID), nil)
		return
	}
	defer file.Close()

	attachment := fmt.Sprintf("attachment; filename=%s", pdfFile)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, pdfFile)
}
