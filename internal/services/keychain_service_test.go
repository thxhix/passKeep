package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/mocks"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"testing"
	"time"
)

func TestKeychainService_GetKeys_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := []*keychain.KeyRecord{
		&keychain.KeyRecord{
			ID:        1,
			KeyUUID:   uuid.New(),
			UserID:    1,
			KeyType:   "",
			Title:     "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockKeychainRepo.On("GetUserKeys", ctx, int64(1), mock.Anything).Return(retObj, nil)

	list, err := s.GetKeys(ctx, 1, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, list)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_GetKeys_Empty(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := []*keychain.KeyRecord{}

	mockKeychainRepo.On("GetUserKeys", ctx, int64(1), mock.Anything).Return(retObj, nil)

	list, err := s.GetKeys(ctx, 1, nil)

	assert.NoError(t, err)
	assert.Empty(t, list)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_GetKeys_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := []*keychain.KeyRecord{}

	mockKeychainRepo.On("GetUserKeys", ctx, int64(1), mock.Anything).Return(retObj, errors.New("some error"))

	list, err := s.GetKeys(ctx, 1, nil)

	assert.Error(t, err)
	assert.Empty(t, list)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_GetKey_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := &keychain.KeyRecord{
		ID:        1,
		KeyUUID:   uuid.New(),
		UserID:    1,
		KeyType:   "",
		Title:     "test",
		Data:      []byte{1, 2, 3},
		Nonce:     []byte{4, 5, 6},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockKeychainRepo.On("GetUserKey", ctx, int64(1), mock.Anything).Return(retObj, nil)
	mockCryptManager.On("Decrypt", retObj.Nonce, retObj.Data).Return(retObj.Data, nil)

	record, _, err := s.GetKey(ctx, 1, "12345")

	assert.NoError(t, err)
	assert.IsType(t, keychain.KeyRecord{}, *record)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_GetKey_Decrypt_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := &keychain.KeyRecord{
		ID:        1,
		KeyUUID:   uuid.New(),
		UserID:    1,
		KeyType:   "",
		Title:     "test",
		Data:      []byte{1, 2, 3},
		Nonce:     []byte{4, 5, 6},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockKeychainRepo.On("GetUserKey", ctx, int64(1), mock.Anything).Return(retObj, nil)
	mockCryptManager.On("Decrypt", retObj.Nonce, retObj.Data).Return([]byte{}, errors.New("some error"))

	_, _, err := s.GetKey(ctx, 1, "12345")

	assert.Error(t, err)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_GetKey_Get_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	retObj := &keychain.KeyRecord{
		ID:        1,
		KeyUUID:   uuid.New(),
		UserID:    1,
		KeyType:   "",
		Title:     "test",
		Data:      []byte{1, 2, 3},
		Nonce:     []byte{4, 5, 6},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockKeychainRepo.On("GetUserKey", ctx, int64(1), mock.Anything).Return(retObj, errors.New("some error"))

	_, _, err := s.GetKey(ctx, 1, "12345")

	assert.Error(t, err)
	mockKeychainRepo.AssertExpectations(t)
}

func TestKeychainService_DeleteKey_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockKeychainRepo.On("DeleteKey", ctx, int64(1), mock.Anything).Return(nil)

	err := s.DeleteKey(ctx, 1, "12345")

	assert.NoError(t, err)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_DeleteKey_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockKeychainRepo.On("DeleteKey", ctx, int64(1), mock.Anything).Return(errors.New("some error"))

	err := s.DeleteKey(ctx, 1, "12345")

	assert.Error(t, err)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCredential_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("credential"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("1", nil)

	in := dto.AddCredentialsDTO{
		Title:    "Title",
		Login:    "test",
		Password: "test",
		Site:     "test",
		Note:     "test",
	}

	id, err := s.AddCredential(ctx, 1, in)

	assert.NoError(t, err)
	assert.Equal(t, "1", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCredential_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("credential"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("", errors.New("some error"))

	in := dto.AddCredentialsDTO{
		Title:    "Title",
		Login:    "test",
		Password: "test",
		Site:     "test",
		Note:     "test",
	}

	id, err := s.AddCredential(ctx, 1, in)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCredential_Validate_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	in := dto.AddCredentialsDTO{
		Password: "test",
		Site:     "test",
		Note:     "test",
	}

	id, err := s.AddCredential(ctx, 1, in)

	assert.Error(t, err)
	assert.Empty(t, id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCard_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("card"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("1", nil)

	in := dto.AddCardDTO{
		Title:   "Title",
		Number:  "4716532755237178",
		ExpDate: "ExpDate",
		CVV:     "CVV",
		Holder:  "Holder",
		Bank:    "Bank",
		Note:    "test",
	}

	id, err := s.AddCard(ctx, 1, in)

	assert.NoError(t, err)
	assert.Equal(t, "1", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCard_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("card"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("", errors.New("some error"))

	in := dto.AddCardDTO{
		Title:   "Title",
		Number:  "4716532755237178",
		ExpDate: "ExpDate",
		CVV:     "CVV",
		Holder:  "Holder",
		Bank:    "Bank",
		Note:    "test",
	}

	id, err := s.AddCard(ctx, 1, in)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddCard_Validate_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	in := dto.AddCardDTO{
		Number:  "1234567890",
		ExpDate: "ExpDate",
		Holder:  "Holder",
		Bank:    "Bank",
		Note:    "test",
	}

	id, err := s.AddCard(ctx, 1, in)

	assert.Error(t, err)
	assert.Empty(t, id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddText_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("text"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("1", nil)

	in := dto.AddTextDTO{
		Title: "Title",
		Text:  "some text",
		Note:  "test",
	}

	id, err := s.AddText(ctx, 1, in)

	assert.NoError(t, err)
	assert.Equal(t, "1", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddText_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("text"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("", errors.New("some error"))

	in := dto.AddTextDTO{
		Title: "Title",
		Text:  "some text",
		Note:  "test",
	}

	id, err := s.AddText(ctx, 1, in)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddText_Validate_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	in := dto.AddTextDTO{
		Note: "test",
	}

	id, err := s.AddText(ctx, 1, in)

	assert.Error(t, err)
	assert.Empty(t, id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddFile_Success(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("file"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("1", nil)

	in := dto.AddFileDTO{
		Title: "Title",
		File:  make([]byte, 0),
		Note:  "test",
	}

	id, err := s.AddFile(ctx, 1, in)

	assert.NoError(t, err)
	assert.Equal(t, "1", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddFile_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	mockCryptManager.On("Encrypt", mock.Anything).Return([]byte{1, 2, 3}, []byte{4, 5, 6}, nil)

	var kt keychain.KeyType
	if tkt, ok := keychain.ParseKeyType("file"); ok {
		kt = tkt
	} else {
		t.Fatal("unknown key type")
	}

	mockKeychainRepo.On(
		"AddKey",
		ctx,
		int64(1),
		kt,
		"Title",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return("", errors.New("some error"))

	in := dto.AddFileDTO{
		Title: "Title",
		File:  make([]byte, 0),
		Note:  "test",
	}

	id, err := s.AddFile(ctx, 1, in)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}

func TestKeychainService_AddFile_Validate_Error(t *testing.T) {
	mockKeychainRepo := new(mocks.KeychainRepositoryMock)
	mockCryptManager := new(mocks.CryptManager)
	s := NewKeychainService(mockKeychainRepo, mockCryptManager)

	ctx := context.Background()

	in := dto.AddFileDTO{
		Note: "test",
	}

	id, err := s.AddFile(ctx, 1, in)

	assert.Error(t, err)
	assert.Empty(t, id)
	mockKeychainRepo.AssertExpectations(t)
	mockCryptManager.AssertExpectations(t)
}
