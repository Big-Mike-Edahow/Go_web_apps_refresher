// main.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

var DB *sql.DB

func main() {
	// Initialize database.
	initDB()

	// Serve static files.
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up the routes.
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/users", getUsers)

	// Start the HTTP server.
	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Query the database to get users
	rows, err := DB.Query("SELECT id, name, email, created FROM users order by id asc")
	if err != nil {
		http.Error(w, "Error fetching users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Created); err != nil {
			http.Error(w, "Error scanning user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	files := []string{
		"./templates/base.html",
		"./templates/index.html",
	}

	indexTemplate, _ := template.ParseFiles(files...)
	indexTemplate.ExecuteTemplate(w, "base", users)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
		"./templates/about.html",
	}

	aboutTemplate, _ := template.ParseFiles(files...)
	aboutTemplate.ExecuteTemplate(w, "base", nil)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, name, email, created FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Created); err != nil {
			log.Println("Error scanning user", err)
			continue
		}
		users = append(users, user)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	var errConn error
	DB, errConn = sql.Open("postgres", connStr)
	if errConn != nil {
		log.Fatalf("Error opening database: %v", errConn)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Cannot connect to the database: %v", err)
	}

	log.Println("Database connected successfully!")
}

// Log requests to the console.
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
