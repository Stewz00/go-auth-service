package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool using the provided connection URL
// It implements connection pooling and handles reconnection automatically
func New(dbURL string) (*DB, error) {
	// Create a connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing database URL: %v", err)
	}

	// Set some reasonable pool limits
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	// Create the connection pool
	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Verify the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
