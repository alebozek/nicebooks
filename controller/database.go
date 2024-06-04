package controller

import (
	"database/sql"
	"fmt"
	_ "github.com/microsoft/go-mssqldb"
	"log"
)

// Devuelve un puntero a la conexion de la base de datos para que podamos realizar operaciones en ella
// Recoge las credenciales por medio de una variable de entorno llamada DATABASE_URL
func NewConnection() *sql.DB {
	const (
		server   = "nicebooks.database.windows.net"
		port     = 1433
		user     = "alebozek"
		password = "cynq4HQcdKhatAeUMkLO"
		database = "nicebooks"
	)

	dbstring := fmt.Sprintf("server=%s; port=%d; user id=%s; database=%s; password=%s; sslmode=require", server, port, user, database, password)
	// declaracion especifica de tipo para obviar errores
	var DB *sql.DB
	var err error
	DB, err = sql.Open("sqlserver", dbstring)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return DB
}
