package main

import (
	"net/http"
)

func main() {
	sv := http.NewServeMux()
	svStruct := http.Server{
		Addr: ":8080",
		Handler: sv,
	}

	svStruct.ListenAndServe()
}