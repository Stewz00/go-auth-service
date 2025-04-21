package service

import (
	"context"
	"errors"
	"time"

	"github.com/Stewz00/go-auth-service/internal/model"
	"github.com/Stewz00/go-auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: 24 * time.Hour, // tokens expire after 24 hours
	}
}

// RegisterUser creates a new user account with a hashed password
func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (*model.User, error) {
	// Hash the password with a cost factor of 12 (recommended minimum)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return s.userRepo.CreateUser(ctx, email, string(hashedPassword))
}

// LoginUser authenticates a user and returns a JWT token
func (s *AuthService) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Increment failed login attempts
		if err := s.userRepo.IncrementFailedAttempts(ctx, user.ID); err != nil {
			if err == repository.ErrTooManyAttempts {
				return "", ErrAccountLocked
			}
			return "", err
		}
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(s.tokenExpiry).Unix(),
		"jti":   generateTokenID(), // Unique token ID for revocation
	})

	// Sign the token
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	// Store session in database
	claims := token.Claims.(jwt.MapClaims)
	err = s.userRepo.CreateSession(
		ctx,
		user.ID,
		claims["jti"].(string),
		time.Unix(claims["exp"].(int64), 0),
	)
	if err != nil {
		return "", err
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the user claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is revoked
	if valid, err := s.userRepo.IsSessionValid(ctx, claims["jti"].(string)); err != nil {
		return nil, err
	} else if !valid {
		return nil, errors.New("token is revoked or expired")
	}

	return claims, nil
}

// LogoutUser revokes the user's token
func (s *AuthService) LogoutUser(ctx context.Context, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	return s.userRepo.RevokeSession(ctx, claims["jti"].(string))
}

// Helper function to generate a unique token ID
func generateTokenID() string {
	// Simple implementation - in production, use a more robust method
	token, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"rand": time.Now().UnixNano(),
	}).SignedString(nil)
	return token
}
