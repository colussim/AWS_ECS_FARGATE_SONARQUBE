package main


import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {

	// Retrieve environment variables
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbName := os.Getenv("DATABASE_NAME")
	dbUsername := os.Getenv("DATABASE_USERNAME")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	passSonar := os.Getenv("PASS_SONAR")
	Index := os.Getenv("DATABASE_PARTNER")

	// Create a PostgreSQL connection string
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=require",
		dbHost, dbPort, dbName, dbUsername, dbPassword)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Read the SQL script from a file : create sonarqube Database and Role
	scriptPath := "sql/script.sql"
	sqlScript, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Fatal(err)
	}
	query := string(sqlScript)
	query = strings.ReplaceAll(query, "?", Index)
	query = strings.Replace(query, "'PASSWD'", fmt.Sprintf("'%s'", passSonar), -1)

	Message := fmt.Sprintf("Create Database sonarqube_part%s and Role sonarqube_%s", Index, Index)
	fmt.Println(Message)
	// Split the SQL script into individual statements
	statements := strings.Split(string(query), ";")
	// Execute each SQL statement
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			log.Println("Error executing SQL statement:", err)
		}
	}

	db.Close()

}
