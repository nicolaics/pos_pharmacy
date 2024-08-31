package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/cashier"
	"github.com/nicolaics/pos_pharmacy/service/companyprofile"
	"github.com/nicolaics/pos_pharmacy/service/customer"
	"github.com/nicolaics/pos_pharmacy/service/medicine"
	"github.com/nicolaics/pos_pharmacy/service/poinvoice"
	"github.com/nicolaics/pos_pharmacy/service/purchaseinvoice"

	// "github.com/nicolaics/pos_pharmacy/service/paymentmethod"
	"github.com/nicolaics/pos_pharmacy/service/supplier"
	"github.com/nicolaics/pos_pharmacy/service/unit"
	"github.com/redis/go-redis/v9"
)

type APIServer struct {
	addr        string
	db          *sql.DB
	redisClient *redis.Client
}

func NewAPIServer(addr string, db *sql.DB, redisClient *redis.Client) *APIServer {
	return &APIServer{
		addr:        addr,
		db:          db,
		redisClient: redisClient,
	}
}

func (s *APIServer) Run() error {
	loggerVar := log.New(os.Stdout, "", log.LstdFlags)

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	// router.HandleFunc("/verify", auth.AuthenticationMiddleware(auth.AuthHandler))

	cashierStore := cashier.NewStore(s.db, s.redisClient)
	customerStore := customer.NewStore(s.db)
	supplierStore := supplier.NewStore(s.db)
	medicineStore := medicine.NewStore(s.db)
	companyProfileStore := companyprofile.NewStore(s.db)

	// paymentMethodStore := paymentmethod.NewStore(s.db)
	unitStore := unit.NewStore(s.db)

	purchaseInvoiceStore := purchaseinvoice.NewStore(s.db)
	poInvoiceStore := poinvoice.NewStore(s.db)

	cashierHandler := cashier.NewHandler(cashierStore)
	cashierHandler.RegisterRoutes(subrouter)

	customerHandler := customer.NewHandler(customerStore, cashierStore)
	customerHandler.RegisterRoutes(subrouter)

	supplierHandler := supplier.NewHandler(supplierStore, cashierStore)
	supplierHandler.RegisterRoutes(subrouter)

	medicineHandler := medicine.NewHandler(medicineStore, cashierStore, unitStore)
	medicineHandler.RegisterRoutes(subrouter)

	companyProfileHandler := companyprofile.NewHandler(companyProfileStore, cashierStore)
	companyProfileHandler.RegisterRoutes(subrouter)

	purchaseInvoiceHandler := purchaseinvoice.NewHandler(purchaseInvoiceStore, cashierStore, supplierStore, 
														companyProfileStore, medicineStore, unitStore)
	purchaseInvoiceHandler.RegisterRoutes(subrouter)

	poInvoiceHandler := poinvoice.NewHandler(poInvoiceStore, cashierStore, supplierStore, companyProfileStore,
											medicineStore, unitStore)
	poInvoiceHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	router.Use(logMiddleware.Func())

	return http.ListenAndServe(s.addr, router)
}
