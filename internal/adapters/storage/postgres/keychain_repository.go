package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
)

// KeychainRepository implements persistence operations for keychain records
// using a *sql.DB Postgres driver.
type KeychainRepository struct {
	db *sql.DB
}

// NewKeychainRepository constructs a new KeychainRepository using the provided
// *sql.DB driver.
func NewKeychainRepository(db *sql.DB) *KeychainRepository {
	return &KeychainRepository{db: db}
}

// AddKey inserts a new keychain record for the given user and returns the
// generated UUID as string. `data` and `nonce` are stored as bytea in Postgres.
//
// ctx controls the database call lifetime.
func (repo *KeychainRepository) AddKey(ctx context.Context, userID int64, keyType keychain.KeyType, title string, data []byte, nonce []byte) (string, error) {
	keyUUID := uuid.New()

	query := "INSERT INTO keychain (key_uuid, user_id, type, title, data, nonce) VALUES ($1, $2, $3, $4, $5, $6)"

	_, err := repo.db.ExecContext(ctx, query, keyUUID, userID, keyType, title, data, nonce)
	if err != nil {
		return "", err
	}

	return keyUUID.String(), nil
}

// DeleteKey soft-deletes a key: sets soft_deleted = true for the given user
// and key UUID. It returns sql.ErrNoRows when no rows were affected.
func (repo *KeychainRepository) DeleteKey(ctx context.Context, userID int64, keyUUID string) error {
	query := "UPDATE keychain SET soft_deleted = true WHERE soft_deleted = false AND user_id = $1 AND key_uuid = $2"

	res, err := repo.db.ExecContext(ctx, query, userID, keyUUID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetUserKeys returns a list of the user's keys. If keyType is nil the method
// returns all types; if non-nil the repository filters by the given type.
func (repo *KeychainRepository) GetUserKeys(ctx context.Context, userID int64, keyType *string) (keys []*keychain.KeyRecord, err error) {
	query := `
		SELECT id, key_uuid, user_id, type, title, created_at, updated_at
		FROM keychain
		WHERE soft_deleted = false
		AND user_id = $1
		AND ($2::text IS NULL OR type = $2)
		ORDER BY created_at DESC
	`

	var argKeyType any
	if keyType == nil {
		argKeyType = nil
	} else {
		argKeyType = *keyType
	}

	rows, err := repo.db.QueryContext(ctx, query, userID, argKeyType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		row := &keychain.KeyRecord{}
		err = rows.Scan(&row.ID, &row.KeyUUID, &row.UserID, &row.KeyType, &row.Title, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, err
		}

		keys = append(keys, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return keys, err
}

// GetUserKey returns the full key record (including encrypted data and nonce)
// for the given key UUID and user. If the record is not found repository should
// return sql.ErrNoRows.
func (repo *KeychainRepository) GetUserKey(ctx context.Context, userID int64, keyUUID string) (*keychain.KeyRecord, error) {
	var kr keychain.KeyRecord

	query := `SELECT id, key_uuid, user_id, type, title, data, nonce, created_at, updated_at FROM keychain WHERE soft_deleted = false AND key_uuid = $1 AND user_id = $2`

	if err := repo.db.QueryRowContext(ctx, query, keyUUID, userID).Scan(&kr.ID, &kr.KeyUUID, &kr.UserID, &kr.KeyType, &kr.Title, &kr.Data, &kr.Nonce, &kr.CreatedAt, &kr.UpdatedAt); err != nil {
		return nil, err
	}
	return &kr, nil
}
