package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nicebooks/controller"
)

var DB *sql.DB

func init() {
	// obtenemos una conexion a la BD
	DB = controller.NewConnection()
}

func main() {
	// router es un objeto que nos permite servir ficheros tanto estáticos como dinámicos
	router := gin.Default()
	// se preprocesan los archivos HTMLX (son archivos HTML pero con snippets de codigo Golang)
	router.LoadHTMLGlob("templates/*")
	// se sirve lo que haya en el directorio para poder tener acceso a tales elementos
	router.Static("/static", "./static")

	// servimos el index
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// servimos la página de registro
	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	// comprobamos que la cookie asignada por el endpoint del login es válida y en ese caso carga la página del dashboard
	router.GET("/dashboard", func(c *gin.Context) {
		if controller.CheckCookies(c, DB) {
			c.HTML(http.StatusOK, "dashboard.html", nil)
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	// registro de usuarios
	router.POST("/register-endpoint", func(c *gin.Context) {
		controller.Register(c, DB)
	})

	// comprobación de credenciales de login
	router.POST("/login-endpoint", func(c *gin.Context) {
		controller.Login(c, DB)
	})

	// comprobamos que no se producen errores y si se producen, se mostrarán por pantalla
	err := router.Run()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
