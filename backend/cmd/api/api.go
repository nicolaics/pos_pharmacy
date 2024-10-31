package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/auth"
	"github.com/nicolaics/pos_pharmacy/service/consumetime"
	"github.com/nicolaics/pos_pharmacy/service/customer"
	"github.com/nicolaics/pos_pharmacy/service/det"
	"github.com/nicolaics/pos_pharmacy/service/doctor"
	"github.com/nicolaics/pos_pharmacy/service/dose"
	"github.com/nicolaics/pos_pharmacy/service/invoice"
	"github.com/nicolaics/pos_pharmacy/service/maindoctorprescmeditem"
	"github.com/nicolaics/pos_pharmacy/service/medicine"
	"github.com/nicolaics/pos_pharmacy/service/mf"
	"github.com/nicolaics/pos_pharmacy/service/patient"
	"github.com/nicolaics/pos_pharmacy/service/paymentmethod"
	"github.com/nicolaics/pos_pharmacy/service/prescription"
	"github.com/nicolaics/pos_pharmacy/service/prescriptionsetusage"
	"github.com/nicolaics/pos_pharmacy/service/production"
	"github.com/nicolaics/pos_pharmacy/service/purchaseinvoice"
	"github.com/nicolaics/pos_pharmacy/service/purchaseorder"
	"github.com/nicolaics/pos_pharmacy/service/supplier"
	"github.com/nicolaics/pos_pharmacy/service/unit"
	"github.com/nicolaics/pos_pharmacy/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
	router *mux.Router
}

func NewAPIServer(addr string, db *sql.DB, router *mux.Router) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
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
	consumeTimeStore := consumetime.NewStore(s.db)
	detStore := det.NewStore(s.db)
	doseStore := dose.NewStore(s.db)
	mfStore := mf.NewStore(s.db)
	prescSetUsageStore := prescriptionsetusage.NewStore(s.db)

	mainDoctorPrescMedItemStore := maindoctorprescmeditem.NewStore(s.db)

	paymentMethodStore := paymentmethod.NewStore(s.db)
	unitStore := unit.NewStore(s.db)

	purchaseInvoiceStore := purchaseinvoice.NewStore(s.db)
	poInvoiceStore := purchaseorder.NewStore(s.db)
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

	purchaseInvoiceHandler := purchaseinvoice.NewHandler(purchaseInvoiceStore, userStore, supplierStore, medicineStore, unitStore, poInvoiceStore)
	purchaseInvoiceHandler.RegisterRoutes(subrouter)

	poInvoiceHandler := purchaseorder.NewHandler(poInvoiceStore, userStore, supplierStore,
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

	mainDoctorPrescMedItemHandler := maindoctorprescmeditem.NewHandler(mainDoctorPrescMedItemStore, userStore, medicineStore, unitStore)
	mainDoctorPrescMedItemHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	s.router.Use(logMiddleware.Func())

	s.router.Use(auth.CorsMiddleware())
	subrouter.Use(auth.AuthMiddleware())

	return http.ListenAndServe(s.addr, s.router)
}
