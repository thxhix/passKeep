package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"github.com/thxhix/passKeeper/internal/domain/user"
	"github.com/thxhix/passKeeper/internal/mocks"
	"testing"
	"time"
)

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	tokenRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	userRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(1, nil)
	passHasher.On("HashPassword", mock.Anything, mock.Anything).Return("123456", nil)
	tokenManager.On("GenerateAccessToken", mock.Anything).Return("accessToken", nil)
	tokenManager.On("GenerateRefreshToken", mock.Anything).Return("refreshToken", uuid.New(), time.Duration(1), nil)
	tokenManager.On("Sha256Hex", mock.Anything).Return("password_hash")

	_, access, refresh, err := s.Register(ctx, "login", "password")

	assert.NoError(t, err)

	assert.Equal(t, "accessToken", access)
	assert.Equal(t, "refreshToken", refresh)

	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	passHasher.AssertExpectations(t)
	tokenManager.AssertExpectations(t)
}

func TestAuthService_Register_Error(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	passHasher.On("HashPassword", mock.Anything, mock.Anything).Return("123456", errors.New("some error"))

	_, _, _, err := s.Register(ctx, "login", "password")

	assert.Error(t, err)

	passHasher.AssertExpectations(t)
}

func TestAuthService_Register_Validate_Error(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	_, _, _, err := s.Register(ctx, "l", "pass")

	assert.Error(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	tokenRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	userRepo.On("GetByLogin", mock.Anything, mock.Anything).Return(&user.UserRecord{ID: 1, Login: "test", PasswordHash: "password"}, nil)
	passHasher.On("CheckPasswordHash", mock.Anything, mock.Anything).Return(true)
	tokenManager.On("GenerateAccessToken", mock.Anything).Return("accessToken", nil)
	tokenManager.On("GenerateRefreshToken", mock.Anything).Return("refreshToken", uuid.New(), time.Duration(1), nil)
	tokenManager.On("Sha256Hex", mock.Anything).Return("password_hash")

	_, access, refresh, err := s.Login(ctx, "login", "password")

	assert.NoError(t, err)

	assert.Equal(t, "accessToken", access)
	assert.Equal(t, "refreshToken", refresh)

	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	passHasher.AssertExpectations(t)
	tokenManager.AssertExpectations(t)
}

func TestAuthService_Login_Wrong_User(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	userRepo.On("GetByLogin", mock.Anything, mock.Anything).Return(&user.UserRecord{}, errors.New("cant find user"))

	_, _, _, err := s.Login(ctx, "login", "password")

	assert.Error(t, err)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_Validate_Error(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	_, _, _, err := s.Login(ctx, "l", "pass")

	assert.Error(t, err)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	tokenRepo.On("Rotate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tokenRepo.On("GetByJTI", mock.Anything, mock.Anything).Return(&token.RefreshTokenRecord{
		JTI:       uuid.New(),
		UserID:    1,
		TokenHash: "hash",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 10),
	}, nil)

	tokenManager.On("ParseRefreshToken", mock.Anything).Return("1", uuid.NewString(), nil)
	tokenManager.On("GenerateAccessToken", mock.Anything).Return("accessToken", nil)
	tokenManager.On("GenerateRefreshToken", mock.Anything).Return("refreshToken", uuid.New(), time.Duration(1), nil)
	tokenManager.On("Sha256Hex", mock.Anything).Return("hash")

	access, refresh, err := s.Refresh(ctx, "refreshToken")

	assert.NoError(t, err)

	assert.Equal(t, "accessToken", access)
	assert.Equal(t, "refreshToken", refresh)

	tokenRepo.AssertExpectations(t)
	tokenManager.AssertExpectations(t)
}

func TestAuthService_Refresh_Wrong_Refresh(t *testing.T) {
	userRepo := new(mocks.UserRepositoryMock)
	tokenRepo := new(mocks.TokenRepositoryMock)
	passHasher := new(mocks.PasswordHasherMock)
	tokenManager := new(mocks.TokenManagerMock)

	ctx := context.Background()

	s := NewAuthService(userRepo, tokenRepo, passHasher, tokenManager)

	tokenRepo.On("GetByJTI", mock.Anything, mock.Anything).Return(&token.RefreshTokenRecord{
		JTI:       uuid.New(),
		UserID:    1,
		TokenHash: "wrong",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 10),
	}, nil)

	tokenManager.On("ParseRefreshToken", mock.Anything).Return("1", uuid.NewString(), nil)
	tokenManager.On("Sha256Hex", mock.Anything).Return("hash")

	_, _, err := s.Refresh(ctx, "refreshToken")

	assert.Error(t, err)

	tokenRepo.AssertExpectations(t)
	tokenManager.AssertExpectations(t)
}
