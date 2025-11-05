package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
)

type KeychainServiceMock struct {
	mock.Mock
}

func (m *KeychainServiceMock) GetKeys(ctx context.Context, userID int64, keyType *keychain.KeyType) (list []*keychain.KeyRecord, err error) {
	args := m.Called(ctx, userID, keyType)
	return args.Get(0).([]*keychain.KeyRecord), args.Error(1)
}

func (m *KeychainServiceMock) GetKey(ctx context.Context, userID int64, keyUUID string) (record *keychain.KeyRecord, decryptedData []byte, err error) {
	args := m.Called(ctx, userID, keyUUID)
	return args.Get(0).(*keychain.KeyRecord), args.Get(1).([]byte), args.Error(2)
}

func (m *KeychainServiceMock) DeleteKey(ctx context.Context, userID int64, keyUUID string) error {
	args := m.Called(ctx, userID, keyUUID)
	return args.Error(0)
}

func (m *KeychainServiceMock) AddCredential(ctx context.Context, userID int64, in dto.AddCredentialsDTO) (string, error) {
	args := m.Called(ctx, userID, in)
	return args.String(0), args.Error(1)
}

func (m *KeychainServiceMock) AddCard(ctx context.Context, userID int64, in dto.AddCardDTO) (string, error) {
	args := m.Called(ctx, userID, in)
	return args.String(0), args.Error(1)
}

func (m *KeychainServiceMock) AddText(ctx context.Context, userID int64, in dto.AddTextDTO) (string, error) {
	args := m.Called(ctx, userID, in)
	return args.String(0), args.Error(1)
}

func (m *KeychainServiceMock) AddFile(ctx context.Context, userID int64, in dto.AddFileDTO) (string, error) {
	args := m.Called(ctx, userID, in)
	return args.String(0), args.Error(1)
}
