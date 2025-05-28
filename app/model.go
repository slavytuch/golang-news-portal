package main

import (
	"database/sql"
)

type NewsItem struct {
	Id    int
	Title string
	Body  string
	Image sql.NullString
}
