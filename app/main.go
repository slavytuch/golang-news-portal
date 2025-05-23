package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

type NewsItem struct {
	Id    int
	Title string
	Body  string
	Image sql.NullString
}

var db *sql.DB

func getNewsById(id int) (*NewsItem, error) {
	row, err := db.Query("SELECT * FROM news WHERE id = ?", id)

	if err != nil {
		return nil, err
	}

	var newsItem NewsItem

	row.Next()
	if err := row.Scan(&newsItem.Id, &newsItem.Title, &newsItem.Body, &newsItem.Image); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &newsItem, nil
}

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
	newsId, err := strconv.Atoi(r.URL.Path[len("/detail/"):])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := getNewsById(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/detail.html"))
	tmpl.Execute(w, newsItem)
}

func editPageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.URL.Path[len("/edit/"):])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := getNewsById(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/edit.html"))
	tmpl.Execute(w, newsItem)
}

func updatePageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.URL.Path[len("/update/"):])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE news SET title = ?, body = ? WHERE id = ?", r.FormValue("title"), r.FormValue("body"), newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, header, _ := r.FormFile("image")

	if file != nil {

		savedImage, err := saveImage(file, header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var currentImage string
		db.QueryRow("SELECT image FROM news where id = ?", newsId).Scan(&currentImage)

		fmt.Print("Current image:" + currentImage)

		_, err = db.Exec("UPDATE news SET image = ? WHERE id = ?", savedImage, newsId)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if currentImage != "" {
			os.Remove("./images/" + currentImage)
		}
	}

	http.Redirect(w, r, "/detail/"+strconv.Itoa(newsId), http.StatusFound)
}

func saveImage(f multipart.File, h *multipart.FileHeader) (*string, error) {
	defer f.Close()

	path := filepath.Join(".", "images")

	fullPath := path + "/" + h.Filename

	fd, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	_, err = io.Copy(fd, f)

	if err != nil {
		return nil, err
	}

	return &h.Filename, nil
}

func createPageHandler(w http.ResponseWriter, r *http.Request) {

	result, err := db.Exec("INSERT INTO news (title, body) VALUES (?, ?)", r.FormValue("title"), r.FormValue("body"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	insertedId, err := result.LastInsertId()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, header, _ := r.FormFile("image")

	if file != nil {

		savedImage, err := saveImage(file, header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE news SET image = ? WHERE id = ?", savedImage, insertedId)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/detail/"+strconv.Itoa(int(insertedId)), http.StatusFound)
}

func creationPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/create.html"))

	tmpl.Execute(w, nil)
}

func main() {

	dbConfig := mysql.NewConfig()

	dbConfig.Addr = "mysql:3306"
	dbConfig.User = "sail"
	dbConfig.Passwd = "password"
	dbConfig.Net = "tcp"
	dbConfig.DBName = "laravel"

	var err error

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

	http.HandleFunc("/", indexPageHandler)
	http.HandleFunc("/detail/", detailPageHandler)
	http.HandleFunc("/edit/", editPageHandler)
	http.HandleFunc("/update/", updatePageHandler)
	http.HandleFunc("/new/", creationPageHandler)
	http.HandleFunc("/create/", createPageHandler)
	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir(path.Join(rootdir, "images/")))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
