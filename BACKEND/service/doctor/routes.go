package doctor

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
	doctorStore types.DoctorStore
	userStore   types.UserStore
}

func NewHandler(doctorStore types.DoctorStore, userStore types.UserStore) *Handler {
	return &Handler{doctorStore: doctorStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/doctor", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/doctor/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/doctor/detail", h.handleGetOne).Methods(http.MethodPost)
	router.HandleFunc("/doctor", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/doctor", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/doctor", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/doctor/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/doctor/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterDoctorPayload

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

	// check if the doctor exists
	_, err = h.doctorStore.GetDoctorByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("doctor with name %s already exists", payload.Name))
		return
	}

	err = h.doctorStore.CreateDoctor(types.Doctor{
		Name: payload.Name,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("doctor %s successfully created by %s", payload.Name, user.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	vars := mux.Vars(r)
	val := vars["val"]

	var doctors []types.Doctor

	if val == "all" {
		doctors, err = h.doctorStore.GetAllDoctors()
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}
	} else {
		id, err := strconv.Atoi(val)
		if err != nil {
			doctors, err = h.doctorStore.GetDoctorsBySimilarName(val)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor %s not found", val))
				return
			}
		} else {
			doctor, err := h.doctorStore.GetDoctorByID(id)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor id %d not found", id))
				return
			}

			doctors = append(doctors, *doctor)
		}
	}

	utils.WriteJSON(w, http.StatusOK, doctors)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOneDoctorPayload

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

	// get doctor data
	doctor, err := h.doctorStore.GetDoctorByID(payload.ID)
	if doctor == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("doctor id %d doesn't exist", payload.ID))
		return
	}

	utils.WriteJSON(w, http.StatusOK, doctor)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteDoctorPayload

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

	// check if the doctor exists
	doctor, err := h.doctorStore.GetDoctorByID(payload.ID)
	if doctor == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("doctor %s doesn't exist", payload.Name))
		return
	}

	err = h.doctorStore.DeleteDoctor(doctor, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("doctor %s deleted by %s", payload.Name, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyDoctorPayload

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

	// check if the doctor exists
	doctor, err := h.doctorStore.GetDoctorByID(payload.ID)
	if err != nil || doctor == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("doctor with id %d doesn't exists", payload.ID))
		return
	}

	_, err = h.doctorStore.GetDoctorByName(payload.NewData.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("doctor with name %s already exist", payload.NewData.Name))
		return
	}

	err = h.doctorStore.ModifyDoctor(doctor.ID, payload.NewData.Name, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("doctor modified into %s by %s",
		payload.NewData.Name, user.Name))
}
