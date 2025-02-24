package main

import (
	"net/http"
)

func healthzCustomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func faviconCustomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("no favicon"))
}

func main() {
	// run with "go build -o out && ./out" command in a new terminal to start server
	// For fast compile and execution you can use "go run .", this wont save a compiled "out" binary file in the root folder
	sv := http.NewServeMux()
	svStruct := http.Server{
		Addr: ":8080",
		Handler: sv,
	}

	// explorer path to index.html - http://localhost:8080/
	// explorer path to assets/logo.png - http://localhost:8080/assets/logo.png
	sv.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	// explorer path to healthzCustomHandler - http://localhost:8080/healthz
	sv.HandleFunc("/healthz", healthzCustomHandler)

	// CustomHandler to avoid 404 on automatic favicon.ico web browsers request
	sv.HandleFunc("/favicon.ico", faviconCustomHandler)

	svStruct.ListenAndServe()
}