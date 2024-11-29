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

	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/types"
	"github.com/nicolaics/pharmacon/utils"
	"github.com/nicolaics/pharmacon/utils/pdf"
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
	consumeTimeStore  types.ConsumeTimeStore
	detStore          types.DetStore
	doseStore         types.DoseStore
	mfStore           types.MfStore
	usageStore        types.UsageStore
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
	usageStore types.UsageStore) *Handler {
	return &Handler{
		prescriptionStore: prescriptionStore,
		userStore:         userStore,
		customerStore:     customerStore,
		medStore:          medStore,
		unitStore:         unitStore,
		invoiceStore:      invoiceStore,
		doctorStore:       doctorStore,
		patientStore:      patientStore,
		consumeTimeStore:  consumeTimeStore,
		detStore:          detStore,
		doseStore:         doseStore,
		mfStore:           mfStore,
		usageStore:        usageStore,
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

	// get customerID info from invoice
	invoiceCustomer, err := h.customerStore.GetCustomerByName(payload.Invoice.CustomerName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not found", payload.Invoice.CustomerName), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.Invoice.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	// get invoice data
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.Invoice.Number, invoiceCustomer.ID, *invoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d not found", payload.Invoice.Number), nil)
		return
	}

	doctor, err := h.doctorStore.GetDoctorByName(payload.DoctorName)
	if doctor == nil {
		err = h.doctorStore.CreateDoctor(types.Doctor{
			Name: payload.DoctorName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating doctor: %v", err), nil)
			return
		}

		doctor, err = h.doctorStore.GetDoctorByName(payload.DoctorName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.DoctorName), nil)
		return
	}

	patient, err := h.patientStore.GetPatientByName(payload.PatientName, payload.PatientAge)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.PatientName,
			Age:  payload.PatientAge,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err), nil)
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.PatientName, payload.PatientAge)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.PatientName), nil)
		return
	}

	prescriptionDate, err := utils.ParseDate(payload.PrescriptionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	// check duplicate
	today := time.Now().Format("2006-01-02 -0700MST")
	startDate, err := utils.ParseStartDate(today)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse start date: %v", err), nil)
		return
	}
	endDate, err := utils.ParseEndDate(today)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to parse end date: %v", err), nil)
		return
	}

	isValid, err := h.prescriptionStore.IsValidPrescriptionNumber(payload.Number, *startDate, *endDate)
	if err != nil || !isValid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription %d exists", payload.Number), nil)
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
		PdfUrl:               "",
	}

	err = h.prescriptionStore.CreatePrescription(presc)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get prescription ID
	prescriptionId, err := h.prescriptionStore.GetPrescriptionID(invoiceId, payload.Number, *prescriptionDate,
		patient.ID, payload.TotalPrice, doctor.ID)
	if err != nil {
		errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
			return
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription %d doesn't exists", payload.Number), nil)
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
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume time: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("det: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("dose: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("mf: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		// get set usage
		setUsage, err := h.usageStore.GetUsageByName(setItem.Usage)
		if setUsage == nil {
			err = h.usageStore.CreateUsage(setItem.Usage)
			if err == nil {
				setUsage, err = h.usageStore.GetUsageByName(setItem.Usage)
			}
		}
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("set usage: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("consume unit: error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine set: %v", err), nil)
			return
		}

		// get medicine set store data
		setItemStoreId, err := h.prescriptionStore.GetSetItemID(setItemStore)
		if err != nil {
			errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
			if errDel != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
				return
			}

			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine set id"), nil)
			return
		}

		// create eticket
		if setItem.PrintEticket {
			eticket := types.Eticket{
				PrescriptionID:        prescriptionId,
				PrescriptionSetItemID: setItemStoreId,
				Number:                setItem.Eticket.Number,
				MedicineQty:           setItem.Eticket.MedicineQty,
				Size:                  setItem.Eticket.Size,
				PdfUrl:                "",
			}

			err = h.prescriptionStore.CreateEticket(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err), nil)
				return
			}

			eticketId, err := h.prescriptionStore.GetEticketID(eticket)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err), nil)
				return
			}

			err = h.prescriptionStore.UpdateEticketID(eticketId, setItemStoreId)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				_ = h.prescriptionStore.DeleteEticket(eticketId)
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err), nil)
				return
			}

			eticketPdf := types.EticketPdfPayload{
				Number:      setItem.Eticket.Number,
				PatientName: patient.Name,
				Usage:       setUsage.Name,
				Dose:        dose.Name,
				SetUnit:     setUnit.Name,
				ConsumeTime: consumeTime.Name,
				MustFinish:  setItem.MustFinish,
				MedicineQty: setItem.Eticket.MedicineQty,
			}

			var eticketFileName string
			if setItem.Eticket.Size == "7x4" {
				eticketFileName, err = pdf.CreateEticket7x4Pdf(eticketPdf, setNumber, h.prescriptionStore)
				if err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err), nil)
					return
				}
			} else if setItem.Eticket.Size == "7x5" {
				eticketFileName, err = pdf.CreateEticket7x5Pdf(eticketPdf, setNumber, h.prescriptionStore)
				if err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err), nil)
					return
				}
			} else {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown eticket size: %s", setItem.Eticket.Size), nil)
				return
			}

			eticketFileNames = append(eticketFileNames, eticketFileName)
			setNumber++

			err = h.prescriptionStore.UpdatePdfUrl("eticket", eticketId, eticketFileName)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update eticket pdf url: %v", err), nil)
				return
			}
		}

		for _, medicine := range setItem.MedicineLists {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
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
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			fractionIdx := strings.Index(medicine.Qty, "/")
			var medicineQty float64
			if fractionIdx == -1 {
				medicineQty, err = strconv.ParseFloat(medicine.Qty, 64)
				if err != nil {
					errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
					if errDel != nil {
						utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
						return
					}

					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err), nil)
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
				DiscountPercentage:    medicine.DiscountPercentage,
				DiscountAmount:        medicine.DiscountAmount,
				Subtotal:              medicine.Subtotal,
			}
			err = h.prescriptionStore.CreatePrescriptionMedicineItem(medicineItem)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError,
					fmt.Errorf("prescription %d, med %s: %v", payload.Number, medicine.MedicineName, err), nil)
				return
			}

			err = utils.CheckStock(medData, unit, medicineQty)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough, need %.2f: %v", medicine.MedicineName, medicineQty, err), nil)
				return
			}
		}
	}

	medicineSets, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescriptionId)
	if err != nil {
		errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
		if errDel != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine items: %v", err), nil)
		return
	}

	prescPdf := types.PrescriptionPdfPayload{
		Number:       payload.Number,
		Date:         *prescriptionDate,
		Patient:      *patient,
		Doctor:       *doctor,
		MedicineSets: medicineSets,
	}
	prescFileName, err := pdf.CreatePrescriptionPdf(prescPdf, h.prescriptionStore, "")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create presc pdf: %v", err), nil)
		return
	}

	err = h.prescriptionStore.UpdatePdfUrl("prescription", prescriptionId, prescFileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update presc pdf url: %v", err), nil)
		return
	}

	logFileNames := make([]string, 0)

	if payload.PrintExtra {
		err = pdf.CreatePrescriptionExtraPdf(prescPdf, prescFileName)
		if err != nil {
			logFileName, _ := logger.WriteServerErrorLog(fmt.Errorf("error create extra pdf: %v", err))
			logFileNames = append(logFileNames, logFileName)
		} else {
			err = h.prescriptionStore.UpdatePrintExtraPdf(prescriptionId)
			if err != nil {
				logFileName, _ := logger.WriteServerErrorLog(fmt.Errorf("error update print extra pdf: %v", err))
				logFileNames = append(logFileNames, logFileName)
			}
		}
	}

	// subtract the stock
	for _, setItem := range medicineSets {
		for _, medicine := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
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
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			err = utils.SubtractStock(h.medStore, medData, unit, medicine.QtyFloat, user)
			if err != nil {
				errDel := h.prescriptionStore.AbsoluteDeletePrescription(presc)
				if errDel != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error absolute delete prescription: %v", errDel), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
				return
			}
		}
	}

	returnPayload := map[string]interface{}{
		"success":         fmt.Sprintf("prescription %d successfully created by %s", payload.Number, user.Name),
		"prescriptionPdf": prescFileName,
		"eticketPdf":      eticketFileNames,
	}
	utils.WriteSuccess(w, http.StatusCreated, returnPayload, logFileNames)
}

// only view the prescription list
func (h *Handler) handleGetPrescriptions(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionsPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors), nil)
	}

	// validate token
	_, err := h.userStore.ValidateUserToken(w, r, false)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid: %v", err), nil)
		return
	}

	vars := mux.Vars(r)
	params := vars["params"]
	val := vars["val"]

	startDate, err := utils.ParseStartDate(payload.StartDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	endDate, err := utils.ParseEndDate(payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	var prescriptions []types.PrescriptionListsReturnPayload

	if val == "all" {
		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDate(*startDate, *endDate)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else if params == "id" {
		id, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		prescription, err := h.prescriptionStore.GetPrescriptionByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription id %d not exist", id), nil)
			return
		}

		invoice, err := h.invoiceStore.GetInvoiceByID(prescription.InvoiceID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("invoice id %d not found", prescription.InvoiceID), nil)
			return
		}

		patient, err := h.patientStore.GetPatientByID(prescription.PatientID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("patient id %d not found", prescription.PatientID), nil)
			return
		}

		doctor, err := h.doctorStore.GetDoctorByID(prescription.DoctorID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("doctor id %d not found", prescription.DoctorID), nil)
			return
		}

		user, err := h.userStore.GetUserByID(prescription.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("user id %d not found", prescription.UserID), nil)
			return
		}

		customer, err := h.customerStore.GetCustomerByID(invoice.CustomerID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("customer id %d not found", invoice.CustomerID), nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndNumber(*startDate, *endDate, number)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else if params == "user" {
		users, err := h.userStore.GetUserBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %s not exists", val), nil)
			return
		}

		for _, user := range users {
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndUserID(*startDate, *endDate, user.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest,
					fmt.Errorf("user %s doesn't create any prescription between %s and %s", val, payload.StartDate, payload.EndDate), nil)
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "patient" {
		patients, err := h.patientStore.GetPatientsBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient %s not exists", val), nil)
			return
		}

		for _, patient := range patients {
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndPatientID(*startDate, *endDate, patient.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest,
					fmt.Errorf("patient %s doesn't have any prescription between %s and %s", val, payload.StartDate, payload.EndDate), nil)
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "doctor" {
		doctors, err := h.doctorStore.GetDoctorsBySearchName(val)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor %s not exists", val), nil)
			return
		}

		for _, doctor := range doctors {
			temp, err := h.prescriptionStore.GetPrescriptionsByDateAndDoctorID(*startDate, *endDate, doctor.ID)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest,
					fmt.Errorf("doctor %s doesn't have any prescription between %s and %s", val, payload.StartDate, payload.EndDate), nil)
				return
			}

			prescriptions = append(prescriptions, temp...)
		}
	} else if params == "invoice-id" {
		iid, err := strconv.Atoi(val)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		prescriptions, err = h.prescriptionStore.GetPrescriptionsByDateAndInvoiceID(*startDate, *endDate, iid)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	} else {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("params undefined"), nil)
		return
	}

	utils.WriteSuccess(w, http.StatusOK, prescriptions, nil)
}

// view one prescription
func (h *Handler) handleGetPrescriptionDetail(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionDetailPayload

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

	// get prescription data
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.PrescriptionID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("prescription id %d doesn't exists", payload.PrescriptionID), nil)
		return
	}

	// get the details of set items and medicine items of the prescription
	items, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	// get user data, the one who inputs the prescription
	inputter, err := h.userStore.GetUserByID(prescription.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", prescription.UserID), nil)
		return
	}

	// get last modified user data
	lastModifiedUser, err := h.userStore.GetUserByID(prescription.LastModifiedByUserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user id %d doesn't exists", prescription.LastModifiedByUserID), nil)
		return
	}

	doctor, err := h.doctorStore.GetDoctorByID(prescription.DoctorID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("doctor id %d not found", prescription.DoctorID), nil)
		return
	}

	patient, err := h.patientStore.GetPatientByID(prescription.PatientID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("patient id %d not found", prescription.PatientID), nil)
		return
	}

	invoice, err := h.invoiceStore.GetInvoiceByID(prescription.InvoiceID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice id %d not found", prescription.InvoiceID), nil)
		return
	}

	customer, err := h.customerStore.GetCustomerByID(invoice.CustomerID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer id %d not found", invoice.CustomerID), nil)
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
		PdfUrl:                 prescription.PdfUrl,

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

	utils.WriteSuccess(w, http.StatusOK, returnPayload, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.DeletePrescription

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
	user, err := h.userStore.ValidateUserToken(w, r, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user token invalid or not admin: %v", err), nil)
		return
	}

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.ID)
	if prescription == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription id %d doesn't exist", payload.ID), nil)
		return
	}

	// get set items
	setItems, err := h.prescriptionStore.GetSetItemsByPrescriptionID(prescription.ID)
	if setItems == nil || err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("set items of presc id %d doesn't exist", payload.ID), nil)
		return
	}

	medicineItems := make([]types.PrescriptionMedicineItemReturn, 0)

	for _, setItem := range setItems {
		medicineItem, err := h.prescriptionStore.GetPrescriptionMedicineItems(setItem.ID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding medicine item: %v", err), nil)
			return
		}

		medicineItems = append(medicineItems, medicineItem...)

		err = h.prescriptionStore.DeletePrescriptionMedicineItem(prescription, setItem.ID, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}
	}

	err = h.prescriptionStore.DeleteSetItem(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	err = h.prescriptionStore.DeletePrescription(prescription, user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	for _, medicineItem := range medicineItems {
		medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName), nil)
			return
		}

		unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = utils.AddStock(h.medStore, medData, unit, medicineItem.QtyFloat, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
			return
		}
	}

	utils.WriteSuccess(w, http.StatusOK, fmt.Sprintf("prescription number %d deleted by %s", prescription.Number, user.Name), nil)
}

func (h *Handler) handleModify(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ModifyPrescriptionPayload

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

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription with id %d doesn't exists", payload.ID), nil)
		return
	}

	// get customerID info from invoice
	invoiceCustomer, err := h.customerStore.GetCustomerByName(payload.NewData.Invoice.CustomerName)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("customer %s not found", payload.NewData.Invoice.CustomerName), nil)
		return
	}

	invoiceDate, err := utils.ParseDate(payload.NewData.Invoice.InvoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	// get invoice data
	invoiceId, err := h.invoiceStore.GetInvoiceID(payload.NewData.Invoice.Number, invoiceCustomer.ID, *invoiceDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invoice number %d not found", payload.NewData.Invoice.Number), nil)
		return
	}

	doctor, err := h.doctorStore.GetDoctorByName(payload.NewData.DoctorName)
	if doctor == nil {
		err = h.doctorStore.CreateDoctor(types.Doctor{
			Name: payload.NewData.DoctorName,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating doctor: %v", err), nil)
			return
		}

		doctor, err = h.doctorStore.GetDoctorByName(payload.NewData.DoctorName)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.NewData.DoctorName), nil)
		return
	}

	patient, err := h.patientStore.GetPatientByName(payload.NewData.PatientName, payload.NewData.PatientAge)
	if patient == nil {
		err = h.patientStore.CreatePatient(types.Patient{
			Name: payload.NewData.PatientName,
			Age:  payload.NewData.PatientAge,
		})
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating patient: %v", err), nil)
			return
		}

		patient, err = h.patientStore.GetPatientByName(payload.NewData.PatientName, payload.NewData.PatientAge)
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s not found", payload.NewData.PatientName), nil)
		return
	}

	prescriptionDate, err := utils.ParseDate(payload.NewData.PrescriptionDate)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error parsing date"), nil)
		return
	}

	oldPrescriptionSetItems, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding prescription items: %v", err), nil)
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
		utils.WriteError(w, http.StatusInternalServerError, err, nil)
		return
	}

	eticketFileNames := make([]string, 0)
	setNumber := 1

	// delete set items
	for _, setItem := range oldPrescriptionSetItems {
		err = h.prescriptionStore.DeletePrescriptionMedicineItem(prescription, setItem.ID, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		err = h.prescriptionStore.DeleteSetItem(prescription, user)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting set item: %v", err), nil)
			return
		}

		for _, medicineItem := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicineItem.MedicineBarcode)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicineItem.MedicineName), nil)
				return
			}

			unit, err := h.unitStore.GetUnitByName(medicineItem.Unit)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			err = utils.AddStock(h.medStore, medData, unit, medicineItem.QtyFloat, user)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
				return
			}
		}
	}

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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
			return
		}

		// get set usage
		setUsage, err := h.usageStore.GetUsageByName(setItem.Usage)
		if setUsage == nil {
			err = h.usageStore.CreateUsage(setItem.Usage)
			if err == nil {
				setUsage, err = h.usageStore.GetUsageByName(setItem.Usage)
			}
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
			utils.WriteError(w, http.StatusInternalServerError, err, nil)
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
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create medicine set: %v", err), nil)
			return
		}

		// get medicine set store data
		setItemStoreId, err := h.prescriptionStore.GetSetItemID(setItemStore)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine set id"), nil)
			return
		}

		// create eticket
		if setItem.PrintEticket {
			err = h.prescriptionStore.DeleteEticketByPrescriptionID(payload.ID)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting eticket: %v", err), nil)
				return
			}

			eticket := types.Eticket{
				PrescriptionID:        payload.ID,
				PrescriptionSetItemID: setItemStoreId,
				Number:                setItem.Eticket.Number,
				MedicineQty:           setItem.Eticket.MedicineQty,
				Size:                  setItem.Eticket.Size,
				PdfUrl:                "",
			}
			err = h.prescriptionStore.CreateEticket(eticket)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err), nil)
				return
			}

			eticketId, err := h.prescriptionStore.GetEticketID(eticket)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create eticket: %v", err), nil)
				return
			}

			err = h.prescriptionStore.UpdateEticketID(eticketId, setItemStoreId)
			if err != nil {
				delErr := h.prescriptionStore.DeleteEticket(eticketId)
				if delErr != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error delete eticket: %v: %v", delErr, err), nil)
					return
				}

				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error update eticket: %v", err), nil)
				return
			}

			eticketPdf := types.EticketPdfPayload{
				Number:      setItem.Eticket.Number,
				PatientName: patient.Name,
				Usage:       setUsage.Name,
				Dose:        dose.Name,
				SetUnit:     setUnit.Name,
				ConsumeTime: consumeTime.Name,
				MustFinish:  setItem.MustFinish,
				MedicineQty: setItem.Eticket.MedicineQty,
			}

			var eticketFileName string
			if setItem.Eticket.Size == "7x4" {
				eticketFileName, err = pdf.CreateEticket7x4Pdf(eticketPdf, setNumber, h.prescriptionStore)
				if err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err), nil)
					return
				}
			} else if setItem.Eticket.Size == "7x5" {
				eticketFileName, err = pdf.CreateEticket7x5Pdf(eticketPdf, setNumber, h.prescriptionStore)
				if err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating eticket pdf for number %d: %v", setItem.Eticket.Number, err), nil)
					return
				}
			} else {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown eticket size: %s", setItem.Eticket.Size), nil)
				return
			}

			eticketFileNames = append(eticketFileNames, eticketFileName)
			setNumber++
		}

		for _, medicine := range setItem.MedicineLists {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
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
				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			fractionIdx := strings.Index(medicine.Qty, "/")
			var medicineQty float64
			if fractionIdx == -1 {
				medicineQty, err = strconv.ParseFloat(medicine.Qty, 64)
				if err != nil {
					utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parse float: %v", err), nil)
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
				DiscountPercentage:    medicine.DiscountPercentage,
				DiscountAmount:        medicine.DiscountAmount,
				Subtotal:              medicine.Subtotal,
			}
			err = h.prescriptionStore.CreatePrescriptionMedicineItem(medicineItem)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError,
					fmt.Errorf("prescription %d, med %s: %v", payload.NewData.Number, medicine.MedicineName, err), nil)
				return
			}

			err = utils.CheckStock(medData, unit, medicineQty)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("stock for %s is not enough", medicine.MedicineName), nil)
				return
			}
		}
	}

	medicineSets, err := h.prescriptionStore.GetPrescriptionSetAndMedicineItems(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error get medicine items: %v", err), nil)
		return
	}

	prescPdf := types.PrescriptionPdfPayload{
		Number:       payload.NewData.Number,
		Date:         *prescriptionDate,
		Patient:      *patient,
		Doctor:       *doctor,
		MedicineSets: medicineSets,
	}
	prescFileName, err := pdf.CreatePrescriptionPdf(prescPdf, h.prescriptionStore, prescription.PdfUrl)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error create presc pdf: %v", err), nil)
		return
	}

	logFileNames := make([]string, 0)

	if payload.NewData.PrintExtra {
		err = pdf.CreatePrescriptionExtraPdf(prescPdf, prescFileName)
		if err != nil {
			logFileName, _ := logger.WriteServerErrorLog(fmt.Errorf("error create extra pdf: %v", err))
			logFileNames = append(logFileNames, logFileName)
		} else {
			err = h.prescriptionStore.UpdatePrintExtraPdf(prescription.ID)
			if err != nil {
				logFileName, _ := logger.WriteServerErrorLog(fmt.Errorf("error update print extra pdf: %v", err))
				logFileNames = append(logFileNames, logFileName)
			}
		}
	}

	// subtract the stock
	for _, setItem := range medicineSets {
		for _, medicine := range setItem.MedicineItems {
			medData, err := h.medStore.GetMedicineByBarcode(medicine.MedicineBarcode)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("medicine %s doesn't exists", medicine.MedicineName), nil)
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
				utils.WriteError(w, http.StatusInternalServerError, err, nil)
				return
			}

			err = utils.SubtractStock(h.medStore, medData, unit, medicine.QtyFloat, user)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating stock: %v", err), nil)
				return
			}
		}
	}

	returnPayload := map[string]interface{}{
		"success":         fmt.Sprintf("prescription modified by %s", user.Name),
		"prescriptionPdf": prescFileName,
		"eticketPdf":      eticketFileNames,
	}
	utils.WriteSuccess(w, http.StatusOK, returnPayload, logFileNames)
}

func (h *Handler) handlePrint(w http.ResponseWriter, r *http.Request) {
	// get JSON Payload
	var payload types.ViewPrescriptionDetailPayload

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

	// check if the prescription exists
	prescription, err := h.prescriptionStore.GetPrescriptionByID(payload.PrescriptionID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest,
			fmt.Errorf("prescription with id %d doesn't exists", payload.PrescriptionID), nil)
		return
	}

	pdfFiles := []string{("static/pdf/prescription/" + prescription.PdfUrl)}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=pdfFiles.zip")
	w.WriteHeader(http.StatusOK)

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	if prescription.PrintExtraPdf {
		pdfFiles = append(pdfFiles, ("static/pdf/prescription/extra/e" + prescription.PdfUrl))
	}

	etickets, err := h.prescriptionStore.GetEticketsByPrescriptionID(prescription.ID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, nil)
		return
	}

	if len(etickets) > 0 {
		for _, eticket := range etickets {
			pdfFiles = append(pdfFiles, ("static/pdf/eticket/" + eticket.PdfUrl))
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
