package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Stewz00/go-auth-service/internal/database"
	"github.com/Stewz00/go-auth-service/internal/model"
	"github.com/jackc/pgx/v4"
)

// Common errors that can be returned by the repository
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrSessionNotFound = errors.New("session not found")
	ErrTooManyAttempts = errors.New("too many failed login attempts")
)

// UserRepository handles all database operations related to users.
// It provides methods for creating, retrieving, and updating user data.
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	var user model.User
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) 
		 VALUES ($1, $2) 
		 RETURNING id, email, created_at`,
		email, passwordHash).Scan(&user.ID, &user.Email, &user.Created)

	if err != nil {
		// Check for unique constraint violation
		if isPgUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, failed_login_attempts 
		 FROM users 
		 WHERE email = $1 AND is_active = true`,
		email).Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.FailedAttempts)

	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateLastLogin updates the last login time and resets failed attempts
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users 
		 SET last_login = CURRENT_TIMESTAMP, 
		     failed_login_attempts = 0 
		 WHERE id = $1`,
		userID)
	return err
}

// IncrementFailedAttempts increments the failed login attempts counter.
// If the number of failed attempts exceeds the threshold (5), the account is deactivated.
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, userID int64) error {
	var attempts int
	err := r.db.Pool.QueryRow(ctx,
		`UPDATE users 
		 SET failed_login_attempts = failed_login_attempts + 1 
		 WHERE id = $1 
		 RETURNING failed_login_attempts`,
		userID).Scan(&attempts)

	if err != nil {
		return err
	}

	// If too many failed attempts, deactivate the account
	if attempts >= 5 {
		_, err = r.db.Pool.Exec(ctx,
			`UPDATE users 
			 SET is_active = false 
			 WHERE id = $1`,
			userID)
		if err == nil {
			return ErrTooManyAttempts
		}
	}

	return err
}

// CreateSession creates a new session for a user
func (r *UserRepository) CreateSession(ctx context.Context, userID int64, tokenID string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO sessions (user_id, token_id, expires_at) 
		 VALUES ($1, $2, $3)`,
		userID, tokenID, expiresAt)
	return err
}

// RevokeSession marks a session as revoked
func (r *UserRepository) RevokeSession(ctx context.Context, tokenID string) error {
	result, err := r.db.Pool.Exec(ctx,
		`UPDATE sessions 
		 SET is_revoked = true 
		 WHERE token_id = $1`,
		tokenID)

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return err
}

// IsSessionValid checks if a session is valid and not expired
func (r *UserRepository) IsSessionValid(ctx context.Context, tokenID string) (bool, error) {
	var isRevoked bool
	var expiresAt time.Time

	err := r.db.Pool.QueryRow(ctx,
		`SELECT is_revoked, expires_at 
		 FROM sessions 
		 WHERE token_id = $1`,
		tokenID).Scan(&isRevoked, &expiresAt)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !isRevoked && time.Now().Before(expiresAt), nil
}

// Helper function to check for PostgreSQL unique constraint violations
func isPgUniqueViolation(err error) bool {
	// pgErr, ok := err.(*pgconn.PgError)
	// return ok && pgErr.Code == "23505"
	return err.Error() == "duplicate key value violates unique constraint"
}
