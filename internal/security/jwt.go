package security

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/config"
	"strconv"
	"time"
)

// JWTManager handles generation and validation of access and refresh JWT tokens.
//
// It relies on secret keys defined in the config and supports HS256 signing.
type JWTManager struct {
	cfg *config.Config
}

// NewJWTManager creates a new JWTManager with the given configuration.
func NewJWTManager(cfg *config.Config) JWTManager {
	return JWTManager{
		cfg: cfg,
	}
}

// Claims defines JWT claims used for both access and refresh tokens.
// It embeds the standard RegisteredClaims (sub, iss, aud, exp, iat, nbf, jti).
type Claims struct {
	jwt.RegisteredClaims
}

// IsValidSecretKey validates the provided secret length.
//
// It returns true if the key is more than 32 bytes,
// which makes HS256 signing secure.
func IsValidSecretKey(key string) bool {
	return len(key) >= 32
}

// GenerateAccessToken issues a new access JWT for the given user ID.
//
// The access token has a short lifespan (configured via
// cfg.JWTAccessExpTimeMinute) and should be included in the
// Authorization header for every API request.
//
// Returns the signed token string or an error if the secret is too short.
func (j *JWTManager) GenerateAccessToken(userID int64) (string, error) {
	if !IsValidSecretKey(j.cfg.JWTAccessSecretKey) {
		return "", ErrSecretTooShort
	}

	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(userID)),
			Issuer:    j.cfg.JWTIssuer,
			Audience:  []string{j.cfg.JWTAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.cfg.JWTAccessExpTimeMinute) * time.Minute)),
			ID:        uuid.NewString(),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(j.cfg.JWTAccessSecretKey))
}

// ParseAccessToken validates and parses an access JWT string.
//
// It checks the signing algorithm, issuer, audience and expiration time.
// Returns the userID (from the `sub` claim) or an error if the token is invalid.
func (j *JWTManager) ParseAccessToken(tokenStr string) (userID string, err error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.JWTAccessSecretKey), nil
	}
	var claims Claims
	tkn, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
		jwt.WithIssuer(j.cfg.JWTIssuer),
		jwt.WithAudience(j.cfg.JWTAudience),
		jwt.WithLeeway(30*time.Second),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !tkn.Valid {
		return "", ErrAccessExpiredToken
	}
	if claims.Subject == "" {
		return "", ErrAccessInvalidSubject
	}
	return claims.Subject, nil
}

// GenerateRefreshToken issues a new refresh JWT for the given user ID.
//
// Refresh tokens have a long lifespan (configured via cfg.JWTRefreshExpTimeDays).
// They are returned along with a unique JTI (token ID) that can be stored
// in a database for session management and revocation.
//
// Returns the signed refresh token, its JTI, and an error if any.
func (j *JWTManager) GenerateRefreshToken(userID int64) (token string, jti uuid.UUID, ttl time.Duration, err error) {
	if !IsValidSecretKey(j.cfg.JWTRefreshSecretKey) {
		return "", uuid.Nil, 0, ErrSecretTooShort
	}

	now := time.Now().UTC()
	jti = uuid.New()
	ttl = time.Duration(j.cfg.JWTRefreshExpTimeDays) * 24 * time.Hour

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(userID)),
			Issuer:    j.cfg.JWTIssuer,
			Audience:  []string{j.cfg.JWTAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti.String(),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := t.SignedString([]byte(j.cfg.JWTRefreshSecretKey))
	if err != nil {
		return "", uuid.Nil, 0, err
	}
	return tok, jti, ttl, nil
}

// ParseRefreshToken validates and parses a refresh JWT string.
//
// It checks the signing algorithm, issuer, audience and expiration time.
// Returns the userID (from the `sub` claim), the JTI (from the `jti` claim),
// or an error if the token is invalid.
func (j *JWTManager) ParseRefreshToken(tokenStr string) (userID, jti string, err error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.JWTRefreshSecretKey), nil
	}
	var claims Claims
	tkn, err := jwt.ParseWithClaims(tokenStr, &claims, keyFunc,
		jwt.WithIssuer(j.cfg.JWTIssuer),
		jwt.WithAudience(j.cfg.JWTAudience),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil || !tkn.Valid {
		return "", "", ErrRefreshExpiredToken
	}
	if claims.Subject == "" || claims.ID == "" {
		return "", "", ErrRefreshInvalidClaims
	}
	return claims.Subject, claims.ID, nil
}

// Sha256Hex calculates SHA256 hash of the input string and returns it as a hex string.
func (j *JWTManager) Sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
