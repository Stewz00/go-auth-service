package interfaces

import (
	"context"
	"time"

	"github.com/Stewz00/go-auth-service/internal/model"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateLastLogin(ctx context.Context, userID int64) error
	IncrementFailedAttempts(ctx context.Context, userID int64) error
	CreateSession(ctx context.Context, userID int64, tokenID string, expiresAt time.Time) error
	RevokeSession(ctx context.Context, tokenID string) error
	IsSessionValid(ctx context.Context, tokenID string) (bool, error)
}
