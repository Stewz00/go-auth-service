package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/Stewz00/go-auth-service/internal/repository"
	"github.com/Stewz00/go-auth-service/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// Validate checks if an email is valid
func isValidEmail(email string) bool {
	// Simple regex for email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		sendJSONError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		sendJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		sendJSONError(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	user, err := h.authService.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrInvalidCredentials {
			code = http.StatusBadRequest
		}
		sendJSONError(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully", "email": user.Email})
}

// Login handles user authentication and returns a JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			sendJSONError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		case service.ErrAccountLocked, repository.ErrTooManyAttempts:
			sendJSONError(w, "Account is locked due to too many failed attempts", http.StatusForbidden)
			return
		default:
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}

// Logout handles user logout by revoking the JWT token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		sendJSONError(w, "No token provided", http.StatusUnauthorized)
		return
	}

	if err := h.authService.LogoutUser(r.Context(), token); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrInvalidToken {
			code = http.StatusUnauthorized
		}
		sendJSONError(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// Helper function to extract JWT token from Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// Helper function to send JSON error responses
func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(AuthResponse{Error: message})
}
