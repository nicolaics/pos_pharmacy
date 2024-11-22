package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/logger"
	"github.com/nicolaics/pharmacon/service/auth"
	"github.com/nicolaics/pharmacon/service/customer"
	"github.com/nicolaics/pharmacon/service/invoice"
	"github.com/nicolaics/pharmacon/service/medicine"
	"github.com/nicolaics/pharmacon/service/payment"
	"github.com/nicolaics/pharmacon/service/pi"
	"github.com/nicolaics/pharmacon/service/poi"
	"github.com/nicolaics/pharmacon/service/prescription"
	"github.com/nicolaics/pharmacon/service/prescription/ct"
	"github.com/nicolaics/pharmacon/service/prescription/det"
	"github.com/nicolaics/pharmacon/service/prescription/doctor"
	"github.com/nicolaics/pharmacon/service/prescription/dose"
	"github.com/nicolaics/pharmacon/service/prescription/mdmi"
	"github.com/nicolaics/pharmacon/service/prescription/mf"
	"github.com/nicolaics/pharmacon/service/prescription/patient"
	"github.com/nicolaics/pharmacon/service/prescription/su"
	"github.com/nicolaics/pharmacon/service/production"
	"github.com/nicolaics/pharmacon/service/supplier"
	"github.com/nicolaics/pharmacon/service/unit"
	"github.com/nicolaics/pharmacon/service/user"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	router *mux.Router
}

func NewAPIServer(addr string, db *sql.DB, router *mux.Router) *APIServer {
	return &APIServer{
		addr:   addr,
		db:     db,
		router: router,
	}
}

func (s *APIServer) Run() error {
	loggerVar := log.New(os.Stdout, "", log.LstdFlags)

	subrouter := s.router.PathPrefix("/api/v1").Subrouter()
	subrouterUnprotected := s.router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	customerStore := customer.NewStore(s.db)
	supplierStore := supplier.NewStore(s.db)
	medicineStore := medicine.NewStore(s.db)
	doctorStore := doctor.NewStore(s.db)
	patientStore := patient.NewStore(s.db)
	consumeTimeStore := ct.NewStore(s.db)
	detStore := det.NewStore(s.db)
	doseStore := dose.NewStore(s.db)
	mfStore := mf.NewStore(s.db)
	prescSetUsageStore := su.NewStore(s.db)

	mainDoctorPrescMedItemStore := mdmi.NewStore(s.db)

	paymentMethodStore := payment.NewStore(s.db)
	unitStore := unit.NewStore(s.db)

	purchaseInvoiceStore := pi.NewStore(s.db)
	poInvoiceStore := poi.NewStore(s.db)
	invoiceStore := invoice.NewStore(s.db)
	prescriptionStore := prescription.NewStore(s.db)
	productionStore := production.NewStore(s.db)

	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)
	userHandler.RegisterUnprotectedRoutes(subrouterUnprotected)

	customerHandler := customer.NewHandler(customerStore, userStore)
	customerHandler.RegisterRoutes(subrouter)

	supplierHandler := supplier.NewHandler(supplierStore, userStore)
	supplierHandler.RegisterRoutes(subrouter)

	medicineHandler := medicine.NewHandler(medicineStore, userStore, unitStore)
	medicineHandler.RegisterRoutes(subrouter)

	doctorHandler := doctor.NewHandler(doctorStore, userStore)
	doctorHandler.RegisterRoutes(subrouter)

	patientHandler := patient.NewHandler(patientStore, userStore)
	patientHandler.RegisterRoutes(subrouter)

	purchaseInvoiceHandler := pi.NewHandler(purchaseInvoiceStore, userStore, supplierStore, medicineStore, unitStore, poInvoiceStore)
	purchaseInvoiceHandler.RegisterRoutes(subrouter)

	poInvoiceHandler := poi.NewHandler(poInvoiceStore, userStore, supplierStore,
		medicineStore, unitStore)
	poInvoiceHandler.RegisterRoutes(subrouter)

	invoiceHandler := invoice.NewHandler(invoiceStore, userStore, customerStore,
		paymentMethodStore, medicineStore, unitStore)
	invoiceHandler.RegisterRoutes(subrouter)

	prescriptionHandler := prescription.NewHandler(prescriptionStore, userStore, customerStore,
		medicineStore, unitStore, invoiceStore,
		doctorStore, patientStore, consumeTimeStore,
		detStore, doseStore, mfStore, prescSetUsageStore)
	prescriptionHandler.RegisterRoutes(subrouter)

	productionHandler := production.NewHandler(productionStore, userStore, medicineStore, unitStore)
	productionHandler.RegisterRoutes(subrouter)

	mainDoctorPrescMedItemHandler := mdmi.NewHandler(mainDoctorPrescMedItemStore, userStore, medicineStore, unitStore)
	mainDoctorPrescMedItemHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	s.router.Use(logMiddleware.Func())

	s.router.Use(auth.CorsMiddleware())
	subrouter.Use(auth.AuthMiddleware())

	return http.ListenAndServe(s.addr, s.router)
}
