package keychain

import "context"

// KeychainRepository defines the interface for managing user keys.
//
// It abstracts CRUD operations for different types of keys stored in the system.
type KeychainRepository interface {
	// GetUserKeys retrieves all keys for a given user.
	//
	// If keyType is not nil, the results are filtered by the specified key type.
	// Returns a slice of KeyRecord pointers or an error if something went wrong.
	GetUserKeys(ctx context.Context, userID int64, keyType *string) ([]*KeyRecord, error)

	// GetUserKey retrieves a single key by its UUID for a given user.
	//
	// Returns the KeyRecord or an error if the key is not found.
	GetUserKey(ctx context.Context, userID int64, keyUUID string) (*KeyRecord, error)

	// AddKey creates a new key for the user.
	//
	// keyType specifies the type of the key (credential, text, file, or card).
	// title is a human-readable name for the key.
	// data and nonce contain the encrypted key data and nonce for AEAD encryption.
	// Returns the UUID of the created key as a string, or an error if creation failed.
	AddKey(ctx context.Context, userID int64, keyType KeyType, title string, data []byte, nonce []byte) (string, error)

	// DeleteKey removes a key by its UUID for a given user.
	//
	// Returns an error if the key does not exist or deletion failed.
	DeleteKey(ctx context.Context, userID int64, keyUUID string) error
}
