package purchaseinvoice

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	purchaseInvoiceStore types.PurchaseInvoiceStore
	userStore            types.UserStore
	supplierStore        types.SupplierStore
	companyProfileStore  types.CompanyProfileStore
	medStore             types.MedicineStore
	unitStore            types.UnitStore
}

func NewHandler(purchaseInvoiceStore types.PurchaseInvoiceStore, userStore types.UserStore,
	supplierStore types.SupplierStore, companyProfileStore types.CompanyProfileStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		purchaseInvoiceStore: purchaseInvoiceStore,
		userStore:            userStore,
		supplierStore:        supplierStore,
		companyProfileStore:  companyProfileStore,
		medStore:             medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice/purchase", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase/{params}/{val}", h.handleGetPurchaseInvoices).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase/detail", h.handleGetPurchaseInvoiceDetail).Methods(http.MethodPost)
	router.HandleFunc("/invoice/purchase", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/invoice/purchase", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/invoice/purchase", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/invoice/purchase/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.PurchaseInvoicePayload

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

	err = h.purchaseInvoiceStore.CreatePurchaseInvoice(types.PurchaseInvoice{
		Number:               payload.Number,
		CompanyID:            payload.CompanyID,
		SupplierID:           payload.SupplierID,
		Subtotal:             payload.Subtotal,
		Discount:             payload.Discount,
		Tax:                  payload.Tax,
		TotalPrice:           payload.TotalPrice,
		Description:          payload.Description,
		UserID:               user.ID,
		InvoiceDate:          payload.InvoiceDate,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get purchaseInvoiceID
	purchaseInvoiceId, err := h.purchaseInvoiceStore.GetPurchaseInvoiceID(payload.Number, payload.CompanyID, payload.SupplierID, payload.Subtotal, payload.TotalPrice, user.ID, payload.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice number %d doesn't exists", payload.Number))
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

		err = h.purchaseInvoiceStore.CreatePurchaseMedicineItems(types.PurchaseMedicineItem{
			PurchaseInvoiceID: purchaseInvoiceId,
			MedicineID:        medData.ID,
			Qty:               medicine.Qty,
			UnitID:            unit.ID,
			PurchasePrice:     medicine.Price,
			PurchaseDiscount:  medicine.Discount,
			PurchaseTax:       medicine.Tax,
			Subtotal:          medicine.Subtotal,
			BatchNumber:       medicine.BatchNumber,
			ExpDate:           medicine.ExpDate,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase invoice %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}
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

	var purchaseInvoices []types.PurchaseInvoiceListsReturnPayload

	if val == "all" {
		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDate(payload.StartDate, payload.EndDate)
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

		purchaseInvoices, err = h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndNumber(payload.StartDate, payload.EndDate, number)
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
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndUserID(payload.StartDate, payload.EndDate, user.ID)
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
			temp, err := h.purchaseInvoiceStore.GetPurchaseInvoicesByDateAndUserID(payload.StartDate, payload.EndDate, supplier.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("supplier %s doesn't create any purchase invoice between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			purchaseInvoices = append(purchaseInvoices, temp...)
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
	var payload types.ViewPurchaseMedicineItemsPayload

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

	// get medicine items of the purchase invoice
	purchaseMedicineItems, err := h.purchaseInvoiceStore.GetPurchaseMedicineItems(purchaseInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get company profile
	company, err := h.companyProfileStore.GetCompanyProfileByID(purchaseInvoice.CompanyID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("company id %d doesn't exists", purchaseInvoice.CompanyID))
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
		Discount:               purchaseInvoice.Discount,
		Tax:                    purchaseInvoice.Tax,
		TotalPrice:             purchaseInvoice.TotalPrice,
		Description:            purchaseInvoice.Description,
		InvoiceDate:            purchaseInvoice.InvoiceDate,
		CreatedAt:              purchaseInvoice.CreatedAt,
		LastModified:           purchaseInvoice.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,

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

		MedicineLists: purchaseMedicineItems,
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

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItems(purchaseInvoice, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseInvoice(purchaseInvoice, user.ID)
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

	err = h.purchaseInvoiceStore.ModifyPurchaseInvoice(payload.ID, types.PurchaseInvoice{
		Number:               payload.NewData.Number,
		CompanyID:            payload.NewData.CompanyID,
		SupplierID:           payload.NewData.SupplierID,
		Subtotal:             payload.NewData.Subtotal,
		Discount:             payload.NewData.Discount,
		Tax:                  payload.NewData.Tax,
		TotalPrice:           payload.NewData.TotalPrice,
		Description:          payload.NewData.Description,
		InvoiceDate:          payload.NewData.InvoiceDate,
		LastModifiedByUserID: user.ID,
	}, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItems(purchaseInvoice, user.ID)
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

		err = h.purchaseInvoiceStore.CreatePurchaseMedicineItems(types.PurchaseMedicineItem{
			PurchaseInvoiceID: payload.ID,
			MedicineID:        medData.ID,
			Qty:               medicine.Qty,
			UnitID:            unit.ID,
			PurchasePrice:     medicine.Price,
			PurchaseDiscount:  medicine.Discount,
			PurchaseTax:       medicine.Tax,
			Subtotal:          medicine.Subtotal,
			BatchNumber:       medicine.BatchNumber,
			ExpDate:           medicine.ExpDate,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("purchase invoice %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase invoice modified by %s", user.Name))
}
