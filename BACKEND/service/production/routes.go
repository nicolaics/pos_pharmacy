package production

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	productionStore types.ProductionStore
	userStore       types.UserStore
	medStore        types.MedicineStore
	unitStore       types.UnitStore
}

func NewHandler(productionStore types.ProductionStore,
	userStore types.UserStore,
	medStore types.MedicineStore,
	unitStore types.UnitStore) *Handler {
	return &Handler{
		productionStore: productionStore,
		userStore:       userStore,
		medStore:        medStore,
		unitStore:       unitStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/production", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/production", h.handleGetNumberOfProductions).Methods(http.MethodGet)
	router.HandleFunc("/production/all/date", h.handleGetProductions).Methods(http.MethodPost)
	router.HandleFunc("/production/detail", h.handleGetProductionDetail).Methods(http.MethodPost)
	router.HandleFunc("/production", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/production", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/production", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/production/all/date", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/production/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterProductionPayload

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

	// get produced medicine data
	producedMedicine, err := h.medStore.GetMedicineByBarcode(payload.ProducedMedicineBarcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("med %s not found, create the meds first", payload.ProducedMedicineName))
		return
	}

	err = h.productionStore.CreateProduction(types.Production{
		BatchNumber:          payload.BatchNumber,
		ProducedMedicineID:   producedMedicine.ID,
		ProducedQty:          payload.ProducedQty,
		ProductionDate:       payload.ProductionDate,
		Description:          payload.Description,
		UpdatedToStock:       payload.UpdatedToStock,
		UpdatedToAccount:     payload.UpdatedToAccount,
		TotalCost:            payload.TotalCost,
		UserID:               user.ID,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get production ID
	productionId, err := h.productionStore.GetProductionID(payload.BatchNumber, producedMedicine.ID,
		payload.ProductionDate, payload.TotalCost, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production batch number %d doesn't exists", payload.BatchNumber))
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = h.productionStore.CreateProductionMedicineItems(types.ProductionMedicineItems{
			ProductionID: productionId,
			MedicineID:   medData.ID,
			Qty:          medicine.Qty,
			UnitID:       unit.ID,
			Cost:         medicine.Cost,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("production batch number %d, med %s: %v", payload.BatchNumber, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("production batch number %d successfully created by %s", payload.BatchNumber, user.Name))
}

// beginning of invoice page, will request here
func (h *Handler) handleGetNumberOfProductions(w http.ResponseWriter, r *http.Request) {
	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	numberOfProductions, err := h.productionStore.GetNumberOfProductions()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"nextNumber": (numberOfProductions + 1)})
}

// only view the production list
func (h *Handler) handleGetProductions(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewProductionsPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))

		// TODO: CHECK HERE
		log.Println("not allowed, redirecting")
		http.Redirect(w, r, "/user/login", http.StatusFound)
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	productions, err := h.productionStore.GetProductionsByDate(payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, productions)
}

// view 1 production with its medicine lists
func (h *Handler) handleGetProductionDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewProductionMedicineItemsPayload

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

	// get production data
	production, err := h.productionStore.GetProductionByBatchNumber(payload.BatchNumber)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production batch number %d doesn't exists", payload.BatchNumber))
		return
	}

	// get medicine items of the production
	productionItems, err := h.productionStore.GetProductionMedicineItems(production.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get user data, the one who inputs the production
	inputter, err := h.userStore.GetUserByID(production.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", production.UserID))
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(production.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", production.LastModifiedByUserID))
		return
	}

	// get produced medicine data
	producedMed, err := h.medStore.GetMedicineByID(production.ProducedMedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d doesn't exist", production.ProducedMedicineID))
		return
	}

	returnPayload := types.ProductionDetailPayload{
		ID:          production.ID,
		BatchNumber: production.BatchNumber,

		ProducedMedicine: struct {
			Barcode string "json:\"barcode\""
			Name    string "json:\"name\""
		}{
			Barcode: producedMed.Barcode,
			Name:    producedMed.Name,
		},

		ProducedQty:      production.ProducedQty,
		ProductionDate:   production.ProductionDate,
		Description:      production.Description,
		UpdatedToStock:   production.UpdatedToStock,
		UpdatedToAccount: production.UpdatedToAccount,
		TotalCost: production.TotalCost,

		User: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   inputter.ID,
			Name: inputter.Name,
		},

		CreatedAt:              production.CreatedAt,
		LastModified:           production.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,

		MedicineLists: productionItems,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeleteProduction

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
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err))
		return
	}

	// check if the production exists
	production, err := h.productionStore.GetProductionByID(payload.ID)
	if production == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("production id %d doesn't exist", payload.ID))
		return
	}

	err = h.productionStore.DeleteProductionMedicineItems(production, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.productionStore.DeleteProduction(production, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("production batch number %d deleted by %s", production.BatchNumber, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyProductionPayload

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

	// check if the production exists
	production, err := h.productionStore.GetProductionByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("production with id %d doesn't exists", payload.ID))
		return
	}

	// check duplicate BatchNumber
	prod, err := h.productionStore.GetProductionByBatchNumber(payload.NewData.BatchNumber)
	if err == nil || prod != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("batch number %d exist already", payload.NewData.BatchNumber))
		return
	}

	// get produced medicine data
	producedMedicine, err := h.medStore.GetMedicineByBarcode(payload.NewData.ProducedMedicineBarcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("med %s not found, create the meds first", payload.NewData.ProducedMedicineName))
		return
	}

	err = h.productionStore.ModifyProduction(payload.ID, types.Production{
		BatchNumber:          payload.NewData.BatchNumber,
		ProducedMedicineID:   producedMedicine.ID,
		ProducedQty:          payload.NewData.ProducedQty,
		ProductionDate:       payload.NewData.ProductionDate,
		Description:          payload.NewData.Description,
		UpdatedToStock:       payload.NewData.UpdatedToStock,
		UpdatedToAccount:     payload.NewData.UpdatedToAccount,
		TotalCost:            payload.NewData.TotalCost,
		LastModifiedByUserID: user.ID,
	}, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get production ID
	productionId, err := h.productionStore.GetProductionID(payload.NewData.BatchNumber, producedMedicine.ID,
		payload.NewData.ProductionDate, payload.NewData.TotalCost, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production batch number %d doesn't exists", payload.NewData.BatchNumber))
		return
	}

	err = h.productionStore.DeleteProductionMedicineItems(production, user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	for _, medicine := range payload.NewData.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			unit, err = h.unitStore.GetUnitByName(medicine.Unit)
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = h.productionStore.CreateProductionMedicineItems(types.ProductionMedicineItems{
			ProductionID: productionId,
			MedicineID:   medData.ID,
			Qty:          medicine.Qty,
			UnitID:       unit.ID,
			Cost:         medicine.Cost,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("production batch number %d, med %s: %v", payload.NewData.BatchNumber, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("production modified by %s", user.Name))
}
