package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) Register(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.String(1), args.String(2), args.Error(3)
}

func (m *AuthServiceMock) Login(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.String(1), args.String(2), args.Error(3)
}

func (m *AuthServiceMock) Refresh(ctx context.Context, incomingRefreshToken string) (accessToken string, refreshToken string, err error) {
	args := m.Called(ctx, incomingRefreshToken)
	return args.String(0), args.String(1), args.Error(2)
}
