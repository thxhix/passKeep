package token

import (
	"github.com/google/uuid"
	"time"
)

// RefreshToken represents a refresh token in memory.
//
// It includes JTI, user ID, hashed token, issuance and expiration times, and optionally the JTI of the replacing token.
type RefreshToken struct {
	JTI        uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	ReplacedBy *uuid.UUID
	CreatedAt  time.Time
}

// RefreshTokenRecord represents a refresh token as stored in the database.
//
// It mirrors RefreshToken but uses int64 for the UserID to match database user IDs.
type RefreshTokenRecord struct {
	JTI        uuid.UUID
	UserID     int64
	TokenHash  string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	ReplacedBy *uuid.UUID
}
