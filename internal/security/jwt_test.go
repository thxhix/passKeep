package security

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/thxhix/passKeeper/internal/config"
	"testing"
)

func TestJWTManager_AccessToken(t *testing.T) {
	jwtCfg := config.JWTConfig{
		JWTAccessSecretKey:     RightSecret,
		JWTRefreshSecretKey:    RightSecret,
		JWTIssuer:              "test-issuer",
		JWTAudience:            "test-audience",
		JWTAccessExpTimeMinute: 5,
	}
	cfg := &config.Config{
		JWTConfig: jwtCfg,
	}

	jm := NewJWTManager(cfg)
	userID := int64(42)

	// Generate access token
	token, err := jm.GenerateAccessToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Parse access token
	sub, err := jm.ParseAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, "42", sub)
}

func TestJWTManager_RefreshToken(t *testing.T) {
	jwtCfg := config.JWTConfig{
		JWTAccessSecretKey:     RightSecret,
		JWTRefreshSecretKey:    RightSecret,
		JWTIssuer:              "test-issuer",
		JWTAudience:            "test-audience",
		JWTAccessExpTimeMinute: 5,
		JWTRefreshExpTimeDays:  30,
	}
	cfg := &config.Config{
		JWTConfig: jwtCfg,
	}

	jm := NewJWTManager(cfg)
	userID := int64(100)

	// Generate refresh token
	token, jti, ttl, err := jm.GenerateRefreshToken(userID)

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEqual(t, uuid.Nil, jti)
	require.True(t, ttl > 0)

	// Parse refresh token
	sub, parsedJTI, err := jm.ParseRefreshToken(token)
	require.NoError(t, err)
	require.Equal(t, "100", sub)
	require.Equal(t, jti.String(), parsedJTI)
}

func TestJWTManager_CheckSecretKey(t *testing.T) {
	shortKey := "shortkey"
	longKey := RightSecret

	require.False(t, IsValidSecretKey(shortKey))
	require.True(t, IsValidSecretKey(longKey))
}

func TestJWTManager_InvalidAccessToken(t *testing.T) {
	jwtCfg := config.JWTConfig{
		JWTAccessSecretKey:     RightSecret,
		JWTRefreshSecretKey:    RightSecret,
		JWTIssuer:              "test-issuer",
		JWTAudience:            "test-audience",
		JWTAccessExpTimeMinute: 5,
	}
	cfg := &config.Config{
		JWTConfig: jwtCfg,
	}
	jm := NewJWTManager(cfg)

	_, err := jm.ParseAccessToken("invalid.token.here")
	require.Error(t, err)
}

func TestJWTManager_InvalidRefreshToken(t *testing.T) {
	jwtCfg := config.JWTConfig{
		JWTAccessSecretKey:     RightSecret,
		JWTRefreshSecretKey:    RightSecret,
		JWTIssuer:              "test-issuer",
		JWTAudience:            "test-audience",
		JWTAccessExpTimeMinute: 5,
	}
	cfg := &config.Config{
		JWTConfig: jwtCfg,
	}
	jm := NewJWTManager(cfg)

	_, _, err := jm.ParseRefreshToken("invalid.token.here")
	require.Error(t, err)
}
