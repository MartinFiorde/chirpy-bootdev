package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// log.Printf("Expire time: %v", expiresIn)
	claims := jwt.RegisteredClaims{
		Issuer: "Chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(tokenSecret), nil
    }

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, keyFunc) 
	if err != nil {
		return uuid.UUID{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok || !token.Valid {
        return uuid.UUID{}, errors.New("invalid token")
    }

	user, err := claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	userId, err := uuid.Parse(user)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	// log.Printf("Headers: %v", headers)
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return authHeader, errors.New("missing authorization header")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
} 