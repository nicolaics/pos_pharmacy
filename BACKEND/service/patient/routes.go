package patient

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
	patientStore types.PatientStore
	userStore types.UserStore
}

func NewHandler(patientStore types.PatientStore, userStore types.UserStore) *Handler {
	return &Handler{patientStore: patientStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/patient", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/patient?{params}={val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/patient/detail", h.handleGetOne).Methods(http.MethodPost)
	router.HandleFunc("/patient", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/patient", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/patient", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/patient?{params}={val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/patient/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterPatientPayload

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

	// check if the patient exists
	_, err = h.patientStore.GetPatientByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("patient with name %s already exists", payload.Name))
		return
	}

	err = h.patientStore.CreatePatient(types.Patient{
		Name: payload.Name,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("patient %s successfully created by %s", payload.Name, user.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	var patients []types.Patient

	if params == "all" && val == "all" {
		patients, err = h.patientStore.GetAllPatients()
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}
	} else if params == "name" {
		patient, err := h.patientStore.GetPatientByName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient %s not found", val))
			return
		}

		patients = append(patients, *patient)
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		patient, err := h.patientStore.GetPatientByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient id %d not found", id))
			return
		}

		patients = append(patients, *patient)
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown query"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, patients)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOnePatientPayload

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

	// get patient data
	patient, err := h.patientStore.GetPatientByID(payload.ID)
	if patient == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("patient id %d doesn't exist", payload.ID))
		return
	}

	utils.WriteJSON(w, http.StatusOK, patient)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePatientPayload

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

	// check if the patient exists
	patient, err := h.patientStore.GetPatientByID(payload.ID)
	if patient == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("patient %s doesn't exist", payload.Name))
		return
	}

	err = h.patientStore.DeletePatient(patient, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("patient %s deleted by %s", payload.Name, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPatientPayload

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

	// check if the patient exists
	patient, err := h.patientStore.GetPatientByID(payload.ID)
	if err != nil || patient == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("patient with id %d doesn't exists", payload.ID))
		return
	}

	_, err = h.patientStore.GetPatientByName(payload.NewData.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("patient with name %s already exist", payload.NewData.Name))
		return
	}

	err = h.patientStore.ModifyPatient(patient.ID, payload.NewData.Name, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("patient modified into %s by %s",
		payload.NewData.Name, user.Name))
}
