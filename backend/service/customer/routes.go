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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("register customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]string{"customer": payload.Name}
		logger.WriteServerErrorLog("register customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check if the customer exists
	_, err = h.custStore.GetCustomerByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s already exists", payload.Name))
		return
	}

	err = h.custStore.CreateCustomer(types.Customer{
		Name: payload.Name,
	})
	if err != nil {
		logData := map[string]string{"customer": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("register customer", user.ID, logData, fmt.Errorf("error create customer: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("customer %s successfully created by %s", payload.Name, user.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logger.WriteServerErrorLog("get all customers", 0, nil, fmt.Errorf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
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
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
			return
		}
	} else {
		id, err := strconv.Atoi(val)
		if err != nil {
			customers, err = h.custStore.GetCustomersBySearchName(val)
			if err != nil {
				logData := map[string]string{"customer": val}
				logFile, _ := logger.WriteServerErrorLog("get all customers", user.ID,
					logData, fmt.Errorf("get customer by search name error: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			if customers == nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s doesn't exist", val))
				return
			}
		} else {
			customer, err := h.custStore.GetCustomerByID(id)
			if err != nil {
				logData := map[string]int{"customer id": id}
				logFile, _ := logger.WriteServerErrorLog("get all customers", user.ID,
					logData, fmt.Errorf("get customer by search id error: %v", err))
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
				return
			}

			if customer == nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %s doesn't exist", id))
				return
			}

			customers = append(customers, *customer)
		}
	}

	utils.WriteSuccess(w, http.StatusOK, customers)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOneCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get one customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("get one customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logger.WriteServerErrorLog("get one customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get customer data
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get one customer", user.ID, logData, fmt.Errorf("error get customer by id: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("internal server error\n(%s)", logFile))
	}

	if customer == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer id %d doesn't exist", payload.ID))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, customer)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("delete customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logData := map[string]int{"customer id": payload.ID}
		logger.WriteServerErrorLog("delete customer", 0, logData, fmt.Errorf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check if the customer exists
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if customer == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer %s doesn't exist", payload.Name))
		return
	}

	err = h.custStore.DeleteCustomer(user, customer)
	if err != nil {
		data := map[string]string{"customer": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete customer", user.ID, data, fmt.Errorf("error delete customer: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("customer %s deleted by %s", payload.Name, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyCustomerPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, nil, fmt.Errorf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]int{"customer id": payload.ID}
		logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// check if the customer exists
	customer, err := h.custStore.GetCustomerByID(payload.ID)
	if err != nil {
		data := map[string]int{"customer id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("error get customer id: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
	}
	if customer == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with id %d doesn't exists", payload.ID))
		return
	}

	_, err = h.custStore.GetCustomerByName(payload.NewData.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("customer with name %s already exist", payload.NewData.Name))
		return
	}

	err = h.custStore.ModifyCustomer(customer.ID, payload.NewData.Name, user)
	if err != nil {
		data := map[string]interface{}{"customer id": payload.ID, "customer": payload.NewData.Name}
		logFile, _ := logger.WriteServerErrorLog("modify customer", 0, data, fmt.Errorf("error modify customer: %v", err))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error\n(%s)", logFile))
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("customer modified into %s by %s",
		payload.NewData.Name, user.Name))
}
