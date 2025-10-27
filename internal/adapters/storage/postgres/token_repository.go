package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"time"
)

type TokensRepository struct {
	db *sql.DB
}

func NewTokensRepository(db *sql.DB) *TokensRepository {
	return &TokensRepository{db: db}
}

var (
	queryInsert = `INSERT INTO auth_refresh_tokens (user_id, jti, token_hash, issued_at, expires_at) VALUES ($1, $2, $3, $4, $5);`
)

func (repo *TokensRepository) Create(ctx context.Context, userID int64, jti uuid.UUID, tokenHash string, issuedAt time.Time, expiresAt time.Time) error {
	_, err := repo.db.ExecContext(ctx, queryInsert, userID, jti, tokenHash, issuedAt, expiresAt)
	return err
}

func (r *TokensRepository) Rotate(ctx context.Context, userID int64, oldJTI uuid.UUID, newJTI uuid.UUID, newHash string, newIssuedAt time.Time, newExpiresAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	const queryUpdate = `UPDATE auth_refresh_tokens SET replaced_by = $1 WHERE jti = $2 AND user_id = $3 AND replaced_by IS NULL AND expires_at > NOW();`

	res, err := tx.ExecContext(ctx, queryUpdate, newJTI, oldJTI, userID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff != 1 {
		return token.ErrTokenAlreadyRotatedOrExpired
	}

	if _, err := tx.ExecContext(ctx, queryInsert, userID, newJTI, newHash, newIssuedAt, newExpiresAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TokensRepository) GetByJTI(ctx context.Context, jti uuid.UUID) (*token.RefreshTokenRecord, error) {
	var rt token.RefreshTokenRecord

	query := `SELECT jti, user_id, token_hash, issued_at, expires_at, replaced_by FROM auth_refresh_tokens WHERE jti = $1`

	if err := r.db.QueryRowContext(ctx, query, jti).Scan(
		&rt.JTI,
		&rt.UserID,
		&rt.TokenHash,
		&rt.IssuedAt,
		&rt.ExpiresAt,
		&rt.ReplacedBy,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, token.ErrTokenDoesntExistsByJTI
		}
		return nil, err
	}
	return &rt, nil
}
