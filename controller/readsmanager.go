package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nicebooks/models"
)

// Devuelve un slice con todos los libros que ha leído el usuario
func GetBooksReadByUser(db *sql.DB, token string) (books []models.Book) {
	var bookList []models.Book

	rows, err := db.Query("SELECT * FROM dbo.GET_READ_BOOKS_BY_TOKEN(@token)", sql.Named("token", token))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Pubdate, &book.UserRating, &book.PublicRating); err != nil {
			log.Println(err)
		}
		bookList = append(bookList, book)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}

	return bookList
}

// Agrega una lectura a la base de datos y si no existe el libro redirecciona a una página para añadirlo
func AddRead(c *gin.Context, db *sql.DB) {
	// almacena si el libro existe
	var bookExists bool
	// almacena si la lectura existe de antes
	var readExists bool
	// recogemos el título del formulario
	title := c.PostForm("title")
	// recogemos la calificación del formulario
	rating := c.PostForm("rating")
	log.Println("Rating: " + rating)

	// se recupera la cookie y se guardan los datos introducidos por el usuario
	token, err := c.Cookie("token")
	if err != nil {
		c.HTML(http.StatusSeeOther, "index.html", gin.H{"error": "Invalid credentials"})
	}

	// comprobamos que un libro con ese título existe
	err = db.QueryRow("SELECT CAST(COUNT(title) as BIT) FROM Books WHERE title LIKE '%'+ @title + '%'", sql.Named("title", title)).Scan(&bookExists)
	if err != nil {
		log.Println(err)
	}

	// Comprobamos que la lectura no exista ya
	err = db.QueryRow("SELECT CAST(COUNT(*) as BIT) FROM Reads WHERE userID = dbo.GET_USER_ID(@token) AND bookID = dbo.GET_BOOK_ID(@title)",
		sql.Named("token", token), sql.Named("title", title)).Scan(&readExists)
	if err != nil {
		log.Println(err)
	}

	// si el libro existe y la lectura no, se añade la lectura
	if bookExists && !readExists {
		_, err = db.Exec(
			"INSERT INTO Reads(userID, bookID, rating) SELECT TOP 1 dbo.GET_USER_ID(@token), id, @rating FROM Books WHERE title LIKE '%'+ @title + '%'",
			sql.Named("token", token), sql.Named("title", title), sql.Named("rating", rating))
		if err != nil {
			log.Println(err)
		}
		c.Redirect(http.StatusSeeOther, "/dashboard")
	} else if readExists {
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Book already exists"})
	} else {
		c.Redirect(http.StatusSeeOther, "/add-book")
	}
}

// Borra una lectura en base al id del libro y el usuario
func DeleteRead(c *gin.Context, db *sql.DB, token string) {
	// recogemos el id del libro
	bookID := c.PostForm("bookID")

	// borramos la lectura de la bd
	rows, err := db.Exec("DELETE FROM Reads WHERE userID = dbo.GET_USER_ID(@token) AND bookID = @bookID",
		sql.Named("token", token), sql.Named("bookID", bookID))
	// si hay algún error (por ejemplo temas de autenticación, etc) mostramos el error
	if err != nil {
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Couldn't delete read."})
		// si es que no se ha insertado mostraremos el error
	} else if nDeleted, err1 := rows.RowsAffected(); err1 != nil || nDeleted == 0 {
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Couldn't delete read."})
	} else {
		// si se completa satisfactoriamente redireccionamos al dashboard
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}
