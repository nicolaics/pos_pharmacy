package cashier

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/service/auth"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	store types.CashierStore
}

func NewHandler(store types.CashierStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/cashier/login", h.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/cashier/login", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/cashier/register", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/cashier", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/delete", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/cashier/delete", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/modify", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/cashier/modify", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/logout", h.handleLogout).Methods(http.MethodPost)
	router.HandleFunc("/cashier/logout", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/init-admin", h.handleInitAdmin).Methods(http.MethodPost)
	router.HandleFunc("/cashier/init-admin", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)

	router.HandleFunc("/cashier/refresh-token", h.handleRefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/cashier/refresh-token", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.LoginCashierPayload

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

	cashier, err := h.store.GetCashierByName(payload.Name)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("not found, invalid name: %v", err))
		return
	}

	// check password match
	if !(auth.ComparePassword(cashier.Password, []byte(payload.Password))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("not found, invalid password"))
		return
	}

	tokenDetails, err := auth.CreateJWT(cashier.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.SaveToken(cashier.ID, tokenDetails)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.UpdateLastLoggedIn(cashier.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	tokens := map[string]string{
		"access_token":  tokenDetails.AccessToken,
		"refresh_token": tokenDetails.RefreshToken,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterCashierPayload

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
	admin, err := h.store.ValidateCashierToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid admin token or not admin"))
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("admin password wrong"))
		return
	}

	// check if the newly created cashier exists
	_, err = h.store.GetCashierByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("cashier with name %s already exists", payload.Name))
		return
	}

	// if it doesn't, we create new cashier
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateCashier(types.Cashier{
		Name:        payload.Name,
		Password:    hashedPassword,
		PhoneNumber: payload.PhoneNumber,
		Admin:       payload.Admin,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("cashier %s successfully created", payload.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.store.ValidateCashierToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid admin token or not admin"))
		return
	}

	cashiers, err := h.store.GetAllCashiers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, cashiers)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var payload types.RemoveCashierPayload

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
	admin, err := h.store.ValidateCashierToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid admin token or not admin"))
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("admin password wrong"))
		return
	}

	cashier, err := h.store.GetCashierByID(payload.ID)
	if cashier == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.store.DeleteCashier(cashier)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s successfully deleted", payload.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	var payload types.ModifyCashierPayload

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
	admin, err := h.store.ValidateCashierToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid admin token or not admin"))
		return
	}

	// validate admin password
	if !(auth.ComparePassword(admin.Password, []byte(payload.AdminPassword))) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("admin password wrong"))
		return
	}

	cashier, err := h.store.GetCashierByID(payload.ID)
	if cashier == nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if cashier.Name != payload.NewName {
		_, err = h.store.GetCashierByName(payload.NewName)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier with name %s already exists", payload.NewName))
		}
	}

	err = h.store.ModifyCashier(cashier.ID, types.Cashier{
		Name:        payload.NewName,
		Password:    payload.NewPassword,
		Admin:       payload.NewAdmin,
		PhoneNumber: payload.NewPhoneNumber,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("%s updated into", payload.NewName))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	accessDetails, err := auth.ExtractTokenFromRedis(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	_, err = h.store.DeleteToken(accessDetails.AccessUUID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	err = h.store.UpdateLastLoggedIn(accessDetails.CashierID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "successfully logged out")
}

func (h *Handler) handleInitAdmin(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.InitAdminPayload

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

	cashiers, err := h.store.GetAllCashiers()
	if err != nil || len(cashiers) != 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("initial admin has exist: %v", err))
		return
	}

	// create new admin
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateCashier(types.Cashier{
		Name:        payload.Name,
		Password:    hashedPassword,
		PhoneNumber: "000",
		Admin:       true,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("cashier %s successfully created", payload.Name))
}

func (h *Handler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshUUID, cashierId, err := auth.ValidateRefreshToken(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
	}

	_, err = h.store.DeleteToken(refreshUUID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	tokenDetails, err := auth.CreateJWT(cashierId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.SaveToken(cashierId, tokenDetails)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	tokens := map[string]string{
		"access_token":  tokenDetails.AccessToken,
		"refresh_token": tokenDetails.RefreshToken,
	}

	utils.WriteJSON(w, http.StatusOK, tokens)
}
