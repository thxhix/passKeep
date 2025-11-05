package security

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHasher_HashAndCheckPassword(t *testing.T) {
	h := NewHasher()

	password := "mySecret123!"

	// Генерация хэша
	hash, err := h.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Проверка правильного пароля
	match := h.CheckPasswordHash(password, hash)
	require.True(t, match, "expected password to match hash")

	// Проверка неправильного пароля
	wrongMatch := h.CheckPasswordHash("wrongPassword", hash)
	require.False(t, wrongMatch, "expected wrong password not to match hash")
}

func TestHasher_HashPassword_UniqueHashes(t *testing.T) {
	h := NewHasher()

	password := "samePassword"

	hash1, err := h.HashPassword(password)
	require.NoError(t, err)
	hash2, err := h.HashPassword(password)
	require.NoError(t, err)

	// Два хэша одного и того же пароля должны быть разными из-за соли
	require.NotEqual(t, hash1, hash2)
}

func TestHasher_CheckPasswordHash_InvalidHash(t *testing.T) {
	h := NewHasher()

	// Некорректный хэш должен вернуть false
	match := h.CheckPasswordHash("password", "invalidHash")
	require.False(t, match)
}
