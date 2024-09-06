package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/nicolaics/pos_pharmacy/logger"
	"github.com/nicolaics/pos_pharmacy/service/companyprofile"
	"github.com/nicolaics/pos_pharmacy/service/customer"
	"github.com/nicolaics/pos_pharmacy/service/doctor"
	"github.com/nicolaics/pos_pharmacy/service/invoice"
	"github.com/nicolaics/pos_pharmacy/service/medicine"
	"github.com/nicolaics/pos_pharmacy/service/patient"
	"github.com/nicolaics/pos_pharmacy/service/poinvoice"
	"github.com/nicolaics/pos_pharmacy/service/prescription"
	"github.com/nicolaics/pos_pharmacy/service/production"
	"github.com/nicolaics/pos_pharmacy/service/purchaseinvoice"
	"github.com/nicolaics/pos_pharmacy/service/user"
	"github.com/nicolaics/pos_pharmacy/service/paymentmethod"
	"github.com/nicolaics/pos_pharmacy/service/supplier"
	"github.com/nicolaics/pos_pharmacy/service/unit"
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
		companyProfileStore, medicineStore, unitStore)
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

	return http.ListenAndServe(s.addr, router)
}

// TODO: VERIFY FOR SESSION MIDDLEWARE, so everytime they access page, check
var (
	store = sessions.NewCookieStore([]byte("test"))
)

func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo admin 로그인에 대해서 static 폴더에 대한 파일은 예외처리

		// 세션 검사
		session, _ := store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// 다음 핸들러 또는 미들웨어로 요청 전달
		next.ServeHTTP(w, r)
	})
}

func checkSession(r *http.Request) bool {
	session, err := store.Get(r, "session-name")
	if err != nil {
		// 세션 가져오기 실패
		return false
	}

	// 세션에 'authenticated' 값이 있는지 확인
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}
	return true
}

// TODO: login successfully, set the authentification to true
// TODO: check as well from the JWT token authorized
// if adminID == "jelly" && password == "walk" {
// 	session, _ := store.Get(r, "session-name")
// 	session.Values["authenticated"] = true
// 	session.Save(r, w)

// 	http.Redirect(w, r, "/admin/", http.StatusFound)
// 	return
// }

// TODO: if failed redirect
// http.Redirect(w, r, "/admin/login", http.StatusFound)

// TODO: after logout, set the authentification to false
// func handleLogout(w http.ResponseWriter, r *http.Request) {

// 	session, _ := store.Get(r, "session-name")
// 	session.Values["authenticated"] = false
// 	session.Save(r, w)

// 	http.Redirect(w, r, "/admin-login", http.StatusFound)
// }
