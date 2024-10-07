package prescription

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	// "github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type Handler struct {
	prescriptionStore         types.PrescriptionStore
	userStore                 types.UserStore
	customerStore             types.CustomerStore
	medStore                  types.MedicineStore
	unitStore                 types.UnitStore
	invoiceStore              types.InvoiceStore
	doctorStore               types.DoctorStore
	patientStore              types.PatientStore
	consumeTimeStore          types.ConsumeTimeStore
	detStore                  types.DetStore
	doseStore                 types.DoseStore
	mfStore                   types.MfStore
	prescriptionSetUsageStore types.PrescriptionSetUsageStore
}

func NewHandler(prescriptionStore types.PrescriptionStore,
	userStore types.UserStore,
	customerStore types.CustomerStore,
	medStore types.MedicineStore,
	unitStore types.UnitStore,
	invoiceStore types.InvoiceStore,
	doctorStore types.DoctorStore,
	patientStore types.PatientStore,
	consumeTimeStore types.ConsumeTimeStore,
	detStore types.DetStore,
	doseStore types.DoseStore,
	mfStore types.MfStore,
	prescriptionSetUsageStore types.PrescriptionSetUsageStore) *Handler {
	return &Handler{
		prescriptionStore:         prescriptionStore,
		userStore:                 userStore,
		customerStore:             customerStore,
		medStore:                  medStore,
		unitStore:                 unitStore,
		invoiceStore:              invoiceStore,
		doctorStore:               doctorStore,
		patientStore:              patientStore,
		consumeTimeStore:          consumeTimeStore,
		detStore:                  detStore,
		doseStore:                 doseStore,
		mfStore:                   mfStore,
		prescriptionSetUsageStore: prescriptionSetUsageStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/prescription", h.handleRegister).Methods(http.MethodPost)

	// TODO: add more get prescriptions
	router.HandleFunc("/prescription/{params}/{val}", h.handleGetPrescriptions).Methods(http.MethodPost)

	router.HandleFunc("/prescription/detail", h.handleGetPrescriptionDetail).Methods(http.MethodPost)
	router.HandleFunc("/prescription", h.handleDelete).Methods(http.MethodDelete)
	router.HandleFunc("/prescription", h.handleModify).Methods(http.MethodPatch)
	router.HandleFunc("/prescription/print", h.handlePrint).Methods(http.MethodPost)

	router.HandleFunc("/prescription", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/prescription/{params}/{val}", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/prescription/detail", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
	router.HandleFunc("/prescription/print", func(w http.ResponseWriter, r *http.Request) { utils.WriteJSONForOptions(w, http.StatusOK, nil) }).Methods(http.MethodOptions)
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

	// get customerID info from invoice
	invoiceCustomer, err := h.customerStore.GetCustomerByName(payload.Invoice.CustomerName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not found", payload.Invoice.CustomerName))
		return
	}

	invoiceDate, err := utils.ParseDate(payload.Invoice.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	// get invoice data
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Invoice.Number, invoiceCustomer.ID, *invoiceDate)
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

	patient, err := h.patientStore.GetPatientByName(payload.PatientName, payload.PatientAge)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.PatientName,
			Age:  payload.PatientAge,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err))
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.PatientName, payload.PatientAge)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.PatientName))
		return
	}

	prescriptionDate, err := utils.ParseDate(payload.PrescriptionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	// check duplicate
	today := time.Now().Format("2006-01-02 -0700MST")
	startDate, err := utils.ParseStartDate(today)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse start date: %v", err))
		return
	}
	endDate, err := utils.ParseEndDate(today)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse end date: %v", err))
		return
	}

	isValid, err := h.prescriptionStore.IsValidPrescriptionNumber(payload.Number, *startDate, *endDate)
	if err != nil || !isValid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription %d exists", payload.Number))
		return
	}

	presc := types.Prescription{
		InvoiceID:            invoiceId,
		Number:               payload.Number,
		PrescriptionDate:     *prescriptionDate,
		PatientID:            patient.ID,
		DoctorID:             doctor.ID,
		Qty:                  payload.Qty,
		Price:                payload.Price,
		TotalPrice:           payload.TotalPrice,
		Description:          payload.Description,
		UserID:               user.ID,
		LastModifiedByUserID: user.ID,
		PDFUrl:               "",
	}

	err = h.prescriptionStore.CreatePrescription(presc)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get prescription ID
	prescriptionId, err := h.prescriptionStore.GetPrescriptionID(invoiceId, payload.Number, *prescriptionDate,
		patient.ID, payload.TotalPrice, doctor.ID)
	if err != nil {
		errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
			return
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription %d doesn't exists", payload.Number))
		return
	}

	eticketFileNames := make([]string, 0)
	setNumber := 1

	for _, setItem := range payload.SetItems {
		// get consume time
		consumeTime, err := h.consumeTimeStore.GetConsumeTimeByName(setItem.ConsumeTime)
		if consumeTime == nil {
			err = h.consumeTimeStore.CreateConsumeTime(setItem.ConsumeTime)
			if err == nil {
				consumeTime, err = h.consumeTimeStore.GetConsumeTimeByName(setItem.ConsumeTime)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume time: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get det
		det, err := h.detStore.GetDetByName(setItem.Det)
		if det == nil {
			err = h.detStore.CreateDet(setItem.Det)
			if err == nil {
				det, err = h.detStore.GetDetByName(setItem.Det)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("det: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get dose
		dose, err := h.doseStore.GetDoseByName(setItem.Dose)
		if dose == nil {
			err = h.doseStore.CreateDose(setItem.Dose)
			if err == nil {
				dose, err = h.doseStore.GetDoseByName(setItem.Dose)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("dose: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get mf
		mf, err := h.mfStore.GetMfByName(setItem.Mf)
		if mf == nil {
			err = h.mfStore.CreateMf(setItem.Mf)
			if err == nil {
				mf, err = h.mfStore.GetMfByName(setItem.Mf)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("mf: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get set usage
		setUsage, err := h.prescriptionSetUsageStore.GetPrescriptionSetUsageByName(setItem.Usage)
		if setUsage == nil {
			err = h.prescriptionSetUsageStore.CreatePrescriptionSetUsage(setItem.Usage)
			if err == nil {
				setUsage, err = h.prescriptionSetUsageStore.GetPrescriptionSetUsageByName(setItem.Usage)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("set usage: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get consume unit
		setUnit, err := h.unitStore.GetUnitByName(setItem.SetUnit)
		if setUnit == nil {
			err = h.unitStore.CreateUnit(setItem.SetUnit)
			if err == nil {
				setUnit, err = h.unitStore.GetUnitByName(setItem.SetUnit)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume unit: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		setItemStore := types.PrescriptionSetItem{
			PrescriptionID: prescriptionId,
			MfID:           mf.ID,
			DoseID:         dose.ID,
			SetUnitID:      setUnit.ID,
			ConsumeTimeID:  consumeTime.ID,
			DetID:          det.ID,
			UsageID:        setUsage.ID,
			MustFinish:     setItem.MustFinish,
			PrintEticket:   setItem.PrintEticket,
		}
		err = h.prescriptionStore.CreateSetItem(setItemStore)
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine set: %v", err))
			return
		}

		// get medicine set store data
		setItemStoreId, err := h.prescriptionStore.GetSetItemID(setItemStore)
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine set id"))
			return
		}

		// create eticket
		if setItem.PrintEticket {
			eticket := types.Eticket{
				PrescriptionID:        prescriptionId,
				PrescriptionSetItemID: setItemStoreId,
				Number:                setItem.Eticket.Number,
				MedicineQty:           setItem.Eticket.MedicineQty,
				PDFUrl:                "",
			}

			err = h.prescriptionStore.CreateEticket(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			eticketId, err := h.prescriptionStore.GetEticketID(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			err = h.prescriptionStore.UpdateEticketID(eticketId, setItemStoreId)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				_ = h.prescriptionStore.DeleteEticket(eticketId)
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			eticketPDF := types.EticketPDFReturnPayload{
				Number:      setItem.Eticket.Number,
				PatientName: patient.Name,
				SetUsage:    setUsage.Name,
				Dose:        dose.Name,
				SetUnit:     setUnit.Name,
				ConsumeTime: consumeTime.Name,
				MustFinish:  setItem.MustFinish,
				MedicineQty: setItem.Eticket.MedicineQty,
			}
			eticketFileName, err := utils.CreateEticket7x4PDF(eticketPDF, setNumber, h.prescriptionStore)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err))
				return
			}

			eticketFileNames = append(eticketFileNames, eticketFileName)
			setNumber++

			err = h.prescriptionStore.UpdatePDFUrl("eticket", eticketId, eticketFileName)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update eticket pdf url: %v", err))
				return
			}
		}

		for _, medicine := range setItem.MedicineLists {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicine.Unit)
			if unit == nil {
				err = h.unitStore.CreateUnit(medicine.Unit)
				if err == nil {
					unit, err = h.unitStore.GetUnitByName(medicine.Unit)
				}
			}
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			fractionIdx := strings.Index(medicine.Qty, "/")
			var medicineQty float64
			if fractionIdx == -1 {
				medicineQty, err = strconv.ParseFloat(medicine.Qty, 64)
				if err != nil {
					errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
					if errDel != nil {
						utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
						return
					}

					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err))
					return
				}
			} else {
				fraction := strings.Split(medicine.Qty, "/")
				numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
				denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

				medicineQty = numerator / denum
			}
			medicineItem := types.PrescriptionMedicineItem{
				PrescriptionSetItemID: setItemStoreId,
				MedicineID:            medData.ID,
				Qty:                   medicineQty,
				UnitID:                unit.ID,
				Price:                 medicine.Price,
				Discount:              medicine.Discount,
				Subtotal:              medicine.Subtotal,
			}
			err = h.prescriptionStore.CreatePrescriptionMedicineItem(medicineItem)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError,
					fmt.Errorf("prescription %d, med %s: %v", payload.Number, medicine.MedicineName, err))
				return
			}

			err = utils.CheckStock(medData, unit, medicineQty)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough, need %.2f: %v", medicine.MedicineName, medicineQty, err))
				return
			}
		}
	}

	medicineSets, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescriptionId)
	if err != nil {
		errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine items: %v", err))
		return
	}

	prescPDF := types.PrescriptionPDFReturn{
		Number:       payload.Number,
		Date:         *prescriptionDate,
		Patient:      *patient,
		Doctor:       *doctor,
		MedicineSets: medicineSets,
	}
	prescFileName, err := utils.CreatePrescriptionPDF(prescPDF, h.prescriptionStore, "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create presc pdf: %v", err))
		return
	}

	err = h.prescriptionStore.UpdatePDFUrl("prescription", prescriptionId, prescFileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update presc pdf url: %v", err))
		return
	}

	// subtract the stock
	for _, setItem := range medicineSets {
		for _, medicine := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicine.Unit)
			if unit == nil {
				err = h.unitStore.CreateUnit(medicine.Unit)
				if err == nil {
					unit, err = h.unitStore.GetUnitByName(medicine.Unit)
				}
			}
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			err = utils.SubtractStock(h.medStore, medData, unit, medicine.QtyFloat, user)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
				return
			}
		}
	}

	returnPayload := map[string]interface{}{
		"success":         fmt.Sprintf("prescription %d successfully created by %s", payload.Number, user.Name),
		"prescriptionPDF": prescFileName,
		"eticketPDF":      eticketFileNames,
	}
	utils.WriteJSON(w, http.StatusCreated, returnPayload)
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

	var prescriptions []types.PrescriptionListsReturnPayload

	if val == "all" {
		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDate(*startDate, *endDate)
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

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndNumber(*startDate, *endDate, number)
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
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndUserID(*startDate, *endDate, user.ID)
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
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndPatientID(*startDate, *endDate, patient.ID)
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
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndDoctorID(*startDate, *endDate, doctor.ID)
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

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndInvoiceID(*startDate, *endDate, iid)
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

// view one prescription
func (h *Handler) handleGetPrescriptionDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionDetailPayload

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

	// get the details of set items and medicine items of the prescription
	items, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
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
		PDFUrl:                 prescription.PDFUrl,

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
			Age  int    "json:\"age\""
		}{
			ID:   patient.ID,
			Name: patient.Name,
			Age:  patient.Age,
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

		MedicineSets: items,
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

	// get set items
	setItems, err := h.prescriptionStore.GetSetItemsByPrescriptionID(prescription.ID)
	if setItems == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("set items of presc id %d doesn't exist", payload.ID))
		return
	}

	medicineItems := make([]types.PrescriptionMedicineItemReturn, 0)

	for _, setItem := range setItems {
		medicineItem, err := h.prescriptionStore.GetPrescriptionMedicineItems(setItem.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding medicine item: %v", err))
			return
		}

		medicineItems = append(medicineItems, medicineItem...)

		err = h.prescriptionStore.DeletePrescriptionMedicineItem(prescription, setItem.ID, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = h.prescriptionStore.DeleteSetItem(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.prescriptionStore.DeletePrescription(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	for _, medicineItem := range medicineItems {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName))
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.QtyFloat, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
			return
		}
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

	// get customerID info from invoice
	invoiceCustomer, err := h.customerStore.GetCustomerByName(payload.NewData.Invoice.CustomerName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not found", payload.NewData.Invoice.CustomerName))
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.Invoice.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	// get invoice data
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.NewData.Invoice.Number, invoiceCustomer.ID, *invoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d not found", payload.NewData.Invoice.Number))
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

	patient, err := h.patientStore.GetPatientByName(payload.NewData.PatientName, payload.NewData.PatientAge)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.NewData.PatientName,
			Age:  payload.NewData.PatientAge,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err))
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.NewData.PatientName, payload.NewData.PatientAge)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.NewData.PatientName))
		return
	}

	prescriptionDate, err := utils.ParseDate(payload.NewData.PrescriptionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"))
		return
	}

	oldPrescriptionSetItems, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding prescription items: %v", err))
		return
	}

	newPresc := types.Prescription{
		InvoiceID:            invoiceId,
		Number:               payload.NewData.Number,
		PrescriptionDate:     *prescriptionDate,
		PatientID:            patient.ID,
		DoctorID:             doctor.ID,
		Qty:                  payload.NewData.Qty,
		Price:                payload.NewData.Price,
		TotalPrice:           payload.NewData.TotalPrice,
		Description:          payload.NewData.Description,
		LastModifiedByUserID: user.ID,
	}
	err = h.prescriptionStore.ModifyPrescription(payload.ID, newPresc, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	eticketFileNames := make([]string, 0)
	setNumber := 1

	// delete set items
	for _, setItem := range oldPrescriptionSetItems {
		err = h.prescriptionStore.DeletePrescriptionMedicineItem(prescription, setItem.ID, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = h.prescriptionStore.DeleteSetItem(prescription, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting set item: %v", err))
			return
		}

		for _, medicineItem := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName))
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			err = utils.AddStock(h.medStore, medData, unit, medicineItem.QtyFloat, user)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
				return
			}
		}
	}

	// tODO: remove absolute delete
	// create new set items
	for _, setItem := range payload.NewData.SetItems {
		// get consume time
		consumeTime, err := h.consumeTimeStore.GetConsumeTimeByName(setItem.ConsumeTime)
		if consumeTime == nil {
			err = h.consumeTimeStore.CreateConsumeTime(setItem.ConsumeTime)
			if err == nil {
				consumeTime, err = h.consumeTimeStore.GetConsumeTimeByName(setItem.ConsumeTime)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume time: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get det
		det, err := h.detStore.GetDetByName(setItem.Det)
		if det == nil {
			err = h.detStore.CreateDet(setItem.Det)
			if err == nil {
				det, err = h.detStore.GetDetByName(setItem.Det)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("det: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get dose
		dose, err := h.doseStore.GetDoseByName(setItem.Dose)
		if dose == nil {
			err = h.doseStore.CreateDose(setItem.Dose)
			if err == nil {
				dose, err = h.doseStore.GetDoseByName(setItem.Dose)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("dose: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get mf
		mf, err := h.mfStore.GetMfByName(setItem.Mf)
		if mf == nil {
			err = h.mfStore.CreateMf(setItem.Mf)
			if err == nil {
				mf, err = h.mfStore.GetMfByName(setItem.Mf)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("mf: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get set usage
		setUsage, err := h.prescriptionSetUsageStore.GetPrescriptionSetUsageByName(setItem.Usage)
		if setUsage == nil {
			err = h.prescriptionSetUsageStore.CreatePrescriptionSetUsage(setItem.Usage)
			if err == nil {
				setUsage, err = h.prescriptionSetUsageStore.GetPrescriptionSetUsageByName(setItem.Usage)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("set usage: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get consume unit
		setUnit, err := h.unitStore.GetUnitByName(setItem.SetUnit)
		if setUnit == nil {
			err = h.unitStore.CreateUnit(setItem.SetUnit)
			if err == nil {
				setUnit, err = h.unitStore.GetUnitByName(setItem.SetUnit)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume unit: error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		setItemStore := types.PrescriptionSetItem{
			PrescriptionID: payload.ID,
			MfID:           mf.ID,
			DoseID:         dose.ID,
			SetUnitID:      setUnit.ID,
			ConsumeTimeID:  consumeTime.ID,
			DetID:          det.ID,
			UsageID:        setUsage.ID,
			MustFinish:     setItem.MustFinish,
			PrintEticket:   setItem.PrintEticket,
		}
		err = h.prescriptionStore.CreateSetItem(setItemStore)
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine set: %v", err))
			return
		}

		// get medicine set store data
		setItemStoreId, err := h.prescriptionStore.GetSetItemID(setItemStore)
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
			if errDel != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine set id"))
			return
		}

		// create eticket
		if setItem.PrintEticket {
			err = h.prescriptionStore.DeleteEticketByPrescriptionID(payload.ID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting eticket: %v", err))
				return
			}

			eticket := types.Eticket{
				PrescriptionID:        payload.ID,
				PrescriptionSetItemID: setItemStoreId,
				Number:                setItem.Eticket.Number,
				MedicineQty:           setItem.Eticket.MedicineQty,
				PDFUrl:                "",
			}
			err = h.prescriptionStore.CreateEticket(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			eticketId, err := h.prescriptionStore.GetEticketID(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			err = h.prescriptionStore.UpdateEticketID(eticketId, setItemStoreId)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				_ = h.prescriptionStore.DeleteEticket(eticketId)
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err))
				return
			}

			eticketPDF := types.EticketPDFReturnPayload{
				Number:      setItem.Eticket.Number,
				PatientName: patient.Name,
				SetUsage:    setUsage.Name,
				Dose:        dose.Name,
				SetUnit:     setUnit.Name,
				ConsumeTime: consumeTime.Name,
				MustFinish:  setItem.MustFinish,
				MedicineQty: setItem.Eticket.MedicineQty,
			}
			eticketFileName, err := utils.CreateEticket7x4PDF(eticketPDF, setNumber, h.prescriptionStore)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err))
				return
			}

			eticketFileNames = append(eticketFileNames, eticketFileName)
			setNumber++
		}

		for _, medicine := range setItem.MedicineLists {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicine.Unit)
			if unit == nil {
				err = h.unitStore.CreateUnit(medicine.Unit)
				if err == nil {
					unit, err = h.unitStore.GetUnitByName(medicine.Unit)
				}
			}
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			fractionIdx := strings.Index(medicine.Qty, "/")
			var medicineQty float64
			if fractionIdx == -1 {
				medicineQty, err = strconv.ParseFloat(medicine.Qty, 64)
				if err != nil {
					errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
					if errDel != nil {
						utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
						return
					}

					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err))
					return
				}
			} else {
				fraction := strings.Split(medicine.Qty, "/")
				numerator, _ := strconv.ParseFloat(strings.TrimSpace(fraction[0]), 64)
				denum, _ := strconv.ParseFloat(strings.TrimSpace(fraction[1]), 64)

				medicineQty = numerator / denum
			}
			medicineItem := types.PrescriptionMedicineItem{
				PrescriptionSetItemID: setItemStoreId,
				MedicineID:            medData.ID,
				Qty:                   medicineQty,
				UnitID:                unit.ID,
				Price:                 medicine.Price,
				Discount:              medicine.Discount,
				Subtotal:              medicine.Subtotal,
			}
			err = h.prescriptionStore.CreatePrescriptionMedicineItem(medicineItem)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError,
					fmt.Errorf("prescription %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err))
				return
			}

			err = utils.CheckStock(medData, unit, medicineQty)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough", medicine.MedicineName))
				return
			}
		}
	}

	medicineSets, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
	if err != nil {
		errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine items: %v", err))
		return
	}

	prescPDF := types.PrescriptionPDFReturn{
		Number:       payload.NewData.Number,
		Date:         *prescriptionDate,
		Patient:      *patient,
		Doctor:       *doctor,
		MedicineSets: medicineSets,
	}
	prescFileName, err := utils.CreatePrescriptionPDF(prescPDF, h.prescriptionStore, prescription.PDFUrl)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create presc pdf: %v", err))
		return
	}

	// subtract the stock
	for _, setItem := range medicineSets {
		for _, medicine := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName))
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicine.Unit)
			if unit == nil {
				err = h.unitStore.CreateUnit(medicine.Unit)
				if err == nil {
					unit, err = h.unitStore.GetUnitByName(medicine.Unit)
				}
			}
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			err = utils.SubtractStock(h.medStore, medData, unit, medicine.QtyFloat, user)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(newPresc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel))
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err))
				return
			}
		}
	}

	returnPayload := map[string]interface{}{
		"success":         fmt.Sprintf("prescription modified by %s", user.Name),
		"prescriptionPDF": prescFileName,
		"eticketPDF":      eticketFileNames,
	}
	utils.WriteJSON(w, http.StatusOK, returnPayload)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionDetailPayload

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

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.PrescriptionID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription with id %d doesn't exists", payload.PrescriptionID))
		return
	}

	pdfFiles := []string{("static/pdf/prescription/" + prescription.PDFUrl)}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=pdfFiles.zip")
	w.WriteHeader(http.StatusOK)

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	etickets, err := h.prescriptionStore.GetEticketsByPrescriptionID(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(etickets) > 0 {
		for _, eticket := range etickets {
			pdfFiles = append(pdfFiles, ("static/pdf/eticket/" + eticket.PDFUrl))
		}
	}

	for _, fileName := range pdfFiles {
		file, err := os.Open(fileName)
		if err != nil {
			http.Error(w, "File not found: "+fileName, http.StatusNotFound)
			return
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(filepath.Base(fileName))
		if err != nil {
			http.Error(w, "Could not create zip", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			http.Error(w, "Could not copy file content", http.StatusInternalServerError)
			return
		}
	}
}
