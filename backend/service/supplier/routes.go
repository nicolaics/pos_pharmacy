package supplier

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
)

type Handler struct {
	supplierStore types.SupplierStore
	userStore     types.UserStore
}

func NewHandler(supplierStore types.SupplierStore, userStore types.UserStore) *Handler {
	return &Handler{supplierStore: supplierStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/supplier", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/supplier/{params}/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/supplier/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/supplier", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/supplier", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/supplier", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/supplier/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/supplier/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterSupplierPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("register supplier", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("register supplier", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate user token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register supplier", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the supplier exists
	temp, err := h.supplierStore.GetSupplierByName(payload.Name)
	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register supplier", user.ID, data, fmt.Errorf("error get supplier by name: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if temp != nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Supplier %s already exists", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	err = h.supplierStore.CreateSupplier(types.Supplier{
		Name:                 payload.Name,
		Address:              payload.Address,
		CompanyPhoneNumber:   payload.CompanyPhoneNumber,
		ContactPersonName:    payload.ContactPersonName,
		ContactPersonNumber:  payload.ContactPersonNumber,
		Terms:                payload.Terms,
		VendorIsTaxable:      payload.VendorIsTaxable,
		LastModifiedByUserID: user.ID,
	})

	if err != nil {
		data := map[string]interface{}{"name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register supplier", user.ID, data, fmt.Errorf("error create supplier: %v", err))
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
		Message: fmt.Sprintf("Supplier %s created by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate user token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get all suppliers", 0, nil, fmt.Errorf("user token invalid: %v", err))
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

	var suppliers []types.SupplierInformationReturnPayload

	if val == "all" {
		suppliers, err = h.supplierStore.GetAllSuppliers()
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get all suppliers", user.ID, nil, fmt.Errorf("error get all suppliers: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "name" {
		suppliers, err = h.supplierStore.GetSupplierBySearchName(val)
		if err != nil {
			data := map[string]interface{}{"searched_supplier": val}
			logFile, _ := logger.WriteServerErrorLog("get all suppliers", user.ID, data, fmt.Errorf("error get supplier by search name: %v", err))
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
			data := map[string]interface{}{"searched_id": val}
			logFile, _ := logger.WriteServerErrorLog("get all suppliers", user.ID, data, fmt.Errorf("error parse id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		supplier, err := h.supplierStore.GetSupplierByID(id)
		if err != nil {
			data := map[string]interface{}{"searched_id": id}
			logFile, _ := logger.WriteServerErrorLog("get all suppliers", user.ID, data, fmt.Errorf("error get supplier by id: %v", err))
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
				Message: fmt.Sprintf("Supplier ID %d doesn't exist", id),
			}
			resp.WriteError(w)
			return
		}

		suppliers = append(suppliers, *supplier)
	} else if params == "cp-name" {
		suppliers, err = h.supplierStore.GetSupplierBySearchContactPersonName(val)
		if err != nil {
			data := map[string]interface{}{"searched_contact_person": val}
			logFile, _ := logger.WriteServerErrorLog("get all suppliers", user.ID, data, fmt.Errorf("error get supplier by search cp name: %v", err))
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
			Message: "Unknown query",
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: suppliers,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetDetailSupplierPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get detail supplier", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get detail supplier", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate user token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get detail supplier", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get supplier data
	supplier, err := h.supplierStore.GetSupplierByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get detail supplier", user.ID, data, fmt.Errorf("error get by id: %v", err))
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
			Message: fmt.Sprintf("Supplier ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: supplier,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteSupplierPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete supplier", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("delete supplier", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate user token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete supplier", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.supplierStore.DeleteSupplier(&types.Supplier{
		ID:   payload.ID,
		Name: payload.Name,
	}, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete supplier", 0, data, fmt.Errorf("error delete: %v", err))
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
		Message: fmt.Sprintf("Supplier %s deleted by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifySupplierPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify supplier", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("modify supplier", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate user token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_name": payload.NewData.Name}
		logFile, _ := logger.WriteServerErrorLog("modify supplier", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the supplier exists
	supplier, err := h.supplierStore.GetSupplierByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_name": payload.NewData.Name}
		logFile, _ := logger.WriteServerErrorLog("modify supplier", user.ID, data, fmt.Errorf("error get by id: %v", err))
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
			Message: fmt.Sprintf("Supplier ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	if supplier.Name != payload.NewData.Name {
		temp, err := h.supplierStore.GetSupplierByName(payload.NewData.Name)
		if err != nil {
			data := map[string]interface{}{"id": payload.ID, "new_name": payload.NewData.Name}
			logFile, _ := logger.WriteServerErrorLog("modify supplier", user.ID, data, fmt.Errorf("error get by name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
		if temp != nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Supplier %s already exists", payload.NewData.Name),
			}
			resp.WriteError(w)
			return
		}
	}

	err = h.supplierStore.ModifySupplier(supplier.ID, types.Supplier{
		Name:                 payload.NewData.Name,
		Address:              payload.NewData.Address,
		CompanyPhoneNumber:   payload.NewData.CompanyPhoneNumber,
		ContactPersonName:    payload.NewData.ContactPersonName,
		ContactPersonNumber:  payload.NewData.ContactPersonNumber,
		Terms:                payload.NewData.Terms,
		VendorIsTaxable:      payload.NewData.VendorIsTaxable,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		data := map[string]interface{}{"id": payload.ID, "new_name": payload.NewData.Name}
		logFile, _ := logger.WriteServerErrorLog("modify supplier", user.ID, data, fmt.Errorf("error modify: %v", err))
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
		Code:    http.StatusInternalServerError,
		Message: fmt.Sprintf("Supplier ID %d modified by %s", payload.ID, user.Name),
	}
	resp.WriteSuccess(w)
}
