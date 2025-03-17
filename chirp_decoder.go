package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MartinFiorde/chirpy-bootdev/internal/auth"
	"github.com/MartinFiorde/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
)

type Parameters struct {
	Body string `json:"body"`
}

type Response struct {
	Error       string `json:"error,omitempty"`
	Valid       bool   `json:"valid,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
	Token       string `json:"token,omitempty"`
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
	Body   string `json:"body"`
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
	userBearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 1"})
		return
	}

	userID, err := auth.ValidateJWT(userBearerToken, cfg.secret)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 2"})
		return
	}

	chirpRequest, err := ChirpsDecodeRequestBody(w, r, userID)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 3"})
		return
	}

	chirp, err := cfg.db.CreateChirps(r.Context(), *chirpRequest)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Database error: "})
		return
	}

	if len(chirp.Body) > 140 {
		log.Printf("Error: Chirp is too long")
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
func ChirpsDecodeRequestBody(w http.ResponseWriter, r *http.Request, userID uuid.UUID) (*database.CreateChirpsParams, error) {
	var chirpJson ChirpParameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirpJson); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		return nil, err
	}

	chirp := database.CreateChirpsParams{
		Body:   chirpJson.Body,
		UserID: userID,
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

func getChirpsHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {

	authorID, err := uuid.Parse(r.URL.Query().Get("author_id"))
	if err != nil {
		log.Printf("Error - Invalid UUID format: %v", err)
		http.Error(w, "Invalid chirp_id format", http.StatusBadRequest) // respondJSON(w, http.StatusBadRequest, Response{Error: "Invalid user_id format"})
		return
	}

	var dbChirps []database.Chirp
	if authorID != uuid.Nil {
		dbChirps, err = cfg.db.GetChirpsByAuthorID(r.Context(), authorID)
		if err != nil {
			http.Error(w, "Error fetching chirps", http.StatusInternalServerError)
			log.Println("DB Error:", err)
			return
		}
	} else {
		dbChirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			http.Error(w, "Error fetching chirps", http.StatusInternalServerError)
			log.Println("DB Error:", err)
			return
		}
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, c := range dbChirps {
		chirps[i] = Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserId:    c.UserID,
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirps)
}

func getChirpByIdHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.Printf("Error - Invalid UUID format: %v", err)
		http.Error(w, "Invalid chirp_id format", http.StatusBadRequest) // respondJSON(w, http.StatusBadRequest, Response{Error: "Invalid user_id format"})
		return
	}

	dbChirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		log.Println("DB Error:", err)
		http.Error(w, "Chirp not found", http.StatusNotFound)
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserId:    dbChirp.UserID,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}

func deleteChirpByID(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.Printf("Error - Invalid UUID format: %v", err)
		http.Error(w, "Invalid chirp_id format", http.StatusBadRequest) // respondJSON(w, http.StatusBadRequest, Response{Error: "Invalid user_id format"})
		return
	}

	chirpdb, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		log.Printf("Error - Invalid UUID format: %v", err)
		respondJSON(w, http.StatusNotFound, Response{Error: "Something went wrong 1"})
		return
	}

	userBearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 2"})
		return
	}

	userID, err := auth.ValidateJWT(userBearerToken, cfg.secret)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 3"})
		return
	}

	if chirpdb.UserID != userID {
		log.Printf("Error - Invalid UUID format: %v", err)
		respondJSON(w, http.StatusForbidden, Response{Error: "Something went wrong 4"})
		return
	}

	err = cfg.db.DeleteChirpById(r.Context(),chirpId)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 5"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
