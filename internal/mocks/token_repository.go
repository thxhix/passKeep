package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"time"
)

type TokenRepositoryMock struct {
	mock.Mock
}

func (m *TokenRepositoryMock) Create(ctx context.Context, userID int64, jti uuid.UUID, tokenHash string, issuedAt time.Time, expiresAt time.Time) error {
	args := m.Called(ctx, userID, jti, tokenHash, issuedAt, expiresAt)
	return args.Error(0)
}

func (m *TokenRepositoryMock) Rotate(ctx context.Context, userID int64, oldJTI uuid.UUID, newJTI uuid.UUID, newHash string, newIssuedAt time.Time, newExpiresAt time.Time) error {
	args := m.Called(ctx, userID, oldJTI, newJTI, newHash, newIssuedAt, newExpiresAt)
	return args.Error(0)
}

func (m *TokenRepositoryMock) GetByJTI(ctx context.Context, jti uuid.UUID) (*token.RefreshTokenRecord, error) {
	args := m.Called(ctx, jti)
	return args.Get(0).(*token.RefreshTokenRecord), args.Error(1)
}
