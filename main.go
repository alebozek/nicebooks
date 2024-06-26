package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nicebooks/controller"
	"nicebooks/models"
	"strconv"
)

var db *sql.DB

func init() {
	// obtenemos una conexion a la BD
	db = controller.NewConnection()
}

func main() {
	// router es un objeto que nos permite servir ficheros tanto estáticos como dinámicos
	router := gin.Default()
	// se preprocesan los archivos HTMLX (son archivos HTML pero con snippets de codigo Golang)
	router.LoadHTMLGlob("templates/*")
	// se sirve lo que haya en el directorio para poder tener acceso a tales elementos
	router.Static("/static", "./static")
	// sirve el favicon
	router.StaticFile("favicon.ico", "./static/favicon.ico")

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
		if controller.CheckCookies(c, db) {
			token, _ := c.Cookie("token")
			books := controller.GetBooksReadByUser(db, token)
			c.HTML(http.StatusOK, "dashboard.html", struct {
				BookList []models.Book
				Err      error
			}{books, nil})
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	// sirve la página de añadir libros
	router.GET("/add-book", func(c *gin.Context) {
		if controller.CheckCookies(c, db) {
			c.HTML(http.StatusOK, "addbook.html", nil)
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	// gestiona el formulario de añadir libros
	router.POST("/add-book", func(c *gin.Context) {
		if controller.CheckCookies(c, db) {
			controller.AddBook(c, db)
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	// registro de usuarios
	router.POST("/register", func(c *gin.Context) {
		controller.Register(c, db)
	})

	// comprobación de credenciales de login
	router.POST("/login", func(c *gin.Context) {
		controller.Login(c, db)
	})

	// añade libros que el usuario ha leído
	router.POST("/add-read", func(c *gin.Context) {
		controller.AddRead(c, db)
	})

	// borra lecturas
	router.POST("/delete-read", func(c *gin.Context) {
		if controller.CheckCookies(c, db) {
			token, _ := c.Cookie("token")
			controller.DeleteRead(c, db, token)
		} else {
			c.HTML(http.StatusSeeOther, "index.html", gin.H{"error": "Invalid credentials"})
		}
	})

	// endpoint para modificar libros y su rating
	router.GET("/edit-read/:id", func(c *gin.Context) {
		if controller.CheckCookies(c, db) {
			token, _ := c.Cookie("token")
			idForm := c.Param("id")
			id, err := strconv.Atoi(idForm)
			if err != nil {
				c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid id."})
			}
			book, err := controller.GetBookByRead(db, token, id)
			if err != nil {
				c.HTML(http.StatusSeeOther, "editbook.html", gin.H{"book": book, "error": err})
			}
			c.HTML(http.StatusSeeOther, "editbook.html", gin.H{"book": book, "error": nil})
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	router.POST("/edit-read", func(c *gin.Context) {
		if controller.CheckCookies(c, db) {
			token, _ := c.Cookie("token")
			controller.EditRead(c, db, token)
		} else {
			c.HTML(http.StatusSeeOther, "index.html", gin.H{"error": "Invalid credentials."})
		}
	})

	// comprobamos que no se producen errores y si se producen, se mostrarán por pantalla
	err := router.Run()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
