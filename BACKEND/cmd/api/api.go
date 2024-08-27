package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/cashier"
)

type APIServer struct {
	addr string
	db *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db: db,
	}
}

func (s *APIServer) Run() error {
	loggerVar:= log.New(os.Stdout, "", log.LstdFlags)
	
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	cashierStore := cashier.NewStore(s.db)
	cashierHandler := cashier.NewHandler(cashierStore)
	cashierHandler.RegisterRoutes(subrouter)
	
	// productStore := product.NewStore(s.db)
	// productHandler := product.NewHandler(productStore)
	// productHandler.RegisterRoutes(subrouter)

	// orderStore := order.NewStore(s.db)

	// cartHandler := cart.NewHandler(orderStore, productStore, userStore)
	// cartHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
    router.Use(logMiddleware.Func())

	return http.ListenAndServe(s.addr, router)
}