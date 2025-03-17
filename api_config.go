package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/MartinFiorde/chirpy-bootdev/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	secret         string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsCustomHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("metrics.html")
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte("server error"))
		return
	}

	htmlString := string(fileBytes)
	dinamicValue := int64(cfg.fileserverHits.Load())
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(htmlString, dinamicValue)))
}

func (cfg *apiConfig) resetCustomHandler(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte("db error"))
		return
	}

	// NOT NEEDED - DeleteAllUsers(...) already delete all related rows in other tables with "ON DELETE CASCADE" setting

	// err = cfg.db.DeleteAllChirps(r.Context())
	// if err != nil {
	// 	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	// 	w.WriteHeader(500)
	// 	w.Write([]byte("db error"))
	// 	return
	// }

	// err = cfg.db.DeleteAllRefreshTokens(r.Context())
	// if err != nil {
	// 	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	// 	w.WriteHeader(500)
	// 	w.Write([]byte("db error"))
	// 	return
	// }

	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)))
}
