package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/cashier"
	"github.com/nicolaics/pos_pharmacy/service/customer"
	"github.com/redis/go-redis/v9"
)

type APIServer struct {
	addr string
	db *sql.DB
	redisClient *redis.Client
}

func NewAPIServer(addr string, db *sql.DB, redisClient *redis.Client) *APIServer {
	return &APIServer{
		addr: addr,
		db: db,
		redisClient: redisClient,
	}
}

func (s *APIServer) Run() error {
	loggerVar:= log.New(os.Stdout, "", log.LstdFlags)
	
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	cashierStore := cashier.NewStore(s.db, s.redisClient)
	customerStore := customer.NewStore(s.db)

	cashierHandler := cashier.NewHandler(cashierStore)
	cashierHandler.RegisterRoutes(subrouter)

	customerHandler := customer.NewHandler(customerStore, cashierStore)
	customerHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
    router.Use(logMiddleware.Func())

	return http.ListenAndServe(s.addr, router)
}