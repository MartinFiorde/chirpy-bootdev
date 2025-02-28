package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)))
}
