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

	// Read environment variables
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	//user := os.Getenv("USER")
	dbName := os.Getenv("DB_NAME")
	password := os.Getenv("PASSWORD")

	// Print environment variables for debugging
	fmt.Println("HOST:", host)
	fmt.Println("PORT:", port)
	//fmt.Println("USER:", user)
	fmt.Println("DB_NAME:", dbName)
	fmt.Println("PASSWORD:", password)

	// Set up postgres and open it
	postgresSetup := fmt.Sprintf("host=%s port=%s user=postgres dbname=%s password=%s sslmode=disable", host, port, dbName, password)
	fmt.Println("Connection String:", postgresSetup) // Debugging line

	db, errSql := sql.Open("postgres", postgresSetup)
	if errSql != nil {
		fmt.Println("There was an error trying to connect to the database:", errSql)
		panic(errSql)
	} else {
		Db = db
		fmt.Println("Successfully connected to the database")
	}
}
