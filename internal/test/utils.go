package test

import (
	"context"
	"time"

	"github.com/Stewz00/go-auth-service/internal/interfaces"
	"github.com/Stewz00/go-auth-service/internal/model"
	"github.com/Stewz00/go-auth-service/internal/repository"
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

// MockUserRepository implements the repository.UserRepository interface
type MockUserRepository struct {
	db *MockDB
}

// Verify that MockUserRepository implements UserRepository interface
var _ interfaces.UserRepository = (*MockUserRepository)(nil)

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		db: NewMockDB(),
	}
}

// CreateUser mocks creating a new user
func (r *MockUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	if _, exists := r.db.users[email]; exists {
		return nil, repository.ErrDuplicateEmail
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

// GetUserByEmail mocks retrieving a user by email
func (r *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, exists := r.db.users[email]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

// UpdateLastLogin mocks updating the last login time
func (r *MockUserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	return nil
}

// IncrementFailedAttempts mocks incrementing failed login attempts
func (r *MockUserRepository) IncrementFailedAttempts(ctx context.Context, userID int64) error {
	return nil
}

// CreateSession mocks creating a new session
func (r *MockUserRepository) CreateSession(ctx context.Context, userID int64, tokenID string, expiresAt time.Time) error {
	r.db.sessions[tokenID] = true
	return nil
}

// RevokeSession mocks revoking a session
func (r *MockUserRepository) RevokeSession(ctx context.Context, tokenID string) error {
	if _, exists := r.db.sessions[tokenID]; !exists {
		return repository.ErrSessionNotFound
	}
	r.db.sessions[tokenID] = false
	return nil
}

// IsSessionValid mocks checking if a session is valid
func (r *MockUserRepository) IsSessionValid(ctx context.Context, tokenID string) (bool, error) {
	valid, exists := r.db.sessions[tokenID]
	if !exists {
		return false, nil
	}
	return valid, nil
}
