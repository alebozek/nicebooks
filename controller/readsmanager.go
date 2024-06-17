package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"nicebooks/models"
)

// Devuelve un slice con todos los libros que ha leído el usuario
func GetBooksReadByUser(DB *sql.DB, token string) (books []models.Book) {
	var bookList []models.Book

	rows, err := DB.Query("SELECT * FROM dbo.GET_READ_BOOKS_BY_TOKEN(@token)", sql.Named("token", token))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.Title, &book.Author, &book.Pubdate); err != nil {
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
func AddRead(c *gin.Context, DB *sql.DB) {
	var bookExists bool
	var readExists bool
	// se recupera la cookie y se guardan los datos introducidos por el usuario
	token, err := c.Cookie("token")
	if err != nil {
		c.HTML(http.StatusSeeOther, "index.html", gin.H{"error": "Invalid credentials"})
	}

	title := c.PostForm("title")
	err = DB.QueryRow("SELECT CAST(COUNT(title) as BIT) FROM Books WHERE title LIKE '%'+ @title + '%'", sql.Named("title", title)).Scan(&bookExists)
	if err != nil {
		log.Println(err)
	}

	err = DB.QueryRow("SELECT CAST(COUNT(*) as BIT) FROM Reads WHERE userID = dbo.GET_USER_ID(@token) AND bookID = dbo.GET_BOOK_ID(@title)",
		sql.Named("token", token), sql.Named("title", title)).Scan(&readExists)
	if err != nil {
		log.Println(err)
	}

	// comprobamos que existe al menos un libro con ese título
	if bookExists && !readExists {
		_, err = DB.Exec(
			"INSERT INTO Reads(userID, bookID) SELECT TOP 1 dbo.GET_USER_ID(@token), id FROM Books WHERE title LIKE '%'+ @title + '%'",
			sql.Named("token", token), sql.Named("title", title))
		if err != nil {
			log.Println(err)
		}
		c.Redirect(http.StatusSeeOther, "/dashboard")
	} else if readExists {
		c.HTML(http.StatusSeeOther, "dashboard.html", gin.H{"BookList": GetBooksReadByUser(DB, token), "Err": "Book already exists"})
	} else {
		c.Redirect(http.StatusSeeOther, "/add-book")
	}

}
