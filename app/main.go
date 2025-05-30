package main

import (
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/slavytuch/go-news-portal/news"
)

func indexPageHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))

	newsList, err := newsRepository.All()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, newsList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func detailPageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := newsRepository.Load(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/detail.html"))
	err = tmpl.Execute(w, newsItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editPageHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := newsRepository.Load(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/edit.html"))
	err = tmpl.Execute(w, newsItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	oldNewsItem, err := newsRepository.Load(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if oldNewsItem == nil {
		http.NotFound(w, r)
		return
	}

	err = newsRepository.Save(news.NewsItem{Id: newsId, Title: r.FormValue("title"), Body: r.FormValue("body"), Image: savedImage})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if savedImage != nil && oldNewsItem.Image != savedImage {
		DeleteImage(*oldNewsItem.Image)
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

func createNewsHandler(w http.ResponseWriter, r *http.Request) {

	file, header, _ := r.FormFile("image")

	var savedImageName *string
	var err error

	if file != nil {

		savedImageName, err = saveImage(file, header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	id, err := newsRepository.Create(r.FormValue("title"), r.FormValue("body"), savedImageName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if id == nil {
		http.Error(w, "Error loading news item", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/detail/"+strconv.Itoa(*id), http.StatusFound)
}

func creationPageHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/create.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func deleteNewsHandler(w http.ResponseWriter, r *http.Request) {
	newsId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newsItem, err := newsRepository.Load(newsId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if newsItem == nil {
		http.NotFound(w, r)
		return
	}

	err = newsRepository.Delete(newsId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if newsItem.Image != nil {
		DeleteImage(*newsItem.Image)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

var newsRepository news.NewsRepository

func main() {

	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	db, err := OpenDatabaseConnection(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
	)

	if err != nil {
		log.Fatal("Error connecting to database: " + err.Error())
	}

	newsRepository = &news.MysqlNewsRepository{DB: db}

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
	mux.HandleFunc("POST /create/", createNewsHandler)
	mux.HandleFunc("POST /delete/{id}", deleteNewsHandler)
	mux.Handle("GET /images/", http.StripPrefix("/images", http.FileServer(http.Dir(path.Join(rootdir, "images/")))))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
