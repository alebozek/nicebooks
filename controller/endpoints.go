package controller

import (
	"database/sql"
	"log"
	"net/http"
	"nicebooks/models"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// const DOMAIN = "nicebooks.onrender.com"
const DOMAIN = "localhost"

// Comprueba si los campos proporcionados por el usuario cumplen con las normativas de seguridad y si el usuario no existe
func ValidateUser(user models.User, db *sql.DB) bool {
	var userOK bool = false
	// creamos comprobadores de regex para los datos del usuario
	// Este regex valida cadenas alfanuméricas que pueden contener espacios, guiones bajos o guiones entre palabras, pero no al principio ni al final.
	regexUsername := regexp.MustCompile("^[A-Za-z0-9]+(?:[ _-][A-Za-z0-9]+)*$")
	// Este regex valida cadenas de al menos 8 caracteres que pueden incluir letras, dígitos y ciertos caracteres especiales (@, $, !, %, *, ?, &, _, -).
	regexPass := regexp.MustCompile("^[A-Za-z\\d@$!%*?&_-]{8,}$")
	// Este regex valida direcciones de correos electrónicos
	regexEmail := regexp.MustCompile("^[\\w\\-\\.]+@([\\w-]+\\.)+[\\w-]{2,}$")

	// comprobamos que todos los campos introducidos por el usuario son válidos
	if regexUsername.MatchString(user.Username) && regexPass.MatchString(user.Password) && regexEmail.MatchString(user.Email) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = @username", sql.Named("username", user.Username)).Scan(&count)
		if count == 0 {
			userOK = true
		}
		if err != nil {
			userOK = false
			log.Println("User validation failed:", err)
		}
	}

	return userOK
}

// Registra un usuario en la BD si el usuario es válido.
func Register(c *gin.Context, db *sql.DB) {
	var user models.User
	// recogemos los datos del formulario
	user.Username = c.PostForm("user")
	user.Password = c.PostForm("pass")
	user.Email = c.PostForm("email")

	// comprobamos si el usuario es válido
	if ValidateUser(user, db) {
		// hasheamos la contraseña del usuario
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.HTML(http.StatusOK, "register.html", gin.H{"error": "Error when hashing the password."})
		}

		user.Password = string(hashedPassword)

		// insertamos el usuario en la base de datos
		_, err = db.Exec(
			"INSERT INTO Users (username, password, email) VALUES (@username, @password, @email)",
			sql.Named("username", user.Username), sql.Named("password", user.Password), sql.Named("email", user.Email))
		// si falla cargamos el error en register.html
		if err != nil {
			log.Println("Insert failed:", err)
			c.HTML(http.StatusOK, "register.html", gin.H{"error": err.Error()})
		}
		// si el usuario es válido y no hay fallos, crearemos una cookie con el nombre de usuario y la contraseña
		c.SetCookie("token", user.Username+":"+user.Password, 3600, "/", DOMAIN, false, true)
		// después redireccionaremos al usuario al login
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		log.Println("Invalid credentials were presented")
		c.HTML(http.StatusOK, "register.html", gin.H{"error": "Invalid user or already exists."})
	}
}

// Comprueba las credenciales provistas por el usuario y en caso de ser válidos redirecciona a la página del dashboard.
// En caso de no ser válidos mostramos un error en la página de login.
func Login(c *gin.Context, db *sql.DB) {
	// recogemos los datos y los guardamos en la variable user
	var user models.User
	user.Username = c.PostForm("user")
	user.Password = c.PostForm("pass")

	// rescatamos el hash de la contraseña
	var hashedPassword string
	err := db.QueryRow("SELECT dbo.GET_PASSHASH(@username)", sql.Named("username", user.Username)).Scan(&hashedPassword)
	if err != nil {
		log.Println(err.Error())
	}

	// comparamos la contraseña introducida con el hash de la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{"error": "Credentials are incorrect or user does not exist."})
	} else {
		c.SetCookie("token", user.Username+":"+hashedPassword, 3600, "/", DOMAIN, true, true)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// Comprueba que el token que contiene la cookie es válido
func CheckCookies(c *gin.Context, db *sql.DB) bool {
	// recogemos el token de la cookie
	token, err := c.Cookie("token")
	if err != nil {
		log.Println("Invalid cookie: " + err.Error() + " / " + token)
	}

	// comprobamos que el token sea válido
	var validToken bool
	err = db.QueryRow("SELECT dbo.CHECK_TOKEN(@token)", sql.Named("token", token)).Scan(&validToken)
	if err != nil {
		log.Println("Cookie validation failed: " + err.Error())
	}

	return validToken
}

// Permite que los usuarios añadan libros a la base de datos
func AddBook(c *gin.Context, DB *sql.DB) {
	// Recogemos los datos introducidos por el usuario
	var book models.Book
	var err error
	book.Title = c.PostForm("title")
	book.Author = c.PostForm("author")
	book.UserRating, err = strconv.ParseFloat(c.PostForm("rating"), 64)
	if err != nil {
		log.Println("Invalid book rating: " + err.Error())
		c.HTML(http.StatusSeeOther, "index.html", gin.H{"error": "Invalid book rating"})
	}
	book.Pubdate, err = time.Parse("2006-01-02", c.PostForm("pubdate"))
	// si no se puede parsear la fecha correctamente mostraremos un error
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusSeeOther, "addbook.html", gin.H{"error": "Error when adding the book."})
	} else {
		// Comprobamos que el libro no exista ya de por sí
		var count int
		err = DB.QueryRow("SELECT COUNT(*) FROM Books WHERE title LIKE '%' + @title + '%' AND author LIKE '%' + @author + '½'",
			sql.Named("title", book.Title), sql.Named("author", book.Author)).Scan(&count)
		if err != nil {
			log.Println(err.Error())
			c.HTML(http.StatusSeeOther, "addbook.html", gin.H{"error": "Error when adding the book."})
		}

		// si existe se redirecciona al usuario mostrando un enlace de error y si no, se inserta
		if count > 0 {
			c.HTML(http.StatusSeeOther, "addbook.html", gin.H{"error": "Book already exists."})
		} else {
			_, err = DB.Exec("INSERT INTO Books (title, author, pubdate) VALUES(@title, @author, @pubdate)",
				sql.Named("title", book.Title), sql.Named("author", book.Author), sql.Named("pubdate", book.Pubdate))
			if err != nil {
				log.Println(err.Error())
				c.HTML(http.StatusSeeOther, "addbook.html", gin.H{"error": "Couldn't add the book."})
			}
			// añadimos el libro a las lecturas del usuario
			AddRead(c, DB)
		}
	}
}
