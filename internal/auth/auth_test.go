package auth

import (
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
	jwt2, _ := MakeJWT(uuid1, secretToken, 24*time.Hour)

	tests := []struct {
		name    string
		jwt     string
		uuid    uuid.UUID
		wantErr bool
	}{
		{
			name:    "Correct JWT",
			jwt:     jwt1,
			uuid:    uuid1,
			wantErr: false,
		},
		{
			name:    "Incorrect JWT",
			jwt:     "invalid jwt",
			uuid:    uuid1,
			wantErr: true,
		},
		{
			name:    "JWT doesn't match different uuid",
			jwt:     jwt2,
			uuid:    uuid1,
			wantErr: true,
		},
		{
			name:    "Empty JWT",
			jwt:     "",
			uuid:    uuid1,
			wantErr: true,
		},
		{
			name:    "Incorrect uuid",
			jwt:     jwt1,
			uuid:    uuid2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, err := ValidateJWT(tt.jwt, secretToken)
			if (err != nil) != tt.wantErr && uuid != tt.uuid {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
