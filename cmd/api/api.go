package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/auth"
	"github.com/nicolaics/pos_pharmacy/service/companyprofile"
	"github.com/nicolaics/pos_pharmacy/service/customer"
	"github.com/nicolaics/pos_pharmacy/service/doctor"
	"github.com/nicolaics/pos_pharmacy/service/invoice"
	"github.com/nicolaics/pos_pharmacy/service/medicine"
	"github.com/nicolaics/pos_pharmacy/service/patient"
	"github.com/nicolaics/pos_pharmacy/service/paymentmethod"
	"github.com/nicolaics/pos_pharmacy/service/poinvoice"
	"github.com/nicolaics/pos_pharmacy/service/prescription"
	"github.com/nicolaics/pos_pharmacy/service/production"
	"github.com/nicolaics/pos_pharmacy/service/purchaseinvoice"
	"github.com/nicolaics/pos_pharmacy/service/supplier"
	"github.com/nicolaics/pos_pharmacy/service/unit"
	"github.com/nicolaics/pos_pharmacy/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	loggerVar := log.New(os.Stdout, "", log.LstdFlags)

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	subrouterUnprotected := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	customerStore := customer.NewStore(s.db)
	supplierStore := supplier.NewStore(s.db)
	medicineStore := medicine.NewStore(s.db)
	companyProfileStore := companyprofile.NewStore(s.db)
	doctorStore := doctor.NewStore(s.db)
	patientStore := patient.NewStore(s.db)

	paymentMethodStore := paymentmethod.NewStore(s.db)
	unitStore := unit.NewStore(s.db)

	purchaseInvoiceStore := purchaseinvoice.NewStore(s.db)
	poInvoiceStore := poinvoice.NewStore(s.db)
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

	companyProfileHandler := companyprofile.NewHandler(companyProfileStore, userStore)
	companyProfileHandler.RegisterRoutes(subrouter)

	doctorHandler := doctor.NewHandler(doctorStore, userStore)
	doctorHandler.RegisterRoutes(subrouter)

	patientHandler := patient.NewHandler(patientStore, userStore)
	patientHandler.RegisterRoutes(subrouter)

	purchaseInvoiceHandler := purchaseinvoice.NewHandler(purchaseInvoiceStore, userStore, supplierStore,
		companyProfileStore, medicineStore, unitStore, poInvoiceStore)
	purchaseInvoiceHandler.RegisterRoutes(subrouter)

	poInvoiceHandler := poinvoice.NewHandler(poInvoiceStore, userStore, supplierStore, companyProfileStore,
		medicineStore, unitStore)
	poInvoiceHandler.RegisterRoutes(subrouter)

	invoiceHandler := invoice.NewHandler(invoiceStore, userStore, customerStore,
		paymentMethodStore, medicineStore, unitStore)
	invoiceHandler.RegisterRoutes(subrouter)

	prescriptionHandler := prescription.NewHandler(prescriptionStore, userStore, customerStore,
													medicineStore, unitStore, invoiceStore,
													doctorStore, patientStore)
	prescriptionHandler.RegisterRoutes(subrouter)

	productionHandler := production.NewHandler(productionStore, userStore, medicineStore, unitStore)
	productionHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	router.Use(logMiddleware.Func())

	router.Use(auth.CorsMiddleware())
	subrouter.Use(auth.AuthMiddleware())
	

	return http.ListenAndServe(s.addr, router)
}
