package medicine

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
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

	router.HandleFunc("/medicine", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/medicine/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterMedicinePayload

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

	// check if the medicine exists
	_, err = h.medStore.GetMedicineByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with name %s already exists", payload.Name))
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
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create first unit error: %v", err))
			return
		}

		firstUnit, err = h.unitStore.GetUnitByName(firstUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("first unit error: %v", err))
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if secondUnit == nil || err != nil {
		err = h.unitStore.CreateUnit(secondUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create second unit error: %v", err))
			return
		}

		secondUnit, err = h.unitStore.GetUnitByName(secondUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("second unit error: %v", err))
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if thirdUnit == nil || err != nil {
		err = h.unitStore.CreateUnit(thirdUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("create third unit error: %v", err))
			return
		}

		thirdUnit, err = h.unitStore.GetUnitByName(thirdUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("third unit error: %v", err))
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
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine %s: %v", payload.Name, err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("medicine %s successfully created by %s", payload.Name, user.Name))
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

	var medicines []types.MedicineListsReturnPayload

	if val == "all" {
		medicines, err = h.medStore.GetAllMedicines()
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}
	} else if params == "name" {
		medicines, err = h.medStore.GetMedicinesBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s not found", val))
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		medicine, err := h.medStore.GetMedicineByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d not found", id))
			return
		}

		medicines = append(medicines, *medicine)
	} else if params == "barcode" {
		medicines, err = h.medStore.GetMedicinesBySearchBarcode(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine barcode %s not found", val))
			return
		}
	} else if params == "description" {
		medicines, err = h.medStore.GetMedicinesBySearchBarcode(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine barcode %s not found", val))
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown query"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, medicines)
}

func (h *Handler) handleGetOne(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.GetOneMedicinePayload

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

	// get medicine data
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if medicine == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine id %d doesn't exist", payload.ID))
		return
	}

	utils.WriteJSON(w, http.StatusOK, medicine)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteMedicinePayload

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

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if medicine == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine %s doesn't exist", payload.Name))
		return
	}

	err = h.medStore.DeleteMedicine(&types.Medicine{
		ID:      medicine.ID,
		Barcode: medicine.Barcode,
		Name:    medicine.Name,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("medicine %s deleted by %s", payload.Name, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyMedicinePayload

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

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil || medicine == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with id %d doesn't exists", payload.ID))
		return
	}

	if medicine.Name != payload.NewData.Name {
		_, err = h.medStore.GetMedicineByName(payload.NewData.Name)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with name %s already exist", payload.NewData.Name))
			return
		}
	}

	if medicine.Barcode != payload.NewData.Barcode {
		_, err = h.medStore.GetMedicineByBarcode(payload.NewData.Barcode)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with barcode %s already exist", payload.NewData.Barcode))
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
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		firstUnit, err = h.unitStore.GetUnitByName(firstUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	secondUnit, err := h.unitStore.GetUnitByName(secondUnitStr)
	if secondUnit == nil {
		err = h.unitStore.CreateUnit(secondUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		secondUnit, err = h.unitStore.GetUnitByName(secondUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	thirdUnit, err := h.unitStore.GetUnitByName(thirdUnitStr)
	if thirdUnit == nil {
		err = h.unitStore.CreateUnit(thirdUnitStr)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		thirdUnit, err = h.unitStore.GetUnitByName(thirdUnitStr)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
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
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("medicine modified into %s by %s",
		payload.NewData.Name, user.Name))
}
