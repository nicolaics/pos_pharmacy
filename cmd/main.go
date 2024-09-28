package main

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/nicolaics/pos_pharmacy/cmd/api"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/db"
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

	server := api.NewAPIServer((":" + config.Envs.Port), db)

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
// func setupStaticRoutes() {
	
// }
