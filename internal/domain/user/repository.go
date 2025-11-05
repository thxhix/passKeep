package user

import "context"

// UserRepository defines the interface for user storage operations.
//
// It supports creating a new user and retrieving a user by login.
type UserRepository interface {
	// Create stores a new user with the given login and password hash.
	// Returns the generated user ID or an error if creation fails.
	Create(ctx context.Context, login string, passwordHash string) (int64, error)

	// GetByLogin retrieves a user by their login.
	// Returns a UserRecord or nil if no user is found.
	GetByLogin(ctx context.Context, login string) (*UserRecord, error)
}
