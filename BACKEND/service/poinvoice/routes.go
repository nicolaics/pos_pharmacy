package poinvoice

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	poInvoiceStore      types.PurchaseOrderInvoiceStore
	cashierStore        types.CashierStore
	supplierStore       types.SupplierStore
	companyProfileStore types.CompanyProfileStore
	medStore            types.MedicineStore
	unitStore           types.UnitStore
}

func NewHandler(poInvoiceStore types.PurchaseOrderInvoiceStore, cashierStore types.CashierStore,
	supplierStore types.SupplierStore, companyProfileStore types.CompanyProfileStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		poInvoiceStore:      poInvoiceStore,
		cashierStore:        cashierStore,
		supplierStore:       supplierStore,
		companyProfileStore: companyProfileStore,
		medStore:            medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice/purchase-order", h.handleNew).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase-order", h.handleGetPurchaseOrderInvoices).Methods(http.MethodGet)
	router.HandleFunc("/invoice/purchase-order/items", h.handleGetPurchaseOrderItems).Methods(http.MethodGet)
	router.HandleFunc("/invoice/purchase-order", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice/purchase-order", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice/purchase-order", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase-order/medicine-items", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleNew(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.NewPurchaseOrderInvoicePayload

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

	// check companyID
	_, err = h.companyProfileStore.GetCompanyProfileByID(payload.CompanyID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("company id %d not found", payload.CompanyID))
		return
	}

	// check supplierID
	_, err = h.supplierStore.GetSupplierByID(payload.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d not found", payload.SupplierID))
		return
	}

	err = h.poInvoiceStore.CreatePurchaseOrderInvoice(types.PurchaseOrderInvoice{
		Number:       payload.Number,
		CompanyID:    payload.CompanyID,
		SupplierID:   payload.SupplierID,
		CashierID:    cashier.ID,
		TotalItems:   payload.TotalItems,
		InvoiceDate:  payload.InvoiceDate,
		LastModified: payload.LastModified,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get purchaseInvoice number
	purchaseOrderInvoice, err := h.poInvoiceStore.GetPurchaseOrderInvoiceByAll(payload.Number, payload.CompanyID, payload.SupplierID, cashier.ID, payload.TotalItems, payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order invoice number %d doesn't exists", payload.Number))
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

		err = h.poInvoiceStore.CreatePurchaseOrderItems(types.PurchaseOrderItem{
			PurchaseOrderInvoiceID: purchaseOrderInvoice.ID,
			MedicineID:             medData.ID,
			OrderQty:               medicine.OrderQty,
			ReceivedQty:            medicine.ReceivedQty,
			UnitID:                 unit.ID,
			Remarks:                medicine.Remarks,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase order invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase order invoice %d successfully created by %s", payload.Number, cashier.Name))
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseOrderInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseOrderInvoicePayload

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

	purchaseOrderInvoices, err := h.poInvoiceStore.GetPurchaseOrderInvoices(payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, purchaseOrderInvoices)
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseOrderItems(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPurchaseOrderItemsPayload

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

	// get purchase order invoice data
	purchaseOrderInvoice, err := h.poInvoiceStore.GetPurchaseOrderInvoiceByID(payload.PurchaseOrderInvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase order invoice id %d doesn't exists", payload.PurchaseOrderInvoiceID))
		return
	}

	// get medicine items of the purchase invoice
	purchaseOrderItems, err := h.poInvoiceStore.GetPurchaseOrderItems(payload.PurchaseOrderInvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get company profile
	company, err := h.companyProfileStore.GetCompanyProfileByID(purchaseOrderInvoice.CompanyID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("company id %d doesn't exists", purchaseOrderInvoice.CompanyID))
		return
	}

	// get supplier data
	supplier, err := h.supplierStore.GetSupplierByID(purchaseOrderInvoice.SupplierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier id %d doesn't exists", purchaseOrderInvoice.SupplierID))
		return
	}

	// get cashier data, the one who inputs the purchase invoice
	inputter, err := h.cashierStore.GetCashierByID(purchaseOrderInvoice.CashierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier id %d doesn't exists", purchaseOrderInvoice.CashierID))
		return
	}

	returnPayload := types.PurchaseOrderInvoiceReturnJSONPayload{
		PurchaseOrderInvoiceID:           purchaseOrderInvoice.ID,
		PurchaseOrderInvoiceNumber:       purchaseOrderInvoice.Number,
		PurchaseOrderInvoiceTotalItems:   purchaseOrderInvoice.TotalItems,
		PurchaseOrderInvoiceInvoiceDate:  purchaseOrderInvoice.InvoiceDate,
		PurchaseOrderInvoiceLastModified: purchaseOrderInvoice.LastModified,

		CompanyID:               company.ID,
		CompanyName:             company.Name,
		CompanyAddress:          company.Address,
		CompanyBusinessNumber:   company.BusinessNumber,
		Pharmacist:              company.Pharmacist,
		PharmacistLicenseNumber: company.PharmacistLicenseNumber,

		SupplierID:                  supplier.ID,
		SupplierName:                supplier.Name,
		SupplierAddress:             supplier.Address,
		SupplierPhoneNumber:         supplier.CompanyPhoneNumber,
		SupplierContactPersonName:   supplier.ContactPersonName,
		SupplierContactPersonNumber: supplier.ContactPersonNumber,
		SupplierTerms:               supplier.Terms,
		SupplierVendorIsTaxable:     supplier.VendorIsTaxable,

		CashierID:   inputter.ID,
		CashierName: inputter.Name,

		MedicineLists: purchaseOrderItems,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePurchaseOrderInvoice

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
	purchaseOrderInvoice, err := h.poInvoiceStore.GetPurchaseOrderInvoiceByID(payload.ID)
	if purchaseOrderInvoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice id %d doesn't exist", payload.ID))
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrderItems(purchaseOrderInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrderInvoice(purchaseOrderInvoice)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("purchase order invoice number %d deleted by %s", purchaseOrderInvoice.Number, cashier.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPurchaseOrderInvoicePayload

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

	// check if the purchase order invoice exists
	_, err = h.poInvoiceStore.GetPurchaseOrderInvoiceByID(payload.PurchaseOrderInvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase order invoice with id %d doesn't exists", payload.PurchaseOrderInvoiceID))
		return
	}

	err = h.poInvoiceStore.ModifyPurchaseOrderInvoice(payload.PurchaseOrderInvoiceID, types.PurchaseOrderInvoice{
		Number:       payload.NewNumber,
		CompanyID:    payload.NewCompanyID,
		SupplierID:   payload.NewSupplierID,
		CashierID:    cashier.ID,
		TotalItems:   payload.NewTotalItems,
		InvoiceDate:  payload.NewInvoiceDate,
		LastModified: payload.NewLastModified,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.poInvoiceStore.DeletePurchaseOrderItems(payload.PurchaseOrderInvoiceID)
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

		err = h.poInvoiceStore.CreatePurchaseOrderItems(types.PurchaseOrderItem{
			PurchaseOrderInvoiceID: payload.PurchaseOrderInvoiceID,
			MedicineID:             medData.ID,
			OrderQty:               medicine.OrderQty,
			ReceivedQty:            medicine.ReceivedQty,
			UnitID:                 unit.ID,
			Remarks:                medicine.Remarks,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase order invoice %d, med %s: %v", payload.NewNumber, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase order invoice modified by %s", cashier.Name))
}
