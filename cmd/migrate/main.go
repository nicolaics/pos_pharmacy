package main

import (
	"os"

	"github.com/golang-migrate/migrate/v4"

	_ "database/sql"
	"log"

	mySqlConfig "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/db"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := db.NewMySQLStorage(mySqlConfig.Config{
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

	driver, err := mysql.WithInstance(db, &mysql.Config{})

	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"mysql",
		driver,
	)

	if err != nil {
		log.Fatal(err)
	}

	// revert using Down
	cmd := os.Args[(len(os.Args) - 1)]

	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}

	if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
}