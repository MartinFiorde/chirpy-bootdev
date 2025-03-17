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
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func postCreateUserHandler(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
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

	jwt, err := auth.MakeJWT(dbUser.ID, cfg.secret, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong <jwt error>"})
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong<fail to generate refresh token>"})
		return
	}

	refreshTokenDB, err := cfg.db.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:  refreshToken,
			UserID: dbUser.ID,
		})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong<fail to save refresh token in db>"})
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        jwt,
		RefreshToken: refreshTokenDB.Token,
	}

	UsersrespondJSON(w, http.StatusOK, user)
}

func postRefresh(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	userBearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 1"})
		return
	}

	tokenDB, err := cfg.db.GetRefreshTokenByToken(r.Context(), userBearerToken)
	if err != nil || tokenDB.RevokedAt.Valid || tokenDB.ExpiresAt.Before(time.Now()) {
		respondJSON(w, http.StatusUnauthorized, Response{Error: "invalid refresh token"})
		return
	}

	jwt, err := auth.MakeJWT2(tokenDB.UserID, cfg.secret)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 1"})
		return
	}

	respondJSON(w, http.StatusOK, Response{Token: jwt})
}

func postRevoke(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	userBearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 4"})
		return
	}

	tokenDB, err := cfg.db.GetRefreshTokenByToken(r.Context(), userBearerToken)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, Response{Error: "invalid refresh token"})
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), tokenDB.Token)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Something went wrong 5"})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
}

func putChangePassword(cfg *apiConfig, w http.ResponseWriter, r *http.Request) {
	userBearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 4"})
		return
	}

	userID, err := auth.ValidateJWT(userBearerToken, cfg.secret)
	if err != nil {
		log.Printf("Error: %v", err)
		respondJSON(w, http.StatusUnauthorized, Response{Error: "Something went wrong 2"})
		return
	}

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

	dbParams := database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashPass,
	}

	dbUser, err := cfg.db.UpdateUser(r.Context(), dbParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondJSON(w, http.StatusInternalServerError, Response{Error: "Database error"})
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	UsersrespondJSON(w, http.StatusOK, user)

}
