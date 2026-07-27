// main.go
// Go - Connecting to MySQL

package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Post struct {
	ID    int
	Title string
	Body  string
	Date  string
}

var db *sql.DB
var tmpl *template.Template

func main() {
	var err error
	dsn := "user:password@tcp(127.0.0.1:3306)/blog_db"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Load all templates from templates directory.
	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	// Serve static files.
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route Handlers.
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post/", getPostHandler)
	http.HandleFunc("/post/create", createPostHandler)

	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

// Favicon handler.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

// Home page handler.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	rows, err := db.Query("SELECT id, title, body, date FROM posts ORDER BY date DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	// Render the home page template.
	tmpl.ExecuteTemplate(w, "index.html", posts)
}

// View post handler.
func getPostHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path like /post/1.
	idStr := r.URL.Path[len("/post/"):]

	row := db.QueryRow("SELECT id, title, body, date FROM posts WHERE id = ?", idStr)

	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Body, &p.Date)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render view post template.
	tmpl.ExecuteTemplate(w, "post.html", p)
}

// Handle Form Submission (Create Post).
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form values from request body.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	// Insert into MySQL database.
	_, err = db.Exec("INSERT INTO posts (title, body) VALUES (?, ?)", title, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to home page after success.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Log requests to the console.
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
