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
	userStore           types.UserStore
	supplierStore       types.SupplierStore
	companyProfileStore types.CompanyProfileStore
	medStore            types.MedicineStore
	unitStore           types.UnitStore
}

func NewHandler(poInvoiceStore types.PurchaseOrderInvoiceStore, userStore types.UserStore,
	supplierStore types.SupplierStore, companyProfileStore types.CompanyProfileStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		poInvoiceStore:      poInvoiceStore,
		userStore:           userStore,
		supplierStore:       supplierStore,
		companyProfileStore: companyProfileStore,
		medStore:            medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice/purchase-order", h.handleNew).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase-order", h.handleGetPurchaseOrderInvoices).Methods(http.MethodGet)
	router.HandleFunc("/invoice/purchase-order/detail", h.handleGetOnePurchaseOrderInvoiceDetail).Methods(http.MethodGet)
	router.HandleFunc("/invoice/purchase-order", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice/purchase-order", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice/purchase-order", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase-order/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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
		Number:      payload.Number,
		CompanyID:   payload.CompanyID,
		SupplierID:  payload.SupplierID,
		UserID:      user.ID,
		TotalItems:  payload.TotalItems,
		InvoiceDate: payload.InvoiceDate,
		ModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get purchaseInvoice ID
	purchaseOrderInvoiceId, err := h.poInvoiceStore.GetPurchaseOrderInvoiceID(payload.Number, payload.CompanyID, payload.SupplierID, user.ID, payload.TotalItems, payload.InvoiceDate)
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
			PurchaseOrderInvoiceID: purchaseOrderInvoiceId,
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

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase order invoice %d successfully created by %s", payload.Number, user.Name))
}

// only view the purchase invoice list
func (h *Handler) handleGetPurchaseOrderInvoices(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewOnePurchaseOrderInvoicePayload

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

	purchaseOrderInvoices, err := h.poInvoiceStore.GetPurchaseOrderInvoices(payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, purchaseOrderInvoices)
}

// only view the purchase invoice list
func (h *Handler) handleGetOnePurchaseOrderInvoiceDetail(w http.ResponseWriter, r *http.Request) {
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
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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

	// get user data, the one who inputs the purchase invoice
	inputter, err := h.userStore.GetUserByID(purchaseOrderInvoice.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseOrderInvoice.UserID))
		return
	}

	// get last modified user
	lastModifiedUser, err := h.userStore.GetUserByID(purchaseOrderInvoice.ModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", purchaseOrderInvoice.ModifiedByUserID))
		return
	}

	returnPayload := types.PurchaseOrderInvoiceDetailPayload{
		ID:                 purchaseOrderInvoice.ID,
		Number:             purchaseOrderInvoice.Number,
		TotalItems:         purchaseOrderInvoice.TotalItems,
		InvoiceDate:        purchaseOrderInvoice.InvoiceDate,
		CreatedAt:          purchaseOrderInvoice.CreatedAt,
		LastModified:       purchaseOrderInvoice.LastModified,
		ModifiedByUserName: lastModifiedUser.Name,

		CompanyProfile: struct {
			ID                      int    "json:\"id\""
			Name                    string "json:\"name\""
			Address                 string "json:\"address\""
			BusinessNumber          string "json:\"businessNumber\""
			Pharmacist              string "json:\"pharmacist\""
			PharmacistLicenseNumber string "json:\"pharmacistLicenseNumber\""
		}{
			ID:                      company.ID,
			Name:                    company.Name,
			Address:                 company.Address,
			BusinessNumber:          company.BusinessNumber,
			Pharmacist:              company.Pharmacist,
			PharmacistLicenseNumber: company.PharmacistLicenseNumber,
		},

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
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err))
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

	err = h.poInvoiceStore.DeletePurchaseOrderInvoice(purchaseOrderInvoice, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("purchase order invoice number %d deleted by %s", purchaseOrderInvoice.Number, user.Name))
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
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
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
		Number:           payload.NewNumber,
		CompanyID:        payload.NewCompanyID,
		SupplierID:       payload.NewSupplierID,
		TotalItems:       payload.NewTotalItems,
		InvoiceDate:      payload.NewInvoiceDate,
		ModifiedByUserID: user.ID,
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

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase order invoice modified by %s", user.Name))
}
