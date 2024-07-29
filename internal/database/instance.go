package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Instance struct {
	db *sql.DB
}

func (database Instance) Initialize() (err error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName)

	database.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = database.db.Ping()
	if err == nil {
		fmt.Println("Successfully connected to the database!")
	}

	return err
}

func (database Instance) Close() {
	database.db.Close()
}
