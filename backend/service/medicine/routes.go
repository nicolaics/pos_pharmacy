package medicine

import (
	// "encoding/json"
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
	router.HandleFunc("/medicine/detail", h.handleGetOne).Methods(http.MethodPost)
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
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check if the medicine exists
	_, err = h.medStore.GetMedicineByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with name %s already exists", payload.Name), nil)
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
	if firstUnit == nil || err != nil {
		err = h.unitStore.CreateUnit(firstUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create first unit error: %v", err), nil)
			return
		}

		firstUnit, err = h.unitStore.GetUnitByName(firstUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("first unit error: %v", err), nil)
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if secondUnit == nil || err != nil {
		err = h.unitStore.CreateUnit(secondUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create second unit error: %v", err), nil)
			return
		}

		secondUnit, err = h.unitStore.GetUnitByName(secondUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("second unit error: %v", err), nil)
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if thirdUnit == nil || err != nil {
		err = h.unitStore.CreateUnit(thirdUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create third unit error: %v", err), nil)
			return
		}

		thirdUnit, err = h.unitStore.GetUnitByName(thirdUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("third unit error: %v", err), nil)
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
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine %s: %v", payload.Name, err), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("medicine %s successfully created by %s", payload.Name, user.Name), nil)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	var medicines []types.MedicineListsReturnPayload

	if val == "all" {
		medicines, err = h.medStore.GetAllMedicines()
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err, nil)
			return
		}
	} else if params == "name" {
		medicines, err = h.medStore.GetMedicinesBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s not found", val), nil)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		medicine, err := h.medStore.GetMedicineByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d not found", id), nil)
			return
		}

		medicines = append(medicines, *medicine)
	} else if params == "barcode" {
		medicines, err = h.medStore.GetMedicinesBySearchBarcode(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine barcode %s not found", val), nil)
			return
		}
	} else if params == "description" {
		medicines, err = h.medStore.GetMedicinesBySearchBarcode(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine barcode %s not found", val), nil)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown query"), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, medicines, nil)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOneMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if medicine == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine id %d doesn't exist", payload.ID), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, medicine, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if medicine == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine %s doesn't exist", payload.Name), nil)
		return
	}

	err = h.medStore.DeleteMedicine(&types.Medicine{
		ID:      medicine.ID,
		Barcode: medicine.Barcode,
		Name:    medicine.Name,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("medicine %s deleted by %s", payload.Name, user.Name), nil)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil || medicine == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with id %d doesn't exists", payload.ID), nil)
		return
	}

	if medicine.Name != payload.NewData.Name {
		_, err = h.medStore.GetMedicineByName(payload.NewData.Name)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with name %s already exist", payload.NewData.Name), nil)
			return
		}
	}

	if medicine.Barcode != payload.NewData.Barcode {
		_, err = h.medStore.GetMedicineByBarcode(payload.NewData.Barcode)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with barcode %s already exist", payload.NewData.Barcode), nil)
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
	if firstUnit == nil {
		err = h.unitStore.CreateUnit(firstUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		firstUnit, err = h.unitStore.GetUnitByName(firstUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if secondUnit == nil {
		err = h.unitStore.CreateUnit(secondUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		secondUnit, err = h.unitStore.GetUnitByName(secondUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if thirdUnit == nil {
		err = h.unitStore.CreateUnit(thirdUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		thirdUnit, err = h.unitStore.GetUnitByName(thirdUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Sprintf("medicine modified into %s by %s",
		payload.NewData.Name, user.Name), nil)
}

func (h *Handler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetMedicineHistoryPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	startDate, err := utils.ParseStartDate(payload.StartDate.Format("2006-01-02 -0700MST"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parse start date: %v", err), nil)
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate.Format("2006-01-02 -0700MST"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parse end date: %v", err), nil)
		return
	}

	medicineHistories, err := h.medStore.GetMedicineHistoryByDate(payload.ID, *startDate, *endDate)
	if medicineHistories == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine id %d doesn't exist", payload.ID), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, medicineHistories, nil)
}

func (h *Handler) handleGetExportCsv(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RequestExportMedicinePayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	medicines, err := h.medStore.GetAllMedicines()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	if medicines == nil {
		utils.WriteSuccess(w, http.StatusOK, "no medicine to export", nil)
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
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating csv writer: %v", err), nil)
		return
	}
	defer f.Close()

	err = export.WriteCsvHeader(writer, headers)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error writing csv header: %v", err), nil)
		return
	}

	logFiles := make([]string, 0)

	for _, med := range(medicines) {
		data := make([]string, 0)
		dataValues := reflect.ValueOf(med)

		for i := 0; i < numFields; i++ {
			if payloadValues.Field(i).Bool() {
				data = append(data, fmt.Sprint(dataValues.Field(i).Interface()))
			}
		}

		err = export.WriteCsvData(writer, data)
		if err != nil {
			logFile, _ := logger.WriteServerErrorLog(fmt.Errorf("failed to write csv data of medicine %s:\n%v", med.Name, err))
			logFiles = append(logFiles, logFile)
		}
	}

	attachment := fmt.Sprintf("attachment; filename=%s", filePath)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", attachment)
	w.WriteHeader(http.StatusOK)

	http.ServeFile(w, r, filePath)

	// utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("export success to %s", fileName), logFiles)
}
