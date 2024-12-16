package mdmi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
)

type Handler struct {
	mainDoctorMedItemStore types.MainDoctorMedItemStore
	userStore              types.UserStore
	medStore               types.MedicineStore
	unitStore              types.UnitStore
}

func NewHandler(mainDoctorMedItemStore types.MainDoctorMedItemStore,
	userStore types.UserStore,
	medStore types.MedicineStore,
	unitStore types.UnitStore) *Handler {
	return &Handler{
		mainDoctorMedItemStore: mainDoctorMedItemStore,
		userStore:              userStore,
		medStore:               medStore,
		unitStore:              unitStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/main-doctor-prescription-item", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/main-doctor-prescription-item/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/main-doctor-prescription-item/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/main-doctor-prescription-item", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/main-doctor-prescription-item/test", h.handleTest).Methods(http.MethodPost)

	router.HandleFunc("/main-doctor-prescription-item", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/main-doctor-prescription-item/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/main-doctor-prescription-item/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/main-doctor-prescription-item/test", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterMainDoctorMedItemPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get unit
	unit, err := h.unitStore.GetUnitByName("KAP")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	noneUnit, err := h.unitStore.GetUnitByName("None")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByName(payload.MedicineName)
	if medicine == nil {
		barcode := "RO-" + utils.GenerateRandomCodeNumbers(3)
		isBarcodeExist, _ := h.mainDoctorMedItemStore.IsMedicineBarcodeExist(barcode)

		for isBarcodeExist {
			barcode = "RO-" + utils.GenerateRandomCodeNumbers(3)
			isBarcodeExist, _ = h.mainDoctorMedItemStore.IsMedicineBarcodeExist(barcode)
		}

		err = h.medStore.CreateMedicine(types.Medicine{
			Barcode:              barcode,
			Name:                 strings.ToUpper(payload.MedicineName),
			Qty:                  0.0,
			FirstUnitID:          unit.ID,
			FirstSubtotal:        0.0,
			FirstPrice:           0.0,
			SecondUnitID:         noneUnit.ID,
			ThirdUnitID:          noneUnit.ID,
			LastModifiedByUserID: user.ID,
		}, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating medicine: %v", err), nil)
			return
		}

		medicine, err = h.medStore.GetMedicineByName(payload.MedicineName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err), nil)
		return
	}

	// check duplicate
	isExist, err := h.mainDoctorMedItemStore.IsMedicineContentsExist(medicine.ID)
	if err != nil || isExist {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine already exist: %v", err), nil)
		return
	}

	for _, medContent := range payload.MedicineContents {
		// get medContent data
		medData, err := h.medStore.GetMedicineByName(medContent.Name)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medContent.Name), nil)
			return
		}

		medUnit, err := h.unitStore.GetUnitByName(medContent.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		fractionIdx := strings.Index(medContent.Qty, "/")
		var medicineQty float64
		if fractionIdx == -1 {
			medicineQty, err = strconv.ParseFloat(medContent.Qty, 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err), nil)
				return
			}
		} else {
			fraction := strings.Split(medContent.Qty, "/")
			numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
			denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

			medicineQty = numerator / denum
		}

		medItem := types.MainDoctorMedItem{
			MedicineID:           medicine.ID,
			MedicineContentID:    medData.ID,
			Qty:                  medicineQty,
			UnitID:               medUnit.ID,
			UserID:               user.ID,
			LastModifiedByUserID: user.ID,
		}
		err = h.mainDoctorMedItemStore.CreateMainDoctorMedItem(medItem)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("error creating presc med item %s for content %s: %v", medicine.Name, medData.Name, err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Errorf("%s created by %s", payload.MedicineName, user.Name), nil)
}

// view all
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	vars := mux.Vars(r)
	val := vars["val"]

	var mainDoctorMedItems []types.MainDoctorMedItemReturn

	if val == "all" {
		mainDoctorMedItems, err = h.mainDoctorMedItemStore.GetAllMainDoctorMedItemByMedicineData()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else {
		medicine, err := h.medStore.GetMedicineByName(val)
		if err != nil || medicine == nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine not found"), nil)
			return
		}

		mainDoctorMedItem, err := h.mainDoctorMedItemStore.GetMainDoctorMedItemByMedicineData(medicine.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		mainDoctorMedItems = append(mainDoctorMedItems, *mainDoctorMedItem)
	}

	utils.WriteSuccess(w, http.StatusOK, mainDoctorMedItems, nil)
}

// view one
func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewMainDoctorMedItemPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get data
	data, err := h.mainDoctorMedItemStore.GetMainDoctorMedItemByMedicineData(payload.MedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine not found: %v", err), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, data, nil)
}

// only can modify the contents
func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyMainDoctorMedItemPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByID(payload.MedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err), nil)
		return
	}

	err = h.mainDoctorMedItemStore.DeleteMainDoctorMedItem(medicine.ID, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting: %v", err), nil)
		return
	}

	for _, medContent := range payload.NewMedicineContents {
		// get medContent data
		medData, err := h.medStore.GetMedicineByName(medContent.Name)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medContent.Name), nil)
			return
		}

		medUnit, err := h.unitStore.GetUnitByName(medContent.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		fractionIdx := strings.Index(medContent.Qty, "/")
		var medicineQty float64
		if fractionIdx == -1 {
			medicineQty, err = strconv.ParseFloat(medContent.Qty, 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err), nil)
				return
			}
		} else {
			fraction := strings.Split(medContent.Qty, "/")
			numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
			denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

			medicineQty = numerator / denum
		}

		medItem := types.MainDoctorMedItem{
			MedicineID:           medicine.ID,
			MedicineContentID:    medData.ID,
			Qty:                  medicineQty,
			UnitID:               medUnit.ID,
			UserID:               user.ID,
			LastModifiedByUserID: user.ID,
		}
		err = h.mainDoctorMedItemStore.CreateMainDoctorMedItem(medItem)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("error creating presc med item %s for content %s: %v", medicine.Name, medData.Name, err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Errorf("%s created by %s", medicine.Name, user.Name), nil)
}

func (h *Handler) handleTest(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterMainDoctorMedItemPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("parsing payload failed: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing payload failed\n(%s)", logFile))
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("invalid payload: %v", errors))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload\n(%s)", logFile))
		return
	}

	// validate token
	user, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		logFile, _ := logger.WriteServerErrorLog(fmt.Sprintf("user token invalid: %v", err))
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid\nPlease log in again"))
		return
	}

	// get unit
	unit, err := h.unitStore.GetUnitByName("KAP")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	noneUnit, err := h.unitStore.GetUnitByName("None")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByName(payload.MedicineName)
	if medicine == nil {
		barcode := "RO-" + utils.GenerateRandomCodeNumbers(3)
		isBarcodeExist, _ := h.mainDoctorMedItemStore.IsMedicineBarcodeExist(barcode)

		for isBarcodeExist {
			barcode = "RO-" + utils.GenerateRandomCodeNumbers(3)
			isBarcodeExist, _ = h.mainDoctorMedItemStore.IsMedicineBarcodeExist(barcode)
		}

		err = h.medStore.CreateMedicine(types.Medicine{
			Barcode:              barcode,
			Name:                 strings.ToUpper(payload.MedicineName),
			Qty:                  0.0,
			FirstUnitID:          unit.ID,
			FirstSubtotal:        0.0,
			FirstPrice:           0.0,
			SecondUnitID:         noneUnit.ID,
			ThirdUnitID:          noneUnit.ID,
			LastModifiedByUserID: user.ID,
		}, user.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating medicine: %v", err), nil)
			return
		}

		medicine, err = h.medStore.GetMedicineByName(payload.MedicineName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err), nil)
		return
	}

	// check duplicate
	isExist, err := h.mainDoctorMedItemStore.IsMedicineContentsExist(medicine.ID)
	if err != nil || isExist {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine already exist: %v", err), nil)
		return
	}

	for _, medContent := range payload.MedicineContents {
		medUnit, err := h.unitStore.GetUnitByName(medContent.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		// get medContent data
		medData, err := h.medStore.GetMedicineByName(medContent.Name)
		if medData == nil {
			barcode := "TEST-" + utils.GenerateRandomCodeNumbers(3)
			err = h.medStore.CreateMedicine(types.Medicine{
				Barcode:              barcode,
				Name:                 strings.ToUpper(medContent.Name),
				Qty:                  0.0,
				FirstUnitID:          medUnit.ID,
				FirstSubtotal:        0.0,
				FirstPrice:           0.0,
				SecondUnitID:         noneUnit.ID,
				ThirdUnitID:          noneUnit.ID,
				LastModifiedByUserID: user.ID,
			}, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating medicine: %v", err), nil)
				return
			}

			medData, err = h.medStore.GetMedicineByName(medContent.Name)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err), nil)
			return
		}

		fractionIdx := strings.Index(medContent.Qty, "/")
		var medicineQty float64
		if fractionIdx == -1 {
			medicineQty, err = strconv.ParseFloat(medContent.Qty, 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err), nil)
				return
			}
		} else {
			fraction := strings.Split(medContent.Qty, "/")
			numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
			denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

			medicineQty = numerator / denum
		}

		medItem := types.MainDoctorMedItem{
			MedicineID:           medicine.ID,
			MedicineContentID:    medData.ID,
			Qty:                  medicineQty,
			UnitID:               medUnit.ID,
			UserID:               user.ID,
			LastModifiedByUserID: user.ID,
		}
		err = h.mainDoctorMedItemStore.CreateMainDoctorMedItem(medItem)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("error creating presc med item %s for content %s: %v", medicine.Name, medData.Name, err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusCreated, fmt.Errorf("%s created by %s", payload.MedicineName, user.Name), nil)
}
