package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"nicebooks/models"
	"regexp"
)

// Comprueba si los campos del usuario no están vacíos y si el usuario no existe
func ValidateUser(user models.User, DB *sql.DB) bool {
	var userOK bool = false
	// creamos comprobadores de regex para los datos del usuario
	regexUsername := regexp.MustCompile("^[A-Za-z0-9]+(?:[ _-][A-Za-z0-9]+)*$")
	regexPass := regexp.MustCompile("^[A-Za-z\\d@$!%*?&_-]{8,}$")
	regexEmail := regexp.MustCompile("^[\\w\\-\\.]+@([\\w-]+\\.)+[\\w-]{2,}$")

	// comprobamos que todos los campos introducidos por el usuario son válidos
	if regexUsername.MatchString(user.Username) && regexPass.MatchString(user.Password) && regexEmail.MatchString(user.Email) {
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM Users WHERE username = @username", sql.Named("username", user.Username)).Scan(&count)
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
func Register(c *gin.Context, DB *sql.DB) {
	var user models.User
	// recogemos los datos del formulario
	user.Username = c.PostForm("user")
	user.Password = c.PostForm("pass")
	user.Email = c.PostForm("email")

	if ValidateUser(user, DB) {
		// hasheamos la contraseña del usuario
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.HTML(http.StatusOK, "register.html", gin.H{"error": "Error when hashing the password."})
		}

		user.Password = string(hashedPassword)

		// insertamos el usuario en la base de datos
		_, err = DB.Exec(
			"INSERT INTO Users (username, password, email) VALUES (@username, @password, @email)",
			sql.Named("username", user.Username), sql.Named("password", user.Password), sql.Named("email", user.Email))
		// si falla cargamos el error en register.html
		if err != nil {
			c.HTML(http.StatusOK, "register.html", gin.H{"error": err.Error()})
		}
		c.SetCookie("token", user.Username+":"+user.Password, 3600, "/", "localhost", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		c.HTML(http.StatusOK, "register.html", gin.H{"error": "Invalid user or already exists."})
	}
}

// Comprueba que los credenciales provistos por el usuario y en caso de ser válidos redirecciona a la página del dashboard.
// En caso de no ser válidos mostramos un error en la página de login.
func Login(c *gin.Context, DB *sql.DB) {
	// recogemos los datos y los guardamos en la variable user
	var user models.User
	user.Username = c.PostForm("user")
	user.Password = c.PostForm("pass")

	// rescatamos el hash de la contraseña
	var hashedPassword string
	err := DB.QueryRow("SELECT dbo.GET_PASSHASH(@username)", sql.Named("username", user.Username)).Scan(&hashedPassword)
	if err != nil {
		log.Println(err.Error())
	}

	// comparamos la contraseña introducida con el hash de la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{"error": "Credentials are incorrect or user does not exist."})
	} else {
		c.SetCookie("token", user.Username+":"+hashedPassword, 3600, "/", "localhost", false, true)
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// Comprueba que el token que contiene la cookie es válido
func CheckCookies(c *gin.Context, DB *sql.DB) bool {
	// recogemos el token de la cookie
	token, err := c.Cookie("token")
	if err != nil {
		log.Println("Invalid cookie: " + err.Error() + " / " + token)
	}

	// comprobamos que el token sea válido
	var validToken bool
	err = DB.QueryRow("SELECT dbo.CHECK_TOKEN(@token)", sql.Named("token", token)).Scan(&validToken)
	if err != nil {
		log.Println("Cookie validation failed: " + err.Error())
	}

	return validToken
}
