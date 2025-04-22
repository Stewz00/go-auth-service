package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Stewz00/go-auth-service/internal/database"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("../../.env.test"); err != nil {
		fmt.Printf("Warning: .env.test file not found: %v\n", err)
	}
}

func setupTestDB(t *testing.T) *database.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := database.New(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up before each test
	_, err = db.Pool.Exec(context.Background(), "TRUNCATE users, sessions CASCADE")
	if err != nil {
		t.Fatalf("Failed to clean test database: %v", err)
	}

	return db
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
		errIs    error
	}{
		{
			name:     "valid user creation",
			email:    "test@example.com",
			password: "hashedpassword",
			wantErr:  false,
		},
		{
			name:     "duplicate email",
			email:    "test@example.com",
			password: "hashedpassword",
			wantErr:  true,
			errIs:    ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.CreateUser(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != tt.errIs {
					t.Errorf("got error %v, want %v", err, tt.errIs)
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
				t.Errorf("got email %v, want %v", user.Email, tt.email)
			}
		})
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	email := "test@example.com"
	password := "hashedpassword"
	_, err := repo.CreateUser(ctx, email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "existing user",
			email:   email,
			wantErr: nil,
		},
		{
			name:    "non-existent user",
			email:   "nonexistent@example.com",
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetUserByEmail(ctx, tt.email)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error but got none")
				}
				if err != tt.wantErr {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
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
				t.Errorf("got email %v, want %v", user.Email, tt.email)
			}
		})
	}
}

func TestUserRepository_IncrementFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	email := "test@example.com"
	password := "hashedpassword"
	user, err := repo.CreateUser(ctx, email, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name      string
		attempts  int
		wantError error
	}{
		{
			name:      "first attempt",
			attempts:  1,
			wantError: nil,
		},
		{
			name:      "fourth attempt",
			attempts:  4,
			wantError: nil,
		},
		{
			name:      "fifth attempt - should lock",
			attempts:  5,
			wantError: ErrTooManyAttempts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastError error
			for i := 0; i < tt.attempts; i++ {
				lastError = repo.IncrementFailedAttempts(ctx, user.ID)
			}

			if tt.wantError != nil {
				if lastError == nil {
					t.Error("expected error but got none")
				}
				if lastError != tt.wantError {
					t.Errorf("got error %v, want %v", lastError, tt.wantError)
				}
			}
		})
	}
}

func TestUserRepository_SessionManagement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.CreateUser(ctx, "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Test CreateSession
	t.Run("create session", func(t *testing.T) {
		tokenID := "test-token"
		expiresAt := time.Now().Add(24 * time.Hour)

		err := repo.CreateSession(ctx, user.ID, tokenID, expiresAt)
		if err != nil {
			t.Errorf("failed to create session: %v", err)
		}

		// Verify session is valid
		valid, err := repo.IsSessionValid(ctx, tokenID)
		if err != nil {
			t.Errorf("failed to check session validity: %v", err)
		}
		if !valid {
			t.Error("expected session to be valid")
		}
	})

	// Test RevokeSession
	t.Run("revoke session", func(t *testing.T) {
		tokenID := "test-token-2"
		expiresAt := time.Now().Add(24 * time.Hour)

		// Create and then revoke session
		err := repo.CreateSession(ctx, user.ID, tokenID, expiresAt)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		err = repo.RevokeSession(ctx, tokenID)
		if err != nil {
			t.Errorf("failed to revoke session: %v", err)
		}

		// Verify session is invalid
		valid, err := repo.IsSessionValid(ctx, tokenID)
		if err != nil {
			t.Errorf("failed to check session validity: %v", err)
		}
		if valid {
			t.Error("expected session to be invalid after revocation")
		}
	})

	// Test IsSessionValid for expired sessions
	t.Run("expired session", func(t *testing.T) {
		tokenID := "test-token-3"
		expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

		err := repo.CreateSession(ctx, user.ID, tokenID, expiresAt)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		valid, err := repo.IsSessionValid(ctx, tokenID)
		if err != nil {
			t.Errorf("failed to check session validity: %v", err)
		}
		if valid {
			t.Error("expected expired session to be invalid")
		}
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user, err := repo.CreateUser(ctx, "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Test updating last login
	err = repo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		t.Errorf("failed to update last login: %v", err)
	}

	// Verify the update by retrieving the user
	updatedUser, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("failed to get updated user: %v", err)
	}

	if updatedUser.FailedAttempts != 0 {
		t.Errorf("expected failed attempts to be reset to 0, got %d", updatedUser.FailedAttempts)
	}
}
