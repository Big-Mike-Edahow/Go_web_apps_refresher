// main.go
// Go - Relational Database Tables

package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Post struct {
	ID        int
	Title     string
	Content   string
	CreatedAt string
}

type Comment struct {
	ID        int
	PostID    int
	Content   string
	CreatedAt string
}

// PageData wraps the data needed for the single post template view
type PageData struct {
	Post     Post
	Comments []Comment
}

var db *sql.DB
var tmpl *template.Template

func main() {
	var err error

	// Pre-parse the HTML templates at startup.
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates: ", err)
	}

	dsn := "user:password@tcp(127.0.0.1:3306)/go_building_web_apps?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to database on port:3306")
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/comment", commentHandler)

	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		_, err := db.Exec("INSERT INTO posts (title, content) VALUES (?, ?)", title, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("SELECT id, title, content, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') FROM posts ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			continue
		}
		posts = append(posts, p)
	}

	// Render index.html template and pass the slice of posts
	err = tmpl.ExecuteTemplate(w, "index.html", posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var p Post
	err = db.QueryRow("SELECT id, title, content, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') FROM posts WHERE id = ?", id).Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rows, err := db.Query("SELECT id, content, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') FROM comments WHERE post_id = ? ORDER BY id DESC", id)
	var comments []Comment
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c Comment
			if err := rows.Scan(&c.ID, &c.Content, &c.CreatedAt); err == nil {
				comments = append(comments, c)
			}
		}
	}

	// Combine data into our structure wrapper
	data := PageData{
		Post:     p,
		Comments: comments,
	}

	// Render post.html template and pass the container object
	err = tmpl.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		postID := r.FormValue("post_id")
		content := r.FormValue("content")
		_, err := db.Exec("INSERT INTO comments (post_id, content) VALUES (?, ?)", postID, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
	}
}

// Log requests to the console.
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
