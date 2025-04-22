package service

import (
	"context"
	"testing"
	"time"

	"github.com/Stewz00/go-auth-service/internal/test"
	"github.com/golang-jwt/jwt/v5"
)

func TestRegisterUser(t *testing.T) {
	mockRepo := test.NewMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	tests := []struct {
		name        string
		email       string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid registration",
			email:    "test@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:        "duplicate email",
			email:       "test@example.com",
			password:    "password123",
			wantErr:     true,
			errContains: "email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.RegisterUser(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if user == nil {
				t.Error("expected user but got nil")
			}
			if user != nil && user.Email != tt.email {
				t.Errorf("got email %q, want %q", user.Email, tt.email)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	mockRepo := test.NewMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	// Register a test user first
	email := "test@example.com"
	password := "password123"
	_, err := authService.RegisterUser(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid login",
			email:    email,
			password: password,
			wantErr:  false,
		},
		{
			name:        "invalid password",
			email:       email,
			password:    "wrongpassword",
			wantErr:     true,
			errContains: "invalid email or password",
		},
		{
			name:        "non-existent user",
			email:       "nonexistent@example.com",
			password:    password,
			wantErr:     true,
			errContains: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.LoginUser(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if token == "" {
				t.Error("expected token but got empty string")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockRepo := test.NewMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	// Create and login a test user to get a valid token
	email := "test@example.com"
	password := "password123"
	_, err := authService.RegisterUser(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	validToken, err := authService.LoginUser(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to login test user: %v", err)
	}

	// Create an expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"sub": 1,
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte("test-secret"))

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:        "expired token",
			token:       expiredTokenString,
			wantErr:     true,
			errContains: "expired",
		},
		{
			name:        "invalid token",
			token:       "invalid.token.string",
			wantErr:     true,
			errContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(context.Background(), tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if claims == nil {
				t.Error("expected claims but got nil")
			}
		})
	}
}

func TestLogoutUser(t *testing.T) {
	mockRepo := test.NewMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	// Create and login a test user
	email := "test@example.com"
	password := "password123"
	_, err := authService.RegisterUser(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	validToken, err := authService.LoginUser(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to login test user: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid logout",
			token:   validToken,
			wantErr: false,
		},
		{
			name:        "invalid token",
			token:       "invalid.token.string",
			wantErr:     true,
			errContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.LogoutUser(context.Background(), tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify token is invalid after logout
			if !tt.wantErr {
				_, err = authService.ValidateToken(context.Background(), tt.token)
				if err == nil {
					t.Error("expected token to be invalid after logout")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
