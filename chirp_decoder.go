package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Parameters struct {
	Body string `json:"body"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Valid bool   `json:"valid,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	params, err := decodeRequestBody(r)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	if len(params.Body) > 140 {
		respondJSON(w, http.StatusBadRequest, Response{Error: "Chirp is too long"})
		return
	}

	respondJSON(w, http.StatusOK, Response{Valid: true, CleanedBody: censorProfanity(params.Body)})
}

// decodeRequestBody decodes the JSON body from the request.
func decodeRequestBody(r *http.Request) (*Parameters, error) {
	var params Parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		return nil, err
	}
	return &params, nil
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
