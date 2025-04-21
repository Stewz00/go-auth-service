package model

import "time"

type User struct {
	ID             int64
	Email          string
	Password       string // hashed
	Created        time.Time
	FailedAttempts int64
}
