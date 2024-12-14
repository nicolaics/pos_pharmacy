package medicine

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"github.com/nicolaics/pharmacon/utils/export"
)

type Handler struct {
	medStore  types.MedicineStore
	userStore types.UserStore
	unitStore types.UnitStore
}

func NewHandler(medStore types.MedicineStore, userStore types.UserStore, unitStore types.UnitStore) *Handler {
	return &Handler{medStore: medStore, userStore: userStore, unitStore: unitStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/medicine", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/medicine/{params}/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/medicine/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/medicine", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/medicine", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/medicine/history", h.handleGetHistory).Methods(http.MethodPost)
	router.HandleFunc("/medicine/export/csv", h.handleGetExportCsv).Methods(http.MethodPost)

	router.HandleFunc("/medicine", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/history", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/export/csv", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("register medicine", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("register medicine", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode}
		logFile, _ := logger.WriteServerErrorLog("register medicine", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the medicine exists
	temp, err := h.medStore.GetMedicineByName(payload.Name)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode}
		logFile, _ := logger.WriteServerErrorLog("register medicine", user.ID, data, fmt.Errorf("error get medicine name: %v", err))
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
			Message: fmt.Sprintf("Medicine %s already exists", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	firstUnitStr := payload.FirstUnit
	secondUnitStr := payload.SecondUnit
	thirdUnitStr := payload.ThirdUnit

	if payload.FirstUnit == "" {
		firstUnitStr = "None"
	}
	if payload.SecondUnit == "" {
		secondUnitStr = "None"
	}
	if payload.ThirdUnit == "" {
		thirdUnitStr = "None"
	}

	firstUnit, err := h.unitStore.GetUnitByName(firstUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode, "first_unit": firstUnitStr}
		logFile, _ := logger.WriteServerErrorLog("register medicine", user.ID, data, fmt.Errorf("get first unit error: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode, "second_unit": secondUnitStr}
		logFile, _ := logger.WriteServerErrorLog("register medicine", user.ID, data, fmt.Errorf("get first unit error: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode, "third_unit": thirdUnitStr}
		logFile, _ := logger.WriteServerErrorLog("register medicine", user.ID, data, fmt.Errorf("get third unit error: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.medStore.CreateMedicine(types.Medicine{
		Barcode:                    payload.Barcode,
		Name:                       payload.Name,
		Qty:                        payload.Qty,
		FirstUnitID:                firstUnit.ID,
		FirstSubtotal:              payload.FirstSubtotal,
		FirstDiscountPercentage:    payload.FirstDiscountPercentage,
		FirstDiscountAmount:        payload.FirstDiscountAmount,
		FirstPrice:                 payload.FirstPrice,
		SecondUnitID:               secondUnit.ID,
		SecondUnitToFirstUnitRatio: payload.SecondUnitToFirstUnitRatio,
		SecondSubtotal:             payload.SecondSubtotal,
		SecondDiscountPercentage:   payload.SecondDiscountPercentage,
		SecondDiscountAmount:       payload.SecondDiscountAmount,
		SecondPrice:                payload.SecondPrice,
		ThirdUnitID:                thirdUnit.ID,
		ThirdUnitToFirstUnitRatio:  payload.ThirdUnitToFirstUnitRatio,
		ThirdSubtotal:              payload.ThirdSubtotal,
		ThirdDiscountPercentage:    payload.ThirdDiscountPercentage,
		ThirdDiscountAmount:        payload.ThirdDiscountAmount,
		ThirdPrice:                 payload.ThirdPrice,
		Description:                payload.Description,
	}, user.ID)
	if err != nil {
		data := map[string]interface{}{"medicine": payload.Name, "barcode": payload.Barcode}
		logFile, _ := logger.WriteServerErrorLog("register medicine", user.ID, data, fmt.Errorf("error create medicine: %v", err))
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
		Message: fmt.Sprintf("Medicine %s successfully created by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get all medicines", 0, nil, fmt.Errorf("user token invalid: %v", err))
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

	var medicines []types.MedicineListsReturnPayload

	if val == "all" {
		medicines, err = h.medStore.GetAllMedicines()
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, nil, fmt.Errorf("error get all medicines: %v", err))
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
		medicines, err = h.medStore.GetMedicinesBySearchName(val)
		if err != nil {
			data := map[string]interface{}{"searched_medicine": val}
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, data, fmt.Errorf("error get medicine by search name: %v", err))
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
			data := map[string]interface{}{"medicine_id": val}
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, data, fmt.Errorf("error parse medicine id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		medicine, err := h.medStore.GetMedicineByID(id)
		if err != nil {
			data := map[string]interface{}{"medicine_id": id}
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, data, fmt.Errorf("error get medicine by id: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}

		medicines = append(medicines, *medicine)
	} else if params == "barcode" {
		medicines, err = h.medStore.GetMedicinesBySearchBarcode(val)
		if err != nil {
			data := map[string]interface{}{"medicine_barcode": val}
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, data, fmt.Errorf("error get medicine by search barcode: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	} else if params == "description" {
		medicines, err = h.medStore.GetMedicinesByDescription(val)
		if err != nil {
			data := map[string]interface{}{"description": val}
			logFile, _ := logger.WriteServerErrorLog("get all medicines", user.ID, data, fmt.Errorf("error get medicines by description: %v", err))
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
		Result: medicines,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetDetailMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get medicine detail", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get medicine detail", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get medicine detail", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get medicine detail", user.ID, data, fmt.Errorf("error get medicine by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
	}
	if medicine == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Medicine ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: medicine,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("delete medicine", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("delete medicine", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "medicine_name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete medicine", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "medicine_name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete medicine", user.ID, data, fmt.Errorf("error get medicine by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if medicine == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Medicine %s doesn't exist", payload.Name),
		}
		resp.WriteError(w)
		return
	}

	err = h.medStore.DeleteMedicine(&types.Medicine{
		ID:      medicine.ID,
		Barcode: medicine.Barcode,
		Name:    medicine.Name,
	}, user)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "medicine_name": payload.Name}
		logFile, _ := logger.WriteServerErrorLog("delete medicine", user.ID, data, fmt.Errorf("error delete medicine: %v", err))
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
		Message: fmt.Sprintf("Medicine %s deleted by %s", payload.Name, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("modify medicine", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("modify medicine", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get medicine by id: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	if medicine == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Medicine ID %d doesn't exist", payload.ID),
		}
		resp.WriteError(w)
		return
	}

	if medicine.Name != payload.NewData.Name {
		temp, err := h.medStore.GetMedicineByName(payload.NewData.Name)
		if err != nil {
			data := map[string]interface{}{"medicine_id": payload.ID, "medicine_name": payload.NewData.Name}
			logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get medicine by name: %v", err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
		}
		if temp != nil {
			resp := utils.Response{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Medicine %s already exists", payload.NewData.Name),
			}
			resp.WriteError(w)
			return
		}
	}

	if medicine.Barcode != payload.NewData.Barcode {
		temp, err := h.medStore.GetMedicineByBarcode(payload.NewData.Barcode)
		if err != nil {
			data := map[string]interface{}{"medicine_id": payload.ID, "medicine_barcode": payload.NewData.Barcode}
			logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get medicine by barcode: %v", err))
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
				Message: fmt.Sprintf("Medicine barcode %s already exists", payload.NewData.Barcode),
			}
			resp.WriteError(w)
			return
		}
	}

	firstUnitStr := payload.NewData.FirstUnit
	secondUnitStr := payload.NewData.SecondUnit
	thirdUnitStr := payload.NewData.ThirdUnit

	if payload.NewData.FirstUnit == "" {
		firstUnitStr = "None"
	}
	if payload.NewData.SecondUnit == "" {
		secondUnitStr = "None"
	}
	if payload.NewData.ThirdUnit == "" {
		thirdUnitStr = "None"
	}

	firstUnit, err := h.unitStore.GetUnitByName(firstUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "first_unit": firstUnitStr}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get first unit: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "second_unit": secondUnitStr}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get second unit: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "third_unit": thirdUnitStr}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error get third unit: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	err = h.medStore.ModifyMedicine(medicine.ID, types.Medicine{
		Barcode:                    payload.NewData.Barcode,
		Name:                       payload.NewData.Name,
		Qty:                        payload.NewData.Qty,
		FirstUnitID:                firstUnit.ID,
		FirstSubtotal:              payload.NewData.FirstSubtotal,
		FirstDiscountPercentage:    payload.NewData.FirstDiscountPercentage,
		FirstDiscountAmount:        payload.NewData.FirstDiscountAmount,
		FirstPrice:                 payload.NewData.FirstPrice,
		SecondUnitID:               secondUnit.ID,
		SecondUnitToFirstUnitRatio: payload.NewData.SecondUnitToFirstUnitRatio,
		SecondSubtotal:             payload.NewData.SecondSubtotal,
		SecondDiscountPercentage:   payload.NewData.SecondDiscountPercentage,
		SecondDiscountAmount:       payload.NewData.SecondDiscountAmount,
		SecondPrice:                payload.NewData.SecondPrice,
		ThirdUnitID:                thirdUnit.ID,
		ThirdUnitToFirstUnitRatio:  payload.NewData.ThirdUnitToFirstUnitRatio,
		ThirdSubtotal:              payload.NewData.ThirdSubtotal,
		ThirdDiscountPercentage:    payload.NewData.ThirdDiscountPercentage,
		ThirdDiscountAmount:        payload.NewData.ThirdDiscountPercentage,
		ThirdPrice:                 payload.NewData.ThirdPrice,
		Description:                payload.NewData.Description,
	}, user)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("modify medicine", user.ID, data, fmt.Errorf("error modify medicine: %v", err))
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
		Message: fmt.Sprintf("Medicine ID %d modified by %s", payload.ID, user.Name),
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetMedicineHistoryPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get medicine history", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get medicine history", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid Payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID}
		logFile, _ := logger.WriteServerErrorLog("get medicine history", 0, data, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	startDate, err := utils.ParseStartDate(payload.StartDate.Format("2006-01-02 -0700MST"))
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "start_date": payload.StartDate}
		logFile, _ := logger.WriteServerErrorLog("get medicine history", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate.Format("2006-01-02 -0700MST"))
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "start_date": payload.StartDate, "end_date": payload.EndDate}
		logFile, _ := logger.WriteServerErrorLog("get medicine history", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	medicineHistories, err := h.medStore.GetMedicineHistoryByDate(payload.ID, *startDate, *endDate)
	if err != nil {
		data := map[string]interface{}{"medicine_id": payload.ID, "start_date": payload.StartDate}
		logFile, _ := logger.WriteServerErrorLog("get medicine history", user.ID, data, fmt.Errorf("error parse date: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
	}

	resp := utils.Response{
		Code:   http.StatusOK,
		Result: medicineHistories,
	}
	resp.WriteSuccess(w)
}

func (h *Handler) handleGetExportCsv(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RequestExportMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	medicines, err := h.medStore.GetAllMedicines()
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", user.ID, nil, fmt.Errorf("error get all medicines: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if medicines == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "No medicines to export",
		}
		resp.WriteError(w)
		return
	}

	fields := reflect.TypeOf(payload)
	numFields := fields.NumField()
	payloadValues := reflect.ValueOf(payload)

	headers := make([]string, 0)

	for i := 0; i < numFields; i++ {
		if payloadValues.Field(i).Bool() {
			headers = append(headers, fields.Field(i).Tag.Get("csv"))
		}
	}

	fileName := fmt.Sprintf("medicines_%s.csv", time.Now().Local().Format("02-Jan-2006"))
	filePath, writer, f, err := export.CreateCsvWriter(fileName)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", user.ID, nil, fmt.Errorf("error creating csv writer: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}
	defer f.Close()

	err = export.WriteCsvHeader(writer, headers)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine csv", user.ID, nil, fmt.Errorf("error writing csv header: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	for _, med := range medicines {
		data := make([]string, 0)
		dataValues := reflect.ValueOf(med)

		for i := 0; i < numFields; i++ {
			if payloadValues.Field(i).Bool() {
				data = append(data, fmt.Sprint(dataValues.Field(i).Interface()))
			}
		}

		err = export.WriteCsvData(writer, data)
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog("get export medicine csv", user.ID, nil,
				fmt.Errorf("failed to write csv data of medicine %s:\n%v", med.Name, err))
			resp := utils.Response{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Log:     logFile,
				Error:   err.Error(),
			}
			resp.WriteError(w)
			return
		}
	}

	attachment := fmt.Sprintf("attachment; filename=%s", filePath)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, filePath)
}

func (h *Handler) handleGetExportXml(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RequestExportMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine xml", 0, nil, fmt.Errorf("parsing payload failed: %v", err))
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
		logFile, _ := logger.WriteServerErrorLog("get export medicine xml", 0, nil, fmt.Errorf("invalid payload: %v", err))
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid payload",
			Log:     logFile,
			Error:   errors.Error(),
		}
		resp.WriteError(w)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine xml", 0, nil, fmt.Errorf("user token invalid: %v", err))
		resp := utils.Response{
			Code:    http.StatusUnauthorized,
			Message: "User token invalid!\nPlease login again!",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	medicines, err := h.medStore.GetAllMedicines()
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine xml", user.ID, nil, fmt.Errorf("error get medicines: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	if medicines == nil {
		resp := utils.Response{
			Code:    http.StatusBadRequest,
			Message: "No medicines to export",
		}
		resp.WriteError(w)
		return
	}

	fileName := fmt.Sprintf("medicines_%s.xml", time.Now().Local().Format("02-Jan-2006"))
	filePath, err := export.CreateXmlData(fileName, medicines)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog("get export medicine xml", 0, nil, fmt.Errorf("error create xml file: %v", err))
		resp := utils.Response{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Log:     logFile,
			Error:   err.Error(),
		}
		resp.WriteError(w)
		return
	}

	attachment := fmt.Sprintf("attachment; filename=%s", filePath)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, filePath)
}
