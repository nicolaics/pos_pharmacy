package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/db"
	"github.com/nicolaics/pos_pharmacy/service/auth"
	"github.com/nicolaics/pos_pharmacy/utils"
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	row := db.QueryRow("SELECT COUNT(*) FROM user")
	if row.Err() != nil {
		log.Fatal(row.Err())
	}

	var cnt int
	err = row.Scan(&cnt)
	if err != nil {
		log.Fatal(err)
	}

	if cnt != 0 {
		log.Fatal("initial admin already exist!")
	}

	password := utils.GenerateRandomCodeAlphanumeric(12)

	// create new admin
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("failed to hashed password: %v", err)
		return
	}

	args := os.Args

	query := `INSERT INTO user (
		name, password, admin, phone_number
		) VALUES (?, ?, ?, ?)`

	_, err = db.Exec(query, args[1], hashedPassword, true, "000")
	if err != nil {
		log.Fatal(err)
	}

	fh, err := os.OpenFile("admin.txt", os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	msg := fmt.Sprintf("Username: %s\nPassword: %s", args[1], password)
	_, err = fh.WriteString(msg)
	if err != nil {
		log.Fatal(err)
	}
}
