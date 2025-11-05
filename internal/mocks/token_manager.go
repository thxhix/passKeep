package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"time"
)

type TokenManagerMock struct {
	mock.Mock
}

func (m *TokenManagerMock) GenerateRefreshToken(userID int64) (token string, jti uuid.UUID, ttl time.Duration, err error) {
	args := m.Called(userID)
	return args.String(0), args.Get(1).(uuid.UUID), args.Get(2).(time.Duration), args.Error(3)
}

func (m *TokenManagerMock) ParseRefreshToken(tokenStr string) (userID, jti string, err error) {
	args := m.Called(tokenStr)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *TokenManagerMock) GenerateAccessToken(userID int64) (string, error) {
	args := m.Called(userID)
	return args.Get(0).(string), args.Error(1)
}

func (m *TokenManagerMock) Sha256Hex(s string) string {
	args := m.Called(s)
	return args.Get(0).(string)
}
