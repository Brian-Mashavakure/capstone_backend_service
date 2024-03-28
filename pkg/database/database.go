package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // don't forget to add it. It doesn't be added automatically
	"os"
)

// Global db variable
var Db *sql.DB

func DatabaseConnect() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error occurred on .env file please check")
	}

	//read env file
	host := os.Getenv("Host")
	port := os.Getenv("PORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("DB_NAME")
	password := os.Getenv("PASSWORD")

	//set up postgres and open it
	postgresSetup := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbName, password)
	db, errSql := sql.Open("postgres", postgresSetup)
	if errSql != nil {
		fmt.Println("There was an error trying to connect to the database", err)
		panic(err)
	} else {
		Db = db
		fmt.Println("Successfully connected to the database")
	}
}
