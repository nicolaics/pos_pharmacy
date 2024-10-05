package maindoctorprescmeditem

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	mainDoctorPrescMedItemStore types.MainDoctorPrescMedItemStore
	userStore                   types.UserStore
	medStore                    types.MedicineStore
	unitStore                   types.UnitStore
}

func NewHandler(mainDoctorPrescMedItemStore types.MainDoctorPrescMedItemStore,
	userStore types.UserStore,
	medStore types.MedicineStore,
	unitStore types.UnitStore) *Handler {
	return &Handler{
		mainDoctorPrescMedItemStore: mainDoctorPrescMedItemStore,
		userStore:                   userStore,
		medStore:                    medStore,
		unitStore:                   unitStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/main-doctor-prescription-item", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/main-doctor-prescription-item/{val}", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/main-doctor-prescription-item/detail", h.handleGetDetail).Methods(http.MethodPost)
	router.HandleFunc("/main-doctor-prescription-item", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/main-doctor-prescription-item", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/main-doctor-prescription-item/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/main-doctor-prescription-item/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterMainDoctorPrescMedItemPayload

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

	// get unit
	unit, err := h.unitStore.GetUnitByName("KAP")
	if unit == nil {
		err = h.unitStore.CreateUnit("KAP")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating unit: %v", err))
			return
		}

		unit, err = h.unitStore.GetUnitByName("KAP")
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	noneUnit, err := h.unitStore.GetUnitByName("None")
	if unit == nil {
		err = h.unitStore.CreateUnit("None")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating unit: %v", err))
			return
		}

		unit, err = h.unitStore.GetUnitByName("None")
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get medicine data
	medicine, err := h.medStore.GetMedicineByName(payload.MedicineName)
	if medicine == nil {
		barcode := "RO-" + utils.GenerateRandomCodeNumbers(3)
		isBarcodeExist, _ := h.mainDoctorPrescMedItemStore.IsMedicineBarcodeExist(barcode)

		for isBarcodeExist {
			barcode = "RO-" + utils.GenerateRandomCodeNumbers(3)
			isBarcodeExist, _ = h.mainDoctorPrescMedItemStore.IsMedicineBarcodeExist(barcode)
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
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating medicine: %v", err))
			return
		}

		medicine, err = h.medStore.GetMedicineByName(payload.MedicineName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err))
		return
	}

	// check duplicate
	isExist, err := h.mainDoctorPrescMedItemStore.IsMedicineContentsExist(medicine.ID)
	if err != nil || isExist {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine already exist: %v", err))
		return
	}

	for _, medContent := range payload.MedicineContents {
		// get medContent data
		medData, err := h.medStore.GetMedicineByName(medContent.Name)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medContent.Name))
			return
		}

		medUnit, err := h.unitStore.GetUnitByName(medContent.Unit)
		if medUnit == nil {
			err = h.unitStore.CreateUnit(medContent.Unit)
			if err == nil {
				medUnit, err = h.unitStore.GetUnitByName(medContent.Unit)
			}
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		fractionIdx := strings.Index(medContent.Qty, "/")
		var medicineQty float64
		if fractionIdx == -1 {
			medicineQty, err = strconv.ParseFloat(medContent.Qty, 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err))
				return
			}
		} else {
			fraction := strings.Split(medContent.Qty, "/")
			numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
			denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

			medicineQty = numerator / denum
		}

		medItem := types.MainDoctorPrescMedItem{
			MedicineID:           medicine.ID,
			MedicineContentID:    medData.ID,
			Qty:                  medicineQty,
			UnitID:               medUnit.ID,
			UserID:               user.ID,
			LastModifiedByUserID: user.ID,
		}
		err = h.mainDoctorPrescMedItemStore.CreateMainDoctorPrescMedItem(medItem)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("error creating presc med item %s for content %s: %v", medicine.Name, medData.Name, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Errorf("%s created by %s", payload.MedicineName, user.Name))
}

// view all
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	vars := mux.Vars(r)
	val := vars["val"]

	var mainDoctorPrescMedItems []types.MainDoctorPrescMedItemReturn

	if val == "all" {
		mainDoctorPrescMedItems, err = h.mainDoctorPrescMedItemStore.GetAllMainDoctorPrescMedItemByMedicineData()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		medicine, err := h.medStore.GetMedicineByName(val)
		if err != nil || medicine == nil{
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine not found"))
			return
		}

		mainDoctorPrescMedItem, err := h.mainDoctorPrescMedItemStore.GetMainDoctorPrescMedItemByMedicineData(medicine.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		mainDoctorPrescMedItems = append(mainDoctorPrescMedItems, *mainDoctorPrescMedItem)
	}

	utils.WriteJSON(w, http.StatusOK, mainDoctorPrescMedItems)
}

// view one
func (h *Handler) handleGetDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewMainDoctorPrescMedItemPayload

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

	// get data
	data, err := h.mainDoctorPrescMedItemStore.GetMainDoctorPrescMedItemByMedicineData(payload.MedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine not found: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, data)
}

// only can modify the contents
func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyMainDoctorPrescMedItemPayload

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

	// get medicine data
	medicine, err := h.medStore.GetMedicineByID(payload.MedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting medicine: %v", err))
		return
	}

	err = h.mainDoctorPrescMedItemStore.DeleteMainDoctorPrescMedItem(medicine.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting: %v", err))
		return
	}

	for _, medContent := range payload.NewMedicineContents {
		// get medContent data
		medData, err := h.medStore.GetMedicineByName(medContent.Name)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medContent.Name))
			return
		}

		medUnit, err := h.unitStore.GetUnitByName(medContent.Unit)
		if medUnit == nil {
			err = h.unitStore.CreateUnit(medContent.Unit)
			if err == nil {
				medUnit, err = h.unitStore.GetUnitByName(medContent.Unit)
			}
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		fractionIdx := strings.Index(medContent.Qty, "/")
		var medicineQty float64
		if fractionIdx == -1 {
			medicineQty, err = strconv.ParseFloat(medContent.Qty, 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err))
				return
			}
		} else {
			fraction := strings.Split(medContent.Qty, "/")
			numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
			denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

			medicineQty = numerator / denum
		}

		medItem := types.MainDoctorPrescMedItem{
			MedicineID:           medicine.ID,
			MedicineContentID:    medData.ID,
			Qty:                  medicineQty,
			UnitID:               medUnit.ID,
			UserID:               user.ID,
			LastModifiedByUserID: user.ID,
		}
		err = h.mainDoctorPrescMedItemStore.CreateMainDoctorPrescMedItem(medItem)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("error creating presc med item %s for content %s: %v", medicine.Name, medData.Name, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Errorf("%s created by %s", medicine.Name, user.Name))
}
