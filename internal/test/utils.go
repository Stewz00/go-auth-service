package test

import (
	"context"
	"errors"
	"time"

	"github.com/Stewz00/go-auth-service/internal/model"
)

// MockDB implements a mock database for testing
type MockDB struct {
	users    map[string]*model.User
	sessions map[string]bool
}

func NewMockDB() *MockDB {
	return &MockDB{
		users:    make(map[string]*model.User),
		sessions: make(map[string]bool),
	}
}

// Mock implementation of UserRepository for testing
type MockUserRepository struct {
	db *MockDB
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		db: NewMockDB(),
	}
}

func (r *MockUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	if _, exists := r.db.users[email]; exists {
		return nil, ErrDuplicateEmail
	}

	user := &model.User{
		ID:       int64(len(r.db.users) + 1),
		Email:    email,
		Password: passwordHash,
		Created:  time.Now(),
	}
	r.db.users[email] = user
	return user, nil
}

func (r *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, exists := r.db.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *MockUserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	return nil
}

func (r *MockUserRepository) IncrementFailedAttempts(ctx context.Context, userID int64) error {
	return nil
}

func (r *MockUserRepository) CreateSession(ctx context.Context, userID int64, tokenID string, expiresAt time.Time) error {
	r.db.sessions[tokenID] = true
	return nil
}

func (r *MockUserRepository) RevokeSession(ctx context.Context, tokenID string) error {
	if _, exists := r.db.sessions[tokenID]; !exists {
		return ErrSessionNotFound
	}
	r.db.sessions[tokenID] = false
	return nil
}

func (r *MockUserRepository) IsSessionValid(ctx context.Context, tokenID string) (bool, error) {
	valid, exists := r.db.sessions[tokenID]
	if !exists {
		return false, nil
	}
	return valid, nil
}

// Error definitions for mocks
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrSessionNotFound = errors.New("session not found")
)
