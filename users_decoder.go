package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UsersParameters struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func postUsersHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	email, err := UsersdecodeRequestBody(r)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), email.Email)
	if err != nil {
        log.Printf("Error creating user: %s", err)
        respondJSON(w, http.StatusInternalServerError, Response{Error: "Database error"})
        return
    }


	responseUser := User{
		ID:        user.ID,          
		CreatedAt: user.CreatedAt,   
		UpdatedAt: user.UpdatedAt,   
		Email:     user.Email,       
	}

	UsersrespondJSON(w, http.StatusCreated, responseUser)
}

// decodeRequestBody decodes the JSON body from the request.
func UsersdecodeRequestBody(r *http.Request) (*UsersParameters, error) {
	var email UsersParameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&email); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		return nil, err
	}
	return &email, nil
}

// respondJSON sends a JSON response with the given status and payload.
func UsersrespondJSON(w http.ResponseWriter, status int, user User) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
