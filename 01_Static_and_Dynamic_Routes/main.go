// main.go
// Go - Static and Dynamic Routes

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/date", dateHandler)

	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func dateHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, "Current Date and Time: %s", currentTime)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
