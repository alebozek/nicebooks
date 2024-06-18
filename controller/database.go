package controller

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
	"log"
	"os"
)

// Devuelve la cadena de conexión a la base de datos acorde con las credenciales del fichero .env
func loadCredentials() string {
	err := godotenv.Load("creds.env")
	if err != nil {
		log.Fatalln("Error loading credentials file.")
	}

	var (
		server   = os.Getenv("DB_SERVER")
		port     = 1433
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASS")
		database = os.Getenv("DB_NAME")
	)

	dbstring := fmt.Sprintf("server=%s; port=%d; user id=%s; database=%s; password=%s; sslmode=require", server, port, user, database, password)
	fmt.Println(dbstring)

	return dbstring
}

// Devuelve un puntero a la conexion de la base de datos para que podamos realizar operaciones en ella
// Recoge las credenciales por medio de una variable de entorno llamada DATABASE_URL
func NewConnection() *sql.DB {
	// declaracion especifica de tipo para obviar errores
	var DB *sql.DB
	var err error
	DB, err = sql.Open("sqlserver", loadCredentials())
	if err != nil {
		log.Fatalln(err.Error())
	}

	return DB
}
