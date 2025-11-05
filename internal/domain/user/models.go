package user

import "time"

// User represents a system user.
type User struct {
	ID        int64
	Login     string
	Password  string
	CreatedAt time.Time
}

// UserRecord represents the database record for a user.
type UserRecord struct {
	ID           int64
	Login        string
	PasswordHash string
}
