package companyprofile

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	companyProfileStore types.CompanyProfileStore
	cashierStore        types.CashierStore
}

func NewHandler(companyProfileStore types.CompanyProfileStore, cashierStore types.CashierStore) *Handler {
	return &Handler{companyProfileStore: companyProfileStore, cashierStore: cashierStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/company-profile", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/company-profile", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/company-profile", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/company-profile", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/company-profile", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterCompanyProfilePayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token or not admin: %v", err))
		return
	}

	// check if the supplier exists
	_, err = h.companyProfileStore.GetCompanyProfileByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("company profile with name %s already exists", payload.Name))
		return
	}

	err = h.companyProfileStore.CreateCompanyProfile(types.CompanyProfile{
		Name:                    payload.Name,
		Address:                 payload.Address,
		BusinessNumber:          payload.BusinessNumber,
		Pharmacist:              payload.Pharmacist,
		PharmacistLicenseNumber: payload.PharmacistLicenseNumber,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("company profile '%s' created by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate cashier token
	_, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token or not admin: %v", err))
		return
	}

	// check if the supplier exists
	companyProfiles, err := h.companyProfileStore.GetAllCompanyProfiles()
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, companyProfiles)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteCompanyProfilePayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token or not admin: %v", err))
		return
	}

	// check if the supplier exists
	companyProfile, err := h.companyProfileStore.GetCompanyProfileByID(payload.ID)
	if err != nil || companyProfile == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("company profile with name %s doesn't exists", payload.Name))
		return
	}

	err = h.companyProfileStore.DeleteCompanyProfile(companyProfile)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("company profile %s deleted by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyCompanyProfilePayload

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

	// validate cashier token
	cashier, err := h.cashierStore.ValidateCashierAccessToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token or not admin: %v", err))
		return
	}

	// check if the company profile exists
	companyProfile, err := h.companyProfileStore.GetCompanyProfileByName(payload.NewName)
	if err != nil || companyProfile == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("company profile with name %s doesn't exists", payload.NewName))
		return
	}

	err = h.companyProfileStore.ModifyCompanyProfile(companyProfile.ID, types.CompanyProfile{
		Name:                    payload.NewName,
		Address:                 payload.NewAddress,
		BusinessNumber:          payload.NewAddress,
		Pharmacist:              payload.NewPharmacist,
		PharmacistLicenseNumber: payload.NewPharmacistLicenseNumber,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("company profile %d modified by %s", payload.ID, cashier.Name))
}
