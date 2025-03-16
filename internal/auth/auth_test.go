package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	uuid1, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
	uuid2, _ := uuid.Parse("3d6f5f94-3b4d-4a71-9835-6397d3da9a56")
	secretToken := "ImSicrit!1234"
	jwt1, _ := MakeJWT(uuid1, secretToken, 24*time.Hour)
	jwt2, _ := MakeJWT(uuid2, secretToken, 24*time.Hour)

	tests := []struct {
		name           string
		tokenString    string
		tokenSecret    string
		expectedUserID uuid.UUID
		funcErr        bool
		resultErr      bool
	}{
		{
			name:           "Correct JWT",
			tokenString:    jwt1,
			tokenSecret:    secretToken,
			expectedUserID: uuid1,
			funcErr:        false,
			resultErr:      false,
		},
		{
			name:           "Incorrect JWT",
			tokenString:    "invalid_jwt",
			tokenSecret:    secretToken,
			expectedUserID: uuid1,
			funcErr:        true,
			resultErr:      true,
		},
		{
			name:           "JWT doesn't match different uuid",
			tokenString:    jwt2,
			tokenSecret:    secretToken,
			expectedUserID: uuid1,
			funcErr:        false,
			resultErr:      true,
		},
		{
			name:           "Incorrect secretToken",
			tokenString:    jwt1,
			tokenSecret:    "incorrectToken",
			expectedUserID: uuid1,
			funcErr:        true,
			resultErr:      true,
		},
		{
			name:           "Empty JWT",
			tokenString:    "",
			tokenSecret:    secretToken,
			expectedUserID: uuid1,
			funcErr:        true,
			resultErr:      true,
		},
		{
			name:           "Incorrect uuid result",
			tokenString:    jwt1,
			tokenSecret:    secretToken,
			expectedUserID: uuid2,
			funcErr:        false,
			resultErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.funcErr {
				t.Errorf("ValidateJWT() error = %v, expected funcErr %v", err, tt.funcErr)
			}
			if (resultUserID != tt.expectedUserID) != tt.resultErr {
				t.Errorf("ValidateJWT() resultUserID = %v, expectedUserID %v", resultUserID, tt.expectedUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	httpHeader1 := http.Header{}
	httpHeader1.Set("Authorization", "Bearer auth_token")
	httpHeader2 := http.Header{}

	tests := []struct {
		name       string
		header     http.Header
		auth_token string
		funcErr    bool
		resultErr  bool
	}{
		{
			name:       "valid header",
			header:     httpHeader1,
			auth_token: "auth_token",
			funcErr:    false,
			resultErr:  false,
		},
		{
			name:       "auth_token doesnt match",
			header:     httpHeader1,
			auth_token: "diferent token",
			funcErr:    false,
			resultErr:  true,
		},
		{
			name:       "empty header",
			header:     httpHeader2,
			auth_token: "",
			funcErr:    true,
			resultErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultToken, err := GetBearerToken(tt.header)
			if (err != nil) != tt.funcErr {
				t.Errorf("ValidateJWT() error = %v, expected funcErr %v", err, tt.funcErr)
			}
			if (resultToken != tt.auth_token) != tt.resultErr {
				t.Errorf("ValidateJWT() resultUserID = %v, expectedUserID %v", resultToken, tt.auth_token)
			}
		})
	}
}
