package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nicebooks/models"
	"strconv"
	"time"
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

func GetBookByRead(db *sql.DB, token string, id int) (models.Book, error) {
	var book models.Book

	err := db.QueryRow("SELECT * FROM dbo.GET_BOOK_READ(@id, @token)",
		sql.Named("id", id), sql.Named("token", token)).Scan(&book.ID, &book.Title, &book.Author, &book.Pubdate, &book.UserRating)
	if err != nil {
		log.Println(err)
	}

	return book, err
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

func EditRead(c *gin.Context, db *sql.DB, token string) {
	var book models.Book
	var err error
	book.ID, err = strconv.Atoi(c.PostForm("bookid"))
	if err != nil {
		log.Println("Invalid book id")
		c.Redirect(http.StatusSeeOther, "/")
	}
	book.Title = c.PostForm("title")
	book.Author = c.PostForm("author")
	book.UserRating, err = strconv.ParseFloat(c.PostForm("rating"), 64)
	if err != nil {
		log.Println("Invalid book rating")
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Could not edit the book."})
	}
	book.Pubdate, err = time.Parse("2006-01-02", c.PostForm("pubdate"))
	// si no se puede parsear la fecha correctamente mostraremos un error
	if err != nil {
		log.Println(err.Error())
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Could not edit the book."})
	}

	log.Println(book)

	// empezamos la transacción
	transaction, err := db.Begin()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Could not edit the book."})
	}
	// ejecutamos la query
	var rows sql.Result
	rows, err = transaction.Exec("EXEC dbo.UPDATE_BOOK_READ @token, @bookID, @title, @author, @pubdate, @rating",
		sql.Named("token", token), sql.Named("bookID", book.ID), sql.Named("title", book.Title),
		sql.Named("author", book.Author), sql.Named("pubdate", book.Pubdate), sql.Named("rating", book.UserRating))
	if err != nil {
		log.Println(err)
		transaction.Rollback()
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(db, token), "Err": "Could not edit the book."})
	}

	nRowsAffected, err1 := rows.RowsAffected()
	if err1 != nil || nRowsAffected != 2 {
		transaction.Rollback()
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"Err": "Could not edit the book."})
	} else {
		transaction.Commit()
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}
