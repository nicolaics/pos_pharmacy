package prescription

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	prescriptionStore types.PrescriptionStore
	userStore         types.UserStore
	customerStore     types.CustomerStore
	medStore          types.MedicineStore
	unitStore         types.UnitStore
	invoiceStore      types.InvoiceStore
	doctorStore       types.DoctorStore
	patientStore      types.PatientStore
}

func NewHandler(prescriptionStore types.PrescriptionStore,
	userStore types.UserStore,
	customerStore types.CustomerStore,
	medStore types.MedicineStore,
	unitStore types.UnitStore,
	invoiceStore types.InvoiceStore,
	doctorStore types.DoctorStore,
	patientStore types.PatientStore) *Handler {
	return &Handler{
		prescriptionStore: prescriptionStore,
		userStore:         userStore,
		customerStore:     customerStore,
		medStore:          medStore,
		unitStore:         unitStore,
		invoiceStore:      invoiceStore,
		doctorStore:       doctorStore,
		patientStore:      patientStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/prescription", h.handleRegister).Methods(http.MethodPost)

	// TODO: add more get prescriptions
	router.HandleFunc("/prescription/{params}/{val}", h.handleGetPrescriptions).Methods(http.MethodPost)

	router.HandleFunc("/prescription/detail", h.handleGetPrescriptionDetail).Methods(http.MethodPost)
	router.HandleFunc("/prescription", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/prescription", h.handleModify).Methods(http.MethodPatch)

	// router.HandleFunc("/prescription", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	// router.HandleFunc("/prescription/all/date", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	// router.HandleFunc("/prescription/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.RegisterPrescriptionPayload

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

	// get userID info from invoice
	invoiceUser, err := h.userStore.GetUserByName(payload.Invoice.UserName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not found", payload.Invoice.UserName))
		return
	}

	// get customerID info from invoice
	invoiceCustomer, err := h.customerStore.GetCustomerByName(payload.Invoice.CustomerName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not found", payload.Invoice.CustomerName))
		return
	}

	// get invoice data
	invoiceId, err := h.invoiceStore.GetInvoiceID(
		payload.Invoice.Number, invoiceUser.ID, invoiceCustomer.ID,
		payload.Invoice.TotalPrice, payload.Invoice.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d not found", payload.Invoice.Number))
		return
	}

	doctor, err := h.doctorStore.GetDoctorByName(payload.DoctorName)
	if doctor == nil {
		err = h.doctorStore.CreateDoctor(types.Doctor{
			Name: payload.DoctorName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating doctor: %v", err))
			return
		}

		doctor, err = h.doctorStore.GetDoctorByName(payload.DoctorName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.DoctorName))
		return
	}

	patient, err := h.patientStore.GetPatientByName(payload.PatientName)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.PatientName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err))
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.PatientName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.DoctorName))
		return
	}

	err = h.prescriptionStore.CreatePrescription(types.Prescription{
		InvoiceID:            invoiceId,
		Number:               payload.Number,
		PrescriptionDate:     payload.PrescriptionDate,
		PatientID:            patient.ID,
		DoctorID:             doctor.ID,
		Qty:                  payload.Qty,
		Price:                payload.Price,
		TotalPrice:           payload.TotalPrice,
		Description:          payload.Description,
		UserID:               user.ID,
		LastModifiedByUserID: user.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get prescription ID
	prescriptionId, err := h.prescriptionStore.GetPrescriptionID(invoiceId, payload.Number, payload.PrescriptionDate,
		payload.PatientName, payload.TotalPrice)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription %d doesn't exists", payload.Number))
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

		err = h.prescriptionStore.CreatePrescriptionMedicineItems(types.PrescriptionMedicineItems{
			PrescriptionID: prescriptionId,
			MedicineID:     medData.ID,
			Qty:            medicine.Qty,
			UnitID:         unit.ID,
			Price:          medicine.Price,
			Discount:       medicine.Discount,
			Subtotal:       medicine.Subtotal,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("prescription %d, med %s: %v", payload.Number, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("prescription %d successfully created by %s", payload.Number, user.Name))
}

// only view the prescription list
func (h *Handler) handleGetPrescriptions(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionsPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
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

	var prescriptions []types.PrescriptionListsReturnPayload

	if val == "all" {
		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDate(payload.StartDate, payload.EndDate)
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

		prescription, err := h.prescriptionStore.GetPrescriptionByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription id %d not exist", id))
			return
		}

		invoice, err := h.invoiceStore.GetInvoiceByID(prescription.InvoiceID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invoice id %d not found", prescription.InvoiceID))
			return
		}

		patient, err := h.patientStore.GetPatientByID(prescription.PatientID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("patient id %d not found", prescription.PatientID))
			return
		}

		doctor, err := h.doctorStore.GetDoctorByID(prescription.DoctorID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("doctor id %d not found", prescription.DoctorID))
			return
		}

		user, err := h.userStore.GetUserByID(prescription.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user id %d not found", prescription.UserID))
			return
		}

		customer, err := h.customerStore.GetCustomerByID(invoice.CustomerID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("customer id %d not found", invoice.CustomerID))
			return
		}

		prescriptions = append(prescriptions, types.PrescriptionListsReturnPayload{
			ID:               prescription.ID,
			Number:           prescription.Number,
			PrescriptionDate: prescription.PrescriptionDate,
			PatientName:      patient.Name,
			DoctorName:       doctor.Name,
			Qty:              prescription.Qty,
			Price:            prescription.Price,
			TotalPrice:       prescription.TotalPrice,
			Description:      prescription.Description,
			UserName:         user.Name,
			Invoice: struct {
				Number       int       "json:\"number\""
				CustomerName string    "json:\"customerName\""
				TotalPrice   float64   "json:\"totalPrice\""
				InvoiceDate  time.Time "json:\"invoiceDate\""
			}{
				Number:       invoice.Number,
				CustomerName: customer.Name,
				TotalPrice:   invoice.TotalPrice,
				InvoiceDate:  invoice.InvoiceDate,
			},
		})
	} else if params == "number" {
		number, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndNumber(payload.StartDate, payload.EndDate, number)
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
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndUserID(payload.StartDate, payload.EndDate, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s doesn't create any prescription between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "patient" {
		patients, err := h.patientStore.GetPatientsBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient %s not exists", val))
			return
		}

		for _, patient := range patients {
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndPatientID(payload.StartDate, payload.EndDate, patient.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient %s doesn't have any prescription between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "doctor" {
		doctors, err := h.doctorStore.GetDoctorsBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor %s not exists", val))
			return
		}

		for _, doctor := range doctors {
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndDoctorID(payload.StartDate, payload.EndDate, doctor.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor %s doesn't have any prescription between %s and %s", val, payload.StartDate, payload.EndDate))
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "invoice-id" {
		iid, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndInvoiceID(payload.StartDate, payload.EndDate, iid)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, prescriptions)
}

// only view the purchase invoice list
func (h *Handler) handleGetPrescriptionDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionMedicineItemsPayload

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

	// get prescription data
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.PrescriptionID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription id %d doesn't exists", payload.PrescriptionID))
		return
	}

	// get medicine items of the prescription
	prescriptionItems, err := h.prescriptionStore.GetPrescriptionMedicineItems(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get user data, the one who inputs the prescription
	inputter, err := h.userStore.GetUserByID(prescription.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", prescription.UserID))
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(prescription.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", prescription.LastModifiedByUserID))
		return
	}

	doctor, err := h.doctorStore.GetDoctorByID(prescription.DoctorID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor id %d not found", prescription.DoctorID))
		return
	}

	patient, err := h.patientStore.GetPatientByID(prescription.PatientID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient id %d not found", prescription.PatientID))
		return
	}

	invoice, err := h.invoiceStore.GetInvoiceByID(prescription.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d not found", prescription.InvoiceID))
		return
	}

	customer, err := h.customerStore.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d not found", invoice.CustomerID))
		return
	}

	returnPayload := types.PrescriptionDetailPayload{
		ID:                     prescription.ID,
		Number:                 prescription.Number,
		PrescriptionDate:       prescription.PrescriptionDate,
		Qty:                    prescription.Qty,
		Price:                  prescription.Price,
		TotalPrice:             prescription.TotalPrice,
		Description:            prescription.Description,
		CreatedAt:              prescription.CreatedAt,
		LastModified:           prescription.LastModified,
		LastModifiedByUserName: lastModifiedUser.Name,

		Invoice: struct {
			Number       int       "json:\"number\""
			CustomerName string    "json:\"customerName\""
			TotalPrice   float64   "json:\"totalPrice\""
			InvoiceDate  time.Time "json:\"invoiceDate\""
		}{
			Number:       invoice.Number,
			CustomerName: customer.Name,
			TotalPrice:   invoice.TotalPrice,
			InvoiceDate:  invoice.InvoiceDate,
		},

		Patient: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   patient.ID,
			Name: patient.Name,
		},

		Doctor: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   doctor.ID,
			Name: doctor.Name,
		},

		User: struct {
			ID   int    "json:\"id\""
			Name string "json:\"name\""
		}{
			ID:   inputter.ID,
			Name: inputter.Name,
		},

		MedicineLists: prescriptionItems,
	}

	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePrescription

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

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.ID)
	if prescription == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription id %d doesn't exist", payload.ID))
		return
	}

	err = h.prescriptionStore.DeletePrescriptionMedicineItems(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.prescriptionStore.DeletePrescription(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("prescription number %d deleted by %s", prescription.Number, user.Name))
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPrescriptionPayload

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

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription with id %d doesn't exists", payload.ID))
		return
	}

	doctor, err := h.doctorStore.GetDoctorByName(payload.NewData.DoctorName)
	if doctor == nil {
		err = h.doctorStore.CreateDoctor(types.Doctor{
			Name: payload.NewData.DoctorName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating doctor: %v", err))
			return
		}

		doctor, err = h.doctorStore.GetDoctorByName(payload.NewData.DoctorName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.NewData.DoctorName))
		return
	}

	patient, err := h.patientStore.GetPatientByName(payload.NewData.PatientName)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.NewData.PatientName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err))
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.NewData.PatientName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.NewData.DoctorName))
		return
	}

	err = h.prescriptionStore.ModifyPrescription(payload.ID, types.Prescription{
		Number:               payload.NewData.Number,
		PrescriptionDate:     payload.NewData.PrescriptionDate,
		PatientID:            patient.ID,
		DoctorID:             doctor.ID,
		Qty:                  payload.NewData.Qty,
		Price:                payload.NewData.Price,
		TotalPrice:           payload.NewData.TotalPrice,
		Description:          payload.NewData.Description,
		LastModifiedByUserID: user.ID,
	}, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.prescriptionStore.DeletePrescriptionMedicineItems(prescription, user)
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

		err = h.prescriptionStore.CreatePrescriptionMedicineItems(types.PrescriptionMedicineItems{
			PrescriptionID: payload.ID,
			MedicineID:     medData.ID,
			Qty:            medicine.Qty,
			UnitID:         unit.ID,
			Price:          medicine.Price,
			Discount:       medicine.Discount,
			Subtotal:       medicine.Subtotal,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("prescription id %d, med %s: %v", payload.ID, medicine.MedicineName, err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, fmt.Sprintf("prescription modified by %s", user.Name))
}
