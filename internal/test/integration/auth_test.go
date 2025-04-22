package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Stewz00/go-auth-service/internal/config"
	"github.com/Stewz00/go-auth-service/internal/database"
	"github.com/Stewz00/go-auth-service/internal/handler"
	"github.com/Stewz00/go-auth-service/internal/repository"
	"github.com/Stewz00/go-auth-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

var (
	testDB     *database.DB
	testRouter *chi.Mux
)

func TestMain(m *testing.M) {
	// Set up test environment
	if err := godotenv.Load("../../../.env.test"); err != nil {
		fmt.Printf("Warning: .env.test file not found: %v\n", err)
	}

	// Set required environment variables if not already set
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8081")
	}
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:***REMOVED***@localhost:5432/authdb_test?sslmode=disable")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize test database
	testDB, err = database.New(cfg.DbURL)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Set up router and handlers
	testRouter = setupTestRouter(testDB, cfg.JwtSecret)

	// Run tests
	code := m.Run()

	// Clean up
	testDB.Close()
	os.Exit(code)
}

func setupTestRouter(db *database.DB, jwtSecret string) *chi.Mux {
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/logout", authHandler.Logout)

	return r
}

func TestRegisterLoginLogoutFlow(t *testing.T) {
	cleanup(t) // Clean up before test

	// Test user data
	user := map[string]string{
		"email":    "integration@test.com",
		"password": "testpassword123",
	}

	var token string

	// 1. Test Registration
	t.Run("register", func(t *testing.T) {
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	// 2. Test Login
	t.Run("login", func(t *testing.T) {
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		token = response.Token

		if token == "" {
			t.Error("expected token in response, got empty string")
		}
	})

	// 3. Test Logout
	t.Run("logout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// 4. Verify can't use token after logout
	t.Run("verify-token-invalid-after-logout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// After logout, using the token should result in Unauthorized (401)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

// TestFailedLoginAttempts tests the account locking mechanism
func TestFailedLoginAttempts(t *testing.T) {
	cleanup(t) // Clean up before test

	// Register test user
	user := map[string]string{
		"email":    "lockout@test.com",
		"password": "testpassword123",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to register test user: got status %d", w.Code)
	}

	// Attempt multiple failed logins
	wrongPassword := map[string]string{
		"email":    "lockout@test.com",
		"password": "wrongpassword",
	}

	for i := 0; i < 5; i++ { // Changed from 6 to 5 attempts
		body, _ := json.Marshal(wrongPassword)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		if i < 4 { // First 4 attempts should fail with unauthorized
			if w.Code != http.StatusUnauthorized {
				t.Errorf("attempt %d: expected status %d, got %d", i+1, http.StatusUnauthorized, w.Code)
			}
		} else { // 5th attempt should trigger the lock
			if w.Code != http.StatusForbidden {
				t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
			}
		}
	}

	// Try logging in with correct password after lockout
	body, _ = json.Marshal(user)
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// Account should be locked, returning Forbidden (403)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	// Verify the error message
	var response struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Error != "Account is locked due to too many failed attempts" {
		t.Errorf("unexpected error message: %s", response.Error)
	}
}

// Helper function to clean up test data
func cleanup(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.Pool.Exec(ctx, "TRUNCATE users, sessions CASCADE")
	if err != nil {
		t.Errorf("failed to clean up test data: %v", err)
	}
}
