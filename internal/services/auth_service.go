package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"github.com/thxhix/passKeeper/internal/domain/user"
	"strconv"
	"time"
)

// PasswordHasher defines password hashing operations required by AuthService.
//
// Implementations must provide a secure, one-way hash for passwords and a method
// to verify a plain password against the stored hash.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

// TokenManager describes token-related helper operations used by AuthService.
//
// Implementations are responsible for creating, parsing and deriving values
// required for refresh and access tokens (JTI, TTL, hashing, etc.).
type TokenManager interface {
	GenerateRefreshToken(userID int64) (token string, jti uuid.UUID, ttl time.Duration, err error)
	ParseRefreshToken(tokenStr string) (userID, jti string, err error)
	GenerateAccessToken(userID int64) (string, error)
	Sha256Hex(s string) string
}

type IAuthService interface {
	Register(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error)
	Login(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error)
	Refresh(ctx context.Context, incomingRefreshToken string) (accessToken string, refreshToken string, err error)
}

// AuthService provides registration, login and refresh workflows.
//
// AuthService relies on user and token repositories, a password hasher and a token
// manager to perform operations. Methods are safe to call from handlers and are
// responsible for validating inputs, coordinating calls to dependencies and
// returning domain-level errors.
type AuthService struct {
	userRepo  user.UserRepository
	tokenRepo token.TokenRepository

	hasher       PasswordHasher
	tokenManager TokenManager
}

// NewAuthService constructs a new AuthService with given dependencies.
func NewAuthService(userRepo user.UserRepository, tokenRepo token.TokenRepository, hasher PasswordHasher, tokenManager TokenManager) AuthService {
	return AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,

		hasher:       hasher,
		tokenManager: tokenManager,
	}
}

// Register creates a new user account and issues auth tokens.
//
// It validates input (login and password), hashes the password, creates a user
// record in the user repository and then generates both access and refresh tokens.
// The refresh token is stored in the token repository (hashed) together with its
// JTI and TTL.
//
// Returns the created user id, an access token, a refresh token and an error.
// On validation or repository error, the returned error describes the failure.
func (s *AuthService) Register(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error) {
	if err := user.ValidateLogin(login); err != nil {
		return 0, "", "", err
	}

	if err := user.ValidatePassword(password); err != nil {
		return 0, "", "", err
	}

	// Generate password hash for security store in storage
	passwordHash, err := s.hasher.HashPassword(password)
	if err != nil {
		return 0, "", "", err
	}

	userId, err = s.userRepo.Create(ctx, login, passwordHash)
	if err != nil {
		return 0, "", "", err
	}

	accessToken, err = s.tokenManager.GenerateAccessToken(userId)
	if err != nil {
		return 0, "", "", err
	}

	refreshToken, refreshJTI, refreshTTL, err := s.tokenManager.GenerateRefreshToken(userId)
	if err != nil {
		return 0, "", "", err
	}

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(refreshTTL)
	refreshHash := s.tokenManager.Sha256Hex(refreshToken)

	if err := s.tokenRepo.Create(ctx, userId, refreshJTI, refreshHash, issuedAt, expiresAt); err != nil {
		return 0, "", "", err
	}

	return userId, accessToken, refreshToken, nil
}

// Login authenticates a user by login and password and returns tokens.
//
// It validates the login format, retrieves the user by login, verifies the
// provided password against stored hash and — on success — generates and stores
// a new refresh token while returning a fresh access token as well.
//
// Returns the user id, an access token, a refresh token and an error. If the
// credentials are invalid, user.ErrInvalidCredentials is returned.
func (s *AuthService) Login(ctx context.Context, login string, password string) (userId int64, accessToken string, refreshToken string, err error) {
	if err := user.ValidateLogin(login); err != nil {
		return 0, "", "", err
	}

	au, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return 0, "", "", user.ErrInvalidCredentials
		}
		return 0, "", "", err
	}

	if !s.hasher.CheckPasswordHash(password, au.PasswordHash) {
		return 0, "", "", user.ErrInvalidCredentials
	}

	userId = au.ID

	accessToken, err = s.tokenManager.GenerateAccessToken(userId)
	if err != nil {
		return 0, "", "", err
	}

	refreshToken, refreshJTI, refreshTTL, err := s.tokenManager.GenerateRefreshToken(userId)
	if err != nil {
		return 0, "", "", err
	}

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(refreshTTL)
	refreshHash := s.tokenManager.Sha256Hex(refreshToken)

	if err := s.tokenRepo.Create(ctx, userId, refreshJTI, refreshHash, issuedAt, expiresAt); err != nil {
		return 0, "", "", err
	}

	return userId, accessToken, refreshToken, nil
}

// Refresh validates an incoming refresh token and rotates it.
//
// The method parses the incoming refresh token to extract user id and JTI,
// verifies the token against the stored token record (by hashed token value and JTI),
// and if valid generates a new access token and a new refresh token. The token
// repository is updated (rotated) atomically with the new JTI and hash.
//
// Returns the new access token, the new refresh token and an error. If the
// incoming token is invalid or does not match stored state, user.ErrInvalidRefreshCredentials
// is returned.
func (s *AuthService) Refresh(ctx context.Context, incomingRefreshToken string) (accessToken string, refreshToken string, err error) {
	userIdStr, jtiStr, err := s.tokenManager.ParseRefreshToken(incomingRefreshToken)
	if err != nil {
		return "", "", err
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return "", "", err
	}
	oldJTI, err := uuid.Parse(jtiStr)
	if err != nil {
		return "", "", err
	}

	incomingTokenHash := s.tokenManager.Sha256Hex(incomingRefreshToken)

	tokenRecord, err := s.tokenRepo.GetByJTI(ctx, oldJTI)
	if err != nil {
		if errors.Is(err, token.ErrTokenDoesntExistsByJTI) {
			return "", "", user.ErrInvalidRefreshCredentials
		}
		return "", "", err
	}

	now := time.Now().UTC()

	if err := token.ValidateToken(now, userId, incomingTokenHash, tokenRecord); err != nil {
		return "", "", user.ErrInvalidRefreshCredentials
	}

	accessToken, err = s.tokenManager.GenerateAccessToken(userId)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, newJTI, newTTL, err := s.tokenManager.GenerateRefreshToken(userId)
	if err != nil {
		return "", "", err
	}

	newHash := s.tokenManager.Sha256Hex(newRefreshToken)

	if err := s.tokenRepo.Rotate(ctx, userId, oldJTI, newJTI, newHash, now, now.Add(newTTL)); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
