package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/nicolaics/pos_pharmacy/cmd/api"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/db"
	"github.com/redis/go-redis/v9"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Envs.RedisDSN,
	})

	initStorage(db, redisClient)

	server := api.NewAPIServer((":" + config.Envs.Port), db, redisClient)

	// check the error, if error is not nill
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB, client *redis.Client) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}
