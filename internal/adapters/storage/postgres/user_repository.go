package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"github.com/thxhix/passKeeper/internal/domain/user"
)

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (repo *UsersRepository) Create(ctx context.Context, login string, passwordHash string) (int64, error) {
	var id int64

	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	err := repo.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return 0, user.ErrDuplicateLogin
			}
		}
		return 0, err
	}

	return id, nil
}

func (repo *UsersRepository) GetByLogin(ctx context.Context, login string) (*user.UserRecord, error) {
	var au user.UserRecord

	query := `SELECT id, login, password FROM users WHERE login = $1`

	if err := repo.db.QueryRowContext(ctx, query, login).Scan(&au.ID, &au.Login, &au.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &au, nil
}
