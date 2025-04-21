package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	JwtSecret string
	DbURL     string
}

// Load reads the configuration from a .env file or environment variables and returns a Config struct.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	// Try to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	dbURL := os.Getenv("DATABASE_URL")

	// Optional: validate required variables
	if port == "" || jwtSecret == "" || dbURL == "" {
		return nil, fmt.Errorf("missing required environment variables: PORT=%q, JWT_SECRET=%q, DATABASE_URL=%q", port, jwtSecret, dbURL)
	}

	cfg := &Config{
		Port:      port,
		JwtSecret: jwtSecret,
		DbURL:     dbURL,
	}
	return cfg, nil
}
