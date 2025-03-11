package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/MartinFiorde/chirpy-bootdev/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening DB: %s", err)
	}
	dbQueries := database.New(db)

	// run with "go build -o out && ./out" command in a new terminal to start server
	// For fast compile and execution you can use "go run .", this wont save a compiled "out" binary file in the root folder
	sv := http.NewServeMux()
	svStruct := http.Server{
		Addr:    ":8080",
		Handler: sv,
	}

	apiCfg := apiConfig{
		db: dbQueries,
	}

	// explorer path to index.html - http://localhost:8080/
	// explorer path to assets/logo.png - http://localhost:8080/assets/logo.png
	sv.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// explorer path to healthzCustomHandler - http://localhost:8080/healthz
	sv.HandleFunc("GET /api/healthz", healthzCustomHandler)

	// explorer paths to metrics and reset - http://localhost:8080/metrics + http://localhost:8080/reset
	sv.HandleFunc("GET /admin/metrics", apiCfg.metricsCustomHandler)
	sv.HandleFunc("POST /admin/reset", apiCfg.resetCustomHandler)

	// CustomHandler to avoid 404 on automatic favicon.ico web browsers request
	sv.HandleFunc("/favicon.ico", faviconCustomHandler)

	// CustomHandler to save users
	sv.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		postUsersHandler(&apiCfg, w, r)
	})

	// CustomHandler to save chirps
	sv.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		postChirpsHandler(&apiCfg, w, r)
	})

	// CustomHandler to get all chirps
	sv.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		getChirpsHandler(&apiCfg, w, r)
	})

	// CustomHandler to get one chirp by id
	sv.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		getChirpByIdHandler(&apiCfg, w, r)
	})

	// CustomHandler to login
	sv.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		postLogin(&apiCfg, w, r)
	})

	svStruct.ListenAndServe()
}
