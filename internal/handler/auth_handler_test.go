package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Stewz00/go-auth-service/internal/service"
	"github.com/Stewz00/go-auth-service/internal/test"
)

func TestAuthHandler_Register(t *testing.T) {
	mockRepo := test.NewMockUserRepository()
	authService := service.NewAuthService(mockRepo, "test-secret")
	handler := NewAuthHandler(authService)

	tests := []struct {
		name           string
		requestBody    map[string]string
		wantStatusCode int
		wantErr        bool
	}{
		{
			name: "valid registration",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			wantStatusCode: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "invalid email",
			requestBody: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErr:        true,
		},
		// TODO: Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Register(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status %v, want %v", w.Code, tt.wantStatusCode)
			}

			var response map[string]string
			json.NewDecoder(w.Body).Decode(&response)

			if tt.wantErr {
				if response["error"] == "" {
					t.Error("expected error message but got none")
				}
			} else {
				if response["email"] != tt.requestBody["email"] {
					t.Errorf("got email %v, want %v", response["email"], tt.requestBody["email"])
				}
			}
		})
	}
}

// TODO: Add tests for Login and Logout handlers
