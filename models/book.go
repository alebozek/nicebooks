package models

import (
	"fmt"
	"time"
)

// Clase que representa los libros que pueden haber leído los usuarios
type Book struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Author  string    `json:"author"`
	Pubdate time.Time `json:"pubdate"`
}

// Transforma la fecha en algo más legible
func (b Book) TransformDate() string {
	year, month, day := b.Pubdate.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func (b Book) String() string {
	return fmt.Sprintf("%s %s %s", b.Title, b.Author, b.Pubdate.String())
}
