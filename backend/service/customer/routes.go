package customer

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
)

type Handler struct {
	custStore types.CustomerStore
	userStore types.UserStore
}

func NewHandler(custStore types.CustomerStore, userStore types.UserStore) *Handler {
	return &Handler{custStore: custStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/customer", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/customer/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/customer/detail", h.handleGetOne).Methods(http.MethodPost)
	router.HandleFunc("/customer", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/customer", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/customer", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/customer/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/customer/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("register customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Parsing payload failed",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("register customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Invalid payload",
			Log: logFile,
			Error: errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]string{"customer": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code: http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the customer exists
	_, err = h.custStore.GetCustomerByName(payload.Name)
	if err == nil {
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Customer with name %s already exists", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	err = h.custStore.CreateCustomer(types.Customer{
		Name: payload.Name,
	})
	if err != nil {
		logData := map[string]string{"customer": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register customer", user.ID, logData, fmt.Errorf("error create customer: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code: http.StatusCreated,
		Message: fmt.Sprintf("customer %s successfully created by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get all customers", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code: http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	vars := mux.Vars(r)
	val := vars["val"]

	log.Println("customer val: ", val)

	var customers []types.Customer

	if val == "all" {
		customers, err = h.custStore.GetAllCustomers()
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get all customers", user.ID, nil, fmt.Errorf("get all customers error: %v", err))
			resp := utils.Response{
				Code: http.StatusInternalServerError,
				Message: "Internal server error",
				Log: logFile,
				Error: err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else {
		id, err := strconv.Atoi(val)
		if err != nil {
			customers, err = h.custStore.GetCustomersBySearchName(val)
			if err != nil {
				logData := map[string]string{"customer": val}
				logFile, _ := logger.WriteServerErrorLog("get all customers", user.ID, logData, fmt.Errorf("get customer by search name error: %v", err))
				resp := utils.Response{
					Code: http.StatusInternalServerError,
					Message: "Internal server error",
					Log: logFile,
					Error: err.Error(),
				}
				resp.WriteError(w)
				return
			}

			if customers == nil {
				resp := utils.Response{
					Code: http.StatusBadRequest,
					Message: fmt.Sprintf("Customer %s doesn't exist", val),
				}
				resp.WriteError(w)
				return
			}
		} else {
			customer, err := h.custStore.GetCustomerByID(id)
			if err != nil {
				logData := map[string]int{"customer id": id}
				logFile, _ := logger.WriteServerErrorLog("get all customers", user.ID, logData, fmt.Errorf("get customer by search id error: %v", err))
				resp := utils.Response{
					Code: http.StatusInternalServerError,
					Message: "Internal server error",
					Log: logFile,
					Error: err.Error(),
				}
				resp.WriteError(w)
				return
			}

			if customer == nil {
				resp := utils.Response{
					Code: http.StatusBadRequest,
					Message: fmt.Sprintf("Customer ID %d doesn't exist", id),
				}
				resp.WriteError(w)
				return
			}

			customers = append(customers, *customer)
		}
	}

	resp := utils.Response{
		Code: http.StatusOK,
		Result: customers,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOneCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get one customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("get one customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Invalid payload",
			Log: logFile,
			Error: errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get one customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code: http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get customer data
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get one customer", user.ID, logData, fmt.Errorf("error get customer by id: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
	}

	if customer == nil {
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Customer ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code: http.StatusOK,
		Result: customer,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Invalid payload",
			Log: logFile,
			Error: errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code: http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the customer exists
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, logData, fmt.Errorf("error get customer by id: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if customer == nil {
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Customer %s doesn't exist", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	err = h.custStore.DeleteCustomer(user, customer)
	if err != nil {
		data := map[string]string{"customer": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete customer", user.ID, data, fmt.Errorf("error delete customer: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code: http.StatusOK,
		Message: fmt.Sprintf("Customer %s deleted by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Failed parsing payload",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: "Invalid payload",
			Log: logFile,
			Error: errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code: http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the customer exists
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if err != nil {
		data := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("error get customer id: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
	}
	if customer == nil {
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Customer ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	temp, err := h.custStore.GetCustomerByName(payload.NewData.Name)
	if err == nil || temp != nil {
		resp := utils.Response{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Customer %s already exist", payload.NewData.Name),
		}
		resp.WriteError(w)
		return
	}

	err = h.custStore.ModifyCustomer(customer.ID, payload.NewData.Name, user)
	if err != nil {
		data := map[string]interface{}{"customer id": payload.ID, "customer": payload.NewData.Name}
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("error modify customer: %v", err))
		resp := utils.Response{
			Code: http.StatusInternalServerError,
			Message: "Internal server error",
			Log: logFile,
			Error: err.Error(),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code: http.StatusOK,
		Message: fmt.Sprintf("Customer %s modified into %s by %s", customer.Name, payload.NewData.Name, user.Name),
	}
	resp.WriteSuccess(w)
}
