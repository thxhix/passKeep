package token

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// TokenRepository defines the interface for managing refresh tokens in storage.
//
// It supports creating, rotating, and retrieving refresh tokens by JTI.
type TokenRepository interface {
	// Create inserts a new refresh token for a given user.
	//
	// Parameters:
	//  - ctx: context for cancellation and deadlines
	//  - userID: ID of the user the token belongs to
	//  - jti: unique token identifier (JTI)
	//  - tokenHash: hashed token value to store securely
	//  - issuedAt: token issuance time
	//  - expiresAt: token expiration time
	//
	// Returns an error if the operation fails.
	Create(ctx context.Context, userID int64, jti uuid.UUID, tokenHash string, issuedAt time.Time, expiresAt time.Time) error

	// Rotate replaces an old token with a new one.
	//
	// Parameters:
	//  - ctx: context for cancellation and deadlines
	//  - userID: ID of the user
	//  - oldJTI: JTI of the old token to be replaced
	//  - newJTI: JTI of the new token
	//  - newHash: hashed value of the new token
	//  - newIssuedAt: issuance time of the new token
	//  - newExpiresAt: expiration time of the new token
	//
	// Returns an error if rotation fails.
	Rotate(ctx context.Context, userID int64, oldJTI uuid.UUID, newJTI uuid.UUID, newHash string, newIssuedAt time.Time, newExpiresAt time.Time) error

	// GetByJTI retrieves a refresh token record by its JTI.
	//
	// Returns the RefreshTokenRecord or an error if not found.
	GetByJTI(ctx context.Context, jti uuid.UUID) (*RefreshTokenRecord, error)
}
