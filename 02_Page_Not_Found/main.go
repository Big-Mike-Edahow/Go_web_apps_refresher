// main.go
// Go - Page Not Found

package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", indexHandler)

	log.Println("Listening and serving HTTP on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", logRequest(http.DefaultServeMux)))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
	http.ServeFile(w, r, "./404.html")
	} else {
	http.ServeFile(w, r, "./index.html")
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
