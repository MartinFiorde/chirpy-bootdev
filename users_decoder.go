package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MartinFiorde/chirpy-bootdev/internal/auth"
	"github.com/MartinFiorde/chirpy-bootdev/internal/database"
	"github.com/google/uuid"
)

type UsersParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ExpiresInSeconds int `json:"expires_in_seconds"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func postUsersHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	params, err := UsersdecodeRequestBody(r)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	hashPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	dbParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashPass,
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), dbParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Database error"})
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbParams.Email,
	}

	UsersrespondJSON(w, http.StatusCreated, user)
}

// decodeRequestBody decodes the JSON body from the request.
func UsersdecodeRequestBody(r *http.Request) (*UsersParameters, error) {
	var email UsersParameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&email); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		return nil, err
	}
	if email.ExpiresInSeconds <= 0 || email.ExpiresInSeconds > 3600 {
		email.ExpiresInSeconds = 3600
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

func postLogin(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	params, err := UsersdecodeRequestBody(r)
	// log.Printf("Seconds: %v", params.ExpiresInSeconds)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong"})
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong <email missing in db>"})
		return
	}

	err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong <pass doesnt match>"})
		return
	}

	jwt, err := auth.MakeJWT(dbUser.ID, cfg.secret, time.Duration(params.ExpiresInSeconds) * time.Second) 
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong <jwt error>"})
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     jwt,
	}

	UsersrespondJSON(w, http.StatusOK, user)
}
