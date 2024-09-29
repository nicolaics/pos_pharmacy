package production

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
	router.HandleFunc("/production/{params}/{val}", h.handleGetProductions).Methods(http.MethodPost)
	router.HandleFunc("/production/detail", h.handleGetProductionDetail).Methods(http.MethodPost)
	router.HandleFunc("/production", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/production", h.handleModify).Methods(http.MethodPatch)

	router.HandleFunc("/production", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/production/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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

	prodDate, err := utils.ParseDate(payload.ProductionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	// get produced unit ID
	producedUnit, err := h.unitStore.GetUnitByName(payload.ProducedUnit)
	if producedUnit == nil {
		err = h.unitStore.CreateUnit(payload.ProducedUnit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		producedUnit, err = h.unitStore.GetUnitByName(payload.ProducedUnit)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// check duplicate
	production, err := h.productionStore.GetProductionByNumber(payload.Number)
	if err == nil || production != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production number %d exists", payload.Number))
		return
	}

	err = h.productionStore.CreateProduction(types.Production{
		Number:               payload.Number,
		ProducedMedicineID:   producedMedicine.ID,
		ProducedQty:          payload.ProducedQty,
		ProducedUnitID:       producedUnit.ID,
		ProductionDate:       *prodDate,
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

	// add to stock
	if payload.UpdatedToStock {
		err = addStock(h, producedMedicine, producedUnit, float64(payload.ProducedQty), user)
		if err != nil {
			err = h.productionStore.AbsoluteDeleteProduction(types.Production{
				Number:               payload.Number,
				ProducedMedicineID:   producedMedicine.ID,
				ProducedQty:          payload.ProducedQty,
				ProducedUnitID:       producedUnit.ID,
				ProductionDate:       *prodDate,
				Description:          payload.Description,
				UpdatedToStock:       payload.UpdatedToStock,
				UpdatedToAccount:     payload.UpdatedToAccount,
				TotalCost:            payload.TotalCost,
			})
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete production: %v", err))
				return	
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}
	}

	// get production ID
	production, err = h.productionStore.GetProductionByNumber(payload.Number)
	if err != nil {
		err = h.productionStore.AbsoluteDeleteProduction(types.Production{
			Number:               payload.Number,
			ProducedMedicineID:   producedMedicine.ID,
			ProducedQty:          payload.ProducedQty,
			ProducedUnitID:       producedUnit.ID,
			ProductionDate:       *prodDate,
			Description:          payload.Description,
			UpdatedToStock:       payload.UpdatedToStock,
			UpdatedToAccount:     payload.UpdatedToAccount,
			TotalCost:            payload.TotalCost,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete production: %v", err))
			return	
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production number %d doesn't exists: %v", payload.Number, err))
		return
	}

	for _, medicine := range payload.MedicineLists {
		medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
		if err != nil {
			err = h.productionStore.AbsoluteDeleteProduction(types.Production{
				Number:               payload.Number,
				ProducedMedicineID:   producedMedicine.ID,
				ProducedQty:          payload.ProducedQty,
				ProducedUnitID:       producedUnit.ID,
				ProductionDate:       *prodDate,
				Description:          payload.Description,
				UpdatedToStock:       payload.UpdatedToStock,
				UpdatedToAccount:     payload.UpdatedToAccount,
				TotalCost:            payload.TotalCost,
			})
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete production: %v", err))
				return	
			}

			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicine.Unit)
		if unit == nil {
			err = h.unitStore.CreateUnit(medicine.Unit)
			if err != nil {
				err = h.productionStore.AbsoluteDeleteProduction(types.Production{
					Number:               payload.Number,
					ProducedMedicineID:   producedMedicine.ID,
					ProducedQty:          payload.ProducedQty,
					ProducedUnitID:       producedUnit.ID,
					ProductionDate:       *prodDate,
					Description:          payload.Description,
					UpdatedToStock:       payload.UpdatedToStock,
					UpdatedToAccount:     payload.UpdatedToAccount,
					TotalCost:            payload.TotalCost,
				})
				if err != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete production: %v", err))
					return	
				}

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
			ProductionID: production.ID,
			MedicineID:   medData.ID,
			Qty:          medicine.Qty,
			UnitID:       unit.ID,
			Cost:         medicine.Cost,
		})
		if err != nil {
			err = h.productionStore.AbsoluteDeleteProduction(types.Production{
				Number:               payload.Number,
				ProducedMedicineID:   producedMedicine.ID,
				ProducedQty:          payload.ProducedQty,
				ProducedUnitID:       producedUnit.ID,
				ProductionDate:       *prodDate,
				Description:          payload.Description,
				UpdatedToStock:       payload.UpdatedToStock,
				UpdatedToAccount:     payload.UpdatedToAccount,
				TotalCost:            payload.TotalCost,
			})
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete production: %v", err))
				return	
			}
			
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("production number %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("production number %d successfully created by %s", payload.Number, user.Name))
}

// beginning of production page, will request here
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
		return
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err))
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	var prods []types.ProductionListsReturnPayload

	if val == "all" {
		prods, err = h.productionStore.GetProductionsByDate(*startDate, *endDate)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		prod, err := h.productionStore.GetProductionByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prod id %d not exist", id))
			return
		}

		user, err := h.userStore.GetUserByID(prod.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user id %d not found", prod.UserID))
			return
		}

		med, err := h.medStore.GetMedicineByID(prod.ProducedMedicineID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("medicine id %d not found", prod.ProducedMedicineID))
			return
		}

		prods = append(prods, types.ProductionListsReturnPayload{
			ID:                   prod.ID,
			Number:               prod.Number,
			ProducedMedicineName: med.Name,
			ProducedQty:          prod.ProducedQty,
			ProductionDate:       prod.ProductionDate,
			Description:          prod.Description,
			UpdatedToStock:       prod.UpdatedToStock,
			UpdatedToAccount:     prod.UpdatedToAccount,
			TotalCost:            prod.TotalCost,
			UserName:             user.Name,
		})
	} else if params == "batch-number" {
		batchNumber, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		prods, err = h.productionStore.GetProductionsByDateAndNumber(*startDate, *endDate, batchNumber)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "user" {
		users, err := h.userStore.GetUserBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val))
			return
		}

		for _, user := range users {
			temp, err := h.productionStore.GetProductionsByDateAndUserID(*startDate, *endDate, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s doesn't create any prod between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			prods = append(prods, temp...)
		}
	} else if params == "produced-medicine-name" {
		medicines, err := h.medStore.GetMedicinesBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s not exists", val))
			return
		}

		for _, medicine := range medicines {
			temp, err := h.productionStore.GetProductionsByDateAndMedicineID(*startDate, *endDate, medicine.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't have any production between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			prods = append(prods, temp...)
		}
	} else if params == "updated-to-stock" {
		var uts bool

		if val == "true" {
			uts = true
		} else {
			uts = false
		}

		prods, err = h.productionStore.GetProductionsByDateAndUpdatedToStock(*startDate, *endDate, uts)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else if params == "updated-to-account" {
		var uta bool

		if val == "true" {
			uta = true
		} else {
			uta = false
		}

		prods, err = h.productionStore.GetProductionsByDateAndUpdatedToAccount(*startDate, *endDate, uta)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, prods)
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
	production, err := h.productionStore.GetProductionByNumber(payload.Number)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production number %d doesn't exists", payload.Number))
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
		ID:     production.ID,
		Number: production.Number,

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
		TotalCost:        production.TotalCost,

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

	// get produced medicine
	tempProducedMedicine, err := h.medStore.GetMedicineByID(production.ProducedMedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d doesn't exists", production.ProducedMedicineID))
		return
	}

	producedMedicine, err := h.medStore.GetMedicineByBarcode(tempProducedMedicine.Barcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d doesn't exists", production.ProducedMedicineID))
		return
	}

	// get produced unit ID
	producedUnit, err := h.unitStore.GetUnitByID(production.ProducedUnitID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("old produced unit id %d not found", production.ProducedUnitID))
		return
	}
	

	err = h.productionStore.DeleteProductionMedicineItems(production, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.productionStore.DeleteProduction(production, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// reset the previous stock
	if production.UpdatedToStock {
		err = subtractStock(h, producedMedicine, producedUnit, float64(production.ProducedQty), user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error subtracting stock: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("production number %d deleted by %s", production.Number, user.Name))
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
	oldProduction, err := h.productionStore.GetProductionByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("production with id %d doesn't exists", payload.ID))
		return
	}

	// get old produced medicine
	tempOldProducedMedicine, err := h.medStore.GetMedicineByID(oldProduction.ProducedMedicineID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d doesn't exists", oldProduction.ProducedMedicineID))
		return
	}

	oldProducedMedicine, err := h.medStore.GetMedicineByBarcode(tempOldProducedMedicine.Barcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine id %d doesn't exists", oldProduction.ProducedMedicineID))
		return
	}
	
	// get old produced unit ID
	oldProducedUnit, err := h.unitStore.GetUnitByID(oldProduction.ProducedUnitID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("old produced unit id %d not found", oldProduction.ProducedUnitID))
		return
	}

	// check duplicate Number
	prod, err := h.productionStore.GetProductionByNumber(payload.NewData.Number)
	if err == nil || prod != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("number %d exist already", payload.NewData.Number))
		return
	}

	// reset the previous stock
	if oldProduction.UpdatedToStock {
		err = subtractStock(h, oldProducedMedicine, oldProducedUnit, float64(oldProduction.ProducedQty), user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error subtracting stock: %v", err))
			return
		}
	}

	newProducedMedicine, err := h.medStore.GetMedicineByBarcode(payload.NewData.ProducedMedicineBarcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", payload.NewData.ProducedMedicineName))
		return
	}

	// get produced unit ID
	newProducedUnit, err := h.unitStore.GetUnitByName(payload.NewData.ProducedUnit)
	if newProducedUnit == nil {
		err = h.unitStore.CreateUnit(payload.NewData.ProducedUnit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		newProducedUnit, err = h.unitStore.GetUnitByName(payload.NewData.ProducedUnit)
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get produced medicine data
	producedMedicine, err := h.medStore.GetMedicineByBarcode(payload.NewData.ProducedMedicineBarcode)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("med %s not found, create the meds first", payload.NewData.ProducedMedicineName))
		return
	}

	prodDate, err := utils.ParseDate(payload.NewData.ProductionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	err = h.productionStore.ModifyProduction(payload.ID, types.Production{
		Number:               payload.NewData.Number,
		ProducedMedicineID:   producedMedicine.ID,
		ProducedQty:          payload.NewData.ProducedQty,
		ProducedUnitID:       newProducedUnit.ID,
		ProductionDate:       *prodDate,
		Description:          payload.NewData.Description,
		UpdatedToStock:       payload.NewData.UpdatedToStock,
		UpdatedToAccount:     payload.NewData.UpdatedToAccount,
		TotalCost:            payload.NewData.TotalCost,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// add to stock
	if payload.NewData.UpdatedToStock {
		err = addStock(h, newProducedMedicine, newProducedUnit, float64(payload.NewData.ProducedQty), user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}
	}

	// get production
	production, err := h.productionStore.GetProductionByNumber(payload.NewData.Number)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("production number %d doesn't exists", payload.NewData.Number))
		return
	}

	err = h.productionStore.DeleteProductionMedicineItems(oldProduction, user)
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
			ProductionID: production.ID,
			MedicineID:   medData.ID,
			Qty:          medicine.Qty,
			UnitID:       unit.ID,
			Cost:         medicine.Cost,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("production number %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("production modified by %s", user.Name))
}

func addStock(h *Handler, medData *types.Medicine, producedUnit *types.Unit, additionalQty float64, user *types.User) error {
	var updatedQty float64

	if medData.FirstUnitID == producedUnit.ID {
		updatedQty = (additionalQty + medData.Qty)
	} else if medData.SecondUnitID == producedUnit.ID {
		updatedQty = (additionalQty * medData.SecondUnitToFirstUnitRatio) + medData.Qty
	} else if medData.ThirdUnitID == producedUnit.ID {
		updatedQty = (additionalQty * medData.ThirdUnitToFirstUnitRatio) + medData.Qty
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	err := h.medStore.UpdateMedicineStock(medData.ID, updatedQty, user)
	if err != nil {
		return err
	}

	return nil
}

func subtractStock(h *Handler, medData *types.Medicine, producedUnit *types.Unit, subtractionQty float64, user *types.User) error {
	var updatedQty float64

	if medData.FirstUnitID == producedUnit.ID {
		updatedQty = (medData.Qty - subtractionQty)
	} else if medData.SecondUnitID == producedUnit.ID {
		updatedQty = medData.Qty - (subtractionQty * medData.SecondUnitToFirstUnitRatio)
	} else if medData.ThirdUnitID == producedUnit.ID {
		updatedQty = medData.Qty - (subtractionQty * medData.ThirdUnitToFirstUnitRatio)
	} else {
		return fmt.Errorf("unknown unit name for %s", medData.Name)
	}

	err := h.medStore.UpdateMedicineStock(medData.ID, updatedQty, user)
	if err != nil {
		return err
	}

	return nil
}
