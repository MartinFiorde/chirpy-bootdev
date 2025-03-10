package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MartinFiorde/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
)

type Parameters struct {
	Body string `json:"body"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Valid bool   `json:"valid,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
}

// respondJSON sends a JSON response with the given status and payload.
func respondJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func censorProfanity(s string) string {
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(s) // más eficiente que Split(s, " ")
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if _, found := profaneWords[lowerWord]; found {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

type ChirpParameters struct {
	Body   string    `json:"body"`
	UserID string `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func postChirpsHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	chirpRequest, err := ChirpsDecodeRequestBody(w, r)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	chirp, err := cfg.db.CreateChirps(r.Context(), *chirpRequest)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Database error: "})
		return
	}

	if len(chirp.Body) > 140 {
		respondJSON(w, http.StatusBadRequest, Response{Error: "Chirp is too long"})
		return
	}

	responseUser := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      censorProfanity(chirp.Body),
		UserId:    chirp.UserID,
	}

	ChirpsRespondJSON(w, http.StatusCreated, responseUser)
}

// decodeRequestBody decodes the JSON body from the request.
func ChirpsDecodeRequestBody(w http.ResponseWriter, r *http.Request) (*database.CreateChirpsParams, error) {
	var chirpJson ChirpParameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirpJson); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		return nil, err
	}

	userUUID, err := uuid.Parse(chirpJson.UserID)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		respondJSON(w, http.StatusBadRequest, Response{Error: "Invalid user_id format"})
		return nil, err
	}

	chirp := database.CreateChirpsParams{
		Body: chirpJson.Body,
		UserID: userUUID,
	}
	return &chirp, nil
}

// respondJSON sends a JSON response with the given status and payload.
func ChirpsRespondJSON(w http.ResponseWriter, status int, chirp Chirp) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
