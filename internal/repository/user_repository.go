package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Stewz00/go-auth-service/internal/database"
	"github.com/Stewz00/go-auth-service/internal/interfaces"
	"github.com/Stewz00/go-auth-service/internal/model"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Common errors that can be returned by the repository
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrSessionNotFound = errors.New("session not found")
	ErrTooManyAttempts = errors.New("too many failed login attempts")
)

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	db *database.DB
}

// Verify that UserRepositoryImpl implements UserRepository interface
var _ interfaces.UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *database.DB) interfaces.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	var user model.User
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) 
		 VALUES ($1, $2) 
		 RETURNING id, email, created_at`,
		email, passwordHash).Scan(&user.ID, &user.Email, &user.Created)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	var isActive bool
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, failed_login_attempts, is_active 
		 FROM users 
		 WHERE email = $1`,
		email).Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.FailedAttempts, &isActive)

	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if !isActive {
		return nil, ErrTooManyAttempts
	}

	return &user, nil
}

// UpdateLastLogin updates the last login time and resets failed attempts
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users 
		 SET last_login = CURRENT_TIMESTAMP, 
		     failed_login_attempts = 0 
		 WHERE id = $1`,
		userID)
	return err
}

// IncrementFailedAttempts increments the failed login attempts counter
func (r *UserRepositoryImpl) IncrementFailedAttempts(ctx context.Context, userID int64) error {
	var attempts int
	err := r.db.Pool.QueryRow(ctx,
		`UPDATE users 
		 SET failed_login_attempts = failed_login_attempts + 1,
		     is_active = CASE WHEN failed_login_attempts + 1 >= 5 THEN false ELSE true END
		 WHERE id = $1 
		 RETURNING failed_login_attempts`,
		userID).Scan(&attempts)

	if err != nil {
		return err
	}

	if attempts >= 5 {
		return ErrTooManyAttempts
	}

	return nil
}

// CreateSession creates a new session for a user
func (r *UserRepositoryImpl) CreateSession(ctx context.Context, userID int64, tokenID string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO sessions (user_id, token_id, expires_at) 
		 VALUES ($1, $2, $3)`,
		userID, tokenID, expiresAt)
	return err
}

// RevokeSession marks a session as revoked
func (r *UserRepositoryImpl) RevokeSession(ctx context.Context, tokenID string) error {
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
func (r *UserRepositoryImpl) IsSessionValid(ctx context.Context, tokenID string) (bool, error) {
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
