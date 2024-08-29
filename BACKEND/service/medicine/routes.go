package medicine

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	medStore     types.MedicineStore
	cashierStore types.CashierStore
}

func NewHandler(medStore types.MedicineStore, cashierStore types.CashierStore) *Handler {
	return &Handler{medStore: medStore, cashierStore: cashierStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/medicine", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/medicine", h.handleGetAll).Methods(http.MethodGet)
	router.HandleFunc("/medicine", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/medicine", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/medicine", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the medicine exists
	_, err = h.medStore.GetMedicineByName(payload.Name)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with name %s already exists", payload.Name))
		return
	}

	err = h.medStore.CreateMedicine(types.Medicine{
		Barcode:        payload.Barcode,
		Name:           payload.Name,
		Qty:            payload.Qty,
		FirstUnitID:    payload.FirstUnitID,
		FirstSubtotal:  payload.FirstSubtotal,
		FirstDiscount:  payload.FirstDiscount,
		FirstPrice:     payload.FirstPrice,
		SecondUnitID:   payload.SecondUnitID,
		SecondSubtotal: payload.SecondSubtotal,
		SecondDiscount: payload.SecondDiscount,
		SecondPrice:    payload.SecondPrice,
		ThirdUnitID:    payload.ThirdUnitID,
		ThirdSubtotal:  payload.ThirdSubtotal,
		ThirdDiscount:  payload.ThirdDiscount,
		ThirdPrice:     payload.ThirdPrice,
		Description:    payload.Description,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("medicine %s successfully created by %s", payload.Name, cashier.Name))
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	medicines, err := h.medStore.GetAllMedicines()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, medicines)
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
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if medicine == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine %s doesn't exist", payload.Name))
		return
	}

	err = h.medStore.DeleteMedicine(medicine)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("medicine %s deleted by %s", payload.Name, cashier.Name))
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
	cashier, err := h.cashierStore.ValidateCashierToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cashier token invalid"))
		return
	}

	// check if the medicine exists
	medicine, err := h.medStore.GetMedicineByID(payload.ID)
	if err != nil || medicine == nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("medicine with id %d doesn't exists", payload.ID))
		return
	}

	if medicine.Name != payload.NewName {
		_, err = h.medStore.GetMedicineByName(payload.NewName)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with name %s already exist", payload.NewName))
			return
		}
	}

	if medicine.Barcode != payload.NewBarcode {
		_, err = h.medStore.GetMedicineByBarcode(payload.NewBarcode)
		if err == nil {
			utils.WriteError(w, http.StatusBadRequest,
				fmt.Errorf("medicine with barcode %s already exist", payload.NewBarcode))
			return
		}
	}

	err = h.medStore.ModifyMedicine(medicine.ID, types.Medicine{
		Barcode:        payload.NewBarcode,
		Name:           payload.NewName,
		Qty:            payload.NewQty,
		FirstUnitID:    payload.NewFirstUnitID,
		FirstSubtotal:  payload.NewFirstSubtotal,
		FirstDiscount:  payload.NewFirstDiscount,
		FirstPrice:     payload.NewFirstPrice,
		SecondUnitID:   payload.NewSecondUnitID,
		SecondSubtotal: payload.NewSecondSubtotal,
		SecondDiscount: payload.NewSecondDiscount,
		SecondPrice:    payload.NewSecondPrice,
		ThirdUnitID:    payload.NewThirdUnitID,
		ThirdSubtotal:  payload.NewThirdSubtotal,
		ThirdDiscount:  payload.NewThirdDiscount,
		ThirdPrice:     payload.NewThirdPrice,
		Description:    payload.NewDescription,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("medicine modified into %s by %s",
		payload.NewName, cashier.Name))
}
