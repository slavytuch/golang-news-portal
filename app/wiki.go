//go:build ignore

package main

import (
    "os"
	"log"
   	"net/http"
	"html/template"
	"regexp"
	"strings"
)

type Page struct {
    Title string
    Body  []byte
}

type PageLink struct {
	Title string
	Link string
}

func (p *Page) save() error {
    filename := "./data/" + p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates *template.Template

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl + ".html", p);
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadPage(title string) (*Page, error) {
    filename := "./data/" + title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    page, err := loadPage(title)
	if (err != nil) {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return;
	}
    
	renderTemplate(w, "view", page);
}

func editHandler(writer http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if (err != nil) {
		page = &Page{Title: title};
	}

	renderTemplate(writer, "edit", page);
}

func saveHandler(writer http.ResponseWriter, request *http.Request, title string) {
	body := request.FormValue("body")
	page := Page{Title: title, Body: []byte(body)}
	err := page.save()

	if (err != nil) {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return;
	}

	http.Redirect(writer, request, "/view/" + title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func indexHandler (w http.ResponseWriter, r *http.Request) {
	page, err := loadPage("FrontPage");

	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	dataList, err1 := os.ReadDir("./data");

	if (err1 != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var pageList []PageLink;

	for _,file := range(dataList) {
		pageTitle := strings.Replace(file.Name(), ".txt", "", -1);
		pageList = append(pageList, PageLink{Title:pageTitle, Link: "/view/" + pageTitle})
	}

	err = templates.ExecuteTemplate(w, "index.html", map[string]any{
		"Title": page.Title,
		"Body": page.Body,
		"List": pageList,
	});
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	var templateDirPath = "./tmpl";
	files, err := os.ReadDir(templateDirPath);

	if (err != nil) {
		log.Fatal("Error loading template folder: " + err.Error());
	}

	var templateList[]string;

	for _, template := range files {
		templateList = append(templateList, templateDirPath + "/" + template.Name())
	}

	templates = template.Must(template.ParseFiles(templateList...))
	
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}