package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	var newsList []NewsItem

	rows, err := db.Query("SELECT * FROM news")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var newsItem NewsItem

		if err := rows.Scan(&newsItem.Id, &newsItem.Title, &newsItem.Body, &newsItem.Image); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newsList = append(newsList, newsItem)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))

	tmpl.Execute(w, newsList)
}

func detailPageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := LoadNewsItem(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/detail.html"))
	tmpl.Execute(w, newsItem)
}

func editPageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := LoadNewsItem(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/edit.html"))
	tmpl.Execute(w, newsItem)
}

func updateNewsHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("image")

	var savedImage *string

	if file != nil {
		savedImage, err = saveImage(file, header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	newsItem := NewsItem{Id: newsId, Title: r.FormValue("title"), Body: r.FormValue("body"), Image: savedImage}
	err = newsItem.Save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/detail/"+strconv.Itoa(newsId), http.StatusFound)
}

func saveImage(f multipart.File, h *multipart.FileHeader) (*string, error) {

	err := SaveImage(f, h.Filename)

	if err != nil {
		return nil, err
	}

	return &h.Filename, nil
}

func createPageHandler(w http.ResponseWriter, r *http.Request) {

	file, header, _ := r.FormFile("image")

	var savedImage *string
	var err error

	if file != nil {

		savedImage, err = saveImage(file, header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	n, err := CreateNewsItem(r.FormValue("title"), r.FormValue("body"), savedImage)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if n == nil {
		http.Error(w, "Error loading news item", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/detail/"+strconv.Itoa(*n), http.StatusFound)
}

func creationPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/create.html"))

	tmpl.Execute(w, nil)
}

func main() {

	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	dbConfig := mysql.NewConfig()

	dbConfig.Addr = os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")
	dbConfig.User = os.Getenv("DB_USERNAME")
	dbConfig.Passwd = os.Getenv("DB_PASSWORD")
	dbConfig.Net = "tcp"
	dbConfig.DBName = os.Getenv("DB_DATABASE")

	db, err = sql.Open("mysql", dbConfig.FormatDSN())

	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	rootdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", indexPageHandler)
	mux.HandleFunc("GET /detail/{id}", detailPageHandler)
	mux.HandleFunc("GET /edit/{id}", editPageHandler)
	mux.HandleFunc("POST /update/{id}", updateNewsHandler)
	mux.HandleFunc("GET /new/", creationPageHandler)
	mux.HandleFunc("POST /create/", createPageHandler)
	mux.Handle("GET /images/", http.StripPrefix("/images", http.FileServer(http.Dir(path.Join(rootdir, "images/")))))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
