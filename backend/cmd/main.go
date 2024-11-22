package main

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/nicolaics/pharmacon/cmd/api"
	"github.com/nicolaics/pharmacon/config"
	"github.com/nicolaics/pharmacon/db"
)

func main() {
	db, err := db.NewMySQLStorage(mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)
	router := mux.NewRouter()

	server := api.NewAPIServer((":" + config.Envs.Port), db, router)

	// check the error, if error is not nill
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}

// TODO: add static routes if the frontend is in the server already
// func setupStaticRoutes(r *mux.Router) {
// 	fs := http.FileServer(http.Dir("./static"))
// 	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", noDirListing(fs)))

// 	// profile image and payment proof file server
// 	profileImgServer := http.FileServer(http.Dir("./profile_img"))
// 	r.PathPrefix("/profile_img/").Handler(http.StripPrefix("/profile_img/", noDirListing(profileImgServer)))

// 	paymentProofServer := http.FileServer(http.Dir("./payment_proof"))
// 	http.Handle("/payment_proof/", http.StripPrefix("/payment_proof", noDirListing(paymentProofServer)))
// }

// func noDirListing(h http.Handler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// URL이 디렉토리를 가리키는 경우 (마지막이 /로 끝나는 경우) 404 오류를 반환
// 		if strings.HasSuffix(r.URL.Path, "/") {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		h.ServeHTTP(w, r)
// 	}
// }
