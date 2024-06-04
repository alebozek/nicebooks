package models

import "time"

type Book struct {
	ID      int
	Title   string
	Author  string
	Pubdate time.Time
}
