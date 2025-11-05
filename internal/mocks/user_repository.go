package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/user"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Create(ctx context.Context, login string, passwordHash string) (int64, error) {
	args := m.Called(ctx, login, passwordHash)
	return int64(args.Int(0)), args.Error(1)
}

func (m *UserRepositoryMock) GetByLogin(ctx context.Context, login string) (*user.UserRecord, error) {
	args := m.Called(ctx, login)

	return args.Get(0).(*user.UserRecord), args.Error(1)
}
