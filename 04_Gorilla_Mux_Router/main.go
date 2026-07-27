// main.go
// Go - Gorilla Mux Router

package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = ":3306"
	DBUser  = "user"
	DBPass  = "password"
	DBDbase = "go_building_web_apps"
	PORT    = ":8000"
)

var database *sql.DB

type Page struct {
	Title   string
	Content string
	Date    string
	GUID    string
}

func main() {
	dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, DBPass, DBDbase)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err.Error())
	} else {
		log.Println("Connected to database...")
	}
	database = db

	routes := mux.NewRouter()

	fileServer := http.FileServer(http.Dir("./static/"))
	routes.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	routes.HandleFunc("/favicon.ico", ServeFavicon)
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
	routes.HandleFunc("/", RedirIndex)
	routes.HandleFunc("/home", ServeIndex)
	http.Handle("/", routes)

	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

func ServeFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/favicon.ico")
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}
	pages, err := database.Query("SELECT page_title, page_content, page_date, page_guid FROM pages ORDER BY ? DESC", "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	defer pages.Close()
	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date, &thisPage.GUID)
		Pages = append(Pages, thisPage)
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title, page_content, page_date FROM pages WHERE page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return
	}

	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}

func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}

func (p Page) TruncatedText() string {
	chars := 0
	for i := range p.Content {
		chars++
		if chars > 150 {
			return p.Content[:i] + ` ...`
		}
	}
	return p.Content
}

// Log requests to the console.
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
