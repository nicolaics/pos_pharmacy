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
	cashierStore       types.CashierStore
	custStore          types.CustomerStore
	paymentMethodStore types.PaymentMethodStore
	medStore           types.MedicineStore
	unitStore          types.UnitStore
}

func NewHandler(invoiceStore types.InvoiceStore, cashierStore types.CashierStore,
	custStore types.CustomerStore, paymentMethodStore types.PaymentMethodStore,
	medStore types.MedicineStore, unitStore types.UnitStore) *Handler {
	return &Handler{
		invoiceStore:       invoiceStore,
		cashierStore:       cashierStore,
		custStore:          custStore,
		paymentMethodStore: paymentMethodStore,
		medStore:           medStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/invoice", h.handleNew).Methods(http.MethodPost)
	router.HandleFunc("/invoice", h.handleGetInvoices).Methods(http.MethodGet)
	// router.HandleFunc("/invoice/medicine-items", h.handleGetMedicineItems).Methods(http.MethodGet)
	// router.HandleFunc("/invoice/purchase", h.handleDelete).Methods(http.MethodDelete)
	// router.HandleFunc("/invoice/purchase", h.handleModify).Methods(http.MethodPatch)

	// router.HandleFunc("/invoice/purchase", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	// router.HandleFunc("/invoice/purchase/medicine-items", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleNew(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.NewInvoicePayload

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
		Number:          payload.Number,
		CashierID:       cashier.ID,
		CustomerID:      payload.CustomerID,
		Subtotal:        payload.Subtotal,
		Discount:        payload.Discount,
		Tax:             payload.Tax,
		TotalPrice:      payload.TotalPrice,
		PaidAmount:      payload.PaidAmount,
		ChangeAmount:    payload.ChangeAmount,
		PaymentMethodID: paymentMethod.ID,
		Description:     payload.Description,
		InvoiceDate:     payload.InvoiceDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	// get invoice id
	invoice, err := h.invoiceStore.GetInvoiceByAll(payload.Number, cashier.ID, payload.CustomerID, payload.TotalPrice, payload.InvoiceDate)
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
			InvoiceID:  invoice.ID,
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

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("invoice %d successfully created by %s", payload.Number, cashier.Name))
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
	_, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	invoices, err := h.invoiceStore.GetInvoicesByDate(payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, invoices)
}

/*
func (h *Handler) handleGetPurchaseMedicineItems(w http.ResponseWriter, r *http.Request) {
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
	_, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	// get purchase invoice data
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.PurchaseInvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("purchase invoice id %d doesn't exists", payload.PurchaseInvoiceID))
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

	// get cashier data, the one who inputs the purchase invoice
	inputter, err := h.cashierStore.GetCashierByID(purchaseInvoice.CashierID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier id %d doesn't exists", purchaseInvoice.CashierID))
		return
	}

	returnPayload := types.PurchaseInvoiceReturnJSONPayload{
		PurchaseInvoiceID:          purchaseInvoice.ID,
		PurchaseInvoiceNumber:      purchaseInvoice.Number,
		PurchaseInvoiceSubtotal:    purchaseInvoice.Subtotal,
		PurchaseInvoiceDiscount:    purchaseInvoice.Discount,
		PurchaseInvoiceTax:         purchaseInvoice.Tax,
		PurchaseInvoiceTotalPrice:  purchaseInvoice.TotalPrice,
		PurchaseInvoiceDescription: purchaseInvoice.Description,
		PurchaseInvoiceInvoiceDate: purchaseInvoice.InvoiceDate,

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

		MedicineLists: purchaseMedicineItems,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

/*
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
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid or not admin: %v", err))
		return
	}

	// check if the purchase invoice exists
	purchaseInvoice, err := h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.ID)
	if purchaseInvoice == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice id %d doesn't exist", payload.ID))
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItems(purchaseInvoice.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseInvoice(purchaseInvoice)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("purchase invoice number %d deleted by %s", purchaseInvoice.Number, cashier.Name))
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
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid: %v", err))
		return
	}

	// check if the purchase invoice exists
	_, err = h.purchaseInvoiceStore.GetPurchaseInvoiceByID(payload.PurchaseInvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("purchase invoice with id %d doesn't exists", payload.PurchaseInvoiceID))
		return
	}

	err = h.purchaseInvoiceStore.ModifyPurchaseInvoice(payload.PurchaseInvoiceID, types.PurchaseInvoice{
		Number:      payload.NewNumber,
		CompanyID:   payload.NewCompanyID,
		SupplierID:  payload.NewSupplierID,
		Subtotal:    payload.NewSubtotal,
		Discount:    payload.NewDiscount,
		Tax:         payload.NewTax,
		TotalPrice:  payload.NewTotalPrice,
		Description: payload.NewDescription,
		CashierID:   cashier.ID,
		InvoiceDate: payload.NewInvoiceDate,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.purchaseInvoiceStore.DeletePurchaseMedicineItems(payload.PurchaseInvoiceID)
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

		err = h.purchaseInvoiceStore.CreatePurchaseMedicineItems(types.PurchaseMedicineItem{
			PurchaseInvoiceID: payload.PurchaseInvoiceID,
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
				fmt.Errorf("purchase invoice %d, med %s: %v", payload.NewNumber, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("purchase invoice modified by %s", cashier.Name))
}
*/
