package main

import (
	"net/http"
)

func main() {
	// run with "go build -o out && ./out" command in a new terminal to start server
	sv := http.NewServeMux()
	svStruct := http.Server{
		Addr: ":8080",
		Handler: sv,
	}

	svStruct.ListenAndServe()
}