package models

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectToDB() {
	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	db := os.Getenv("DB_NAME")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbInfo := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, db, username, password)
	var err error
	DB, err = sql.Open("postgres", dbInfo)
	if err != nil {
		panic(err)
	}
}
