package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
)

type KeychainRepositoryMock struct {
	mock.Mock
}

func (m *KeychainRepositoryMock) GetUserKeys(ctx context.Context, userID int64, keyType *string) ([]*keychain.KeyRecord, error) {
	args := m.Called(ctx, userID, keyType)
	return args.Get(0).([]*keychain.KeyRecord), args.Error(1)
}

func (m *KeychainRepositoryMock) GetUserKey(ctx context.Context, userID int64, keyUUID string) (*keychain.KeyRecord, error) {
	args := m.Called(ctx, userID, keyUUID)
	return args.Get(0).(*keychain.KeyRecord), args.Error(1)
}

func (m *KeychainRepositoryMock) AddKey(ctx context.Context, userID int64, keyType keychain.KeyType, title string, data []byte, nonce []byte) (string, error) {
	args := m.Called(ctx, userID, keyType, title, data, nonce)
	return args.String(0), args.Error(1)
}

func (m *KeychainRepositoryMock) DeleteKey(ctx context.Context, userID int64, keyUUID string) error {
	args := m.Called(ctx, userID, keyUUID)
	return args.Error(0)
}
