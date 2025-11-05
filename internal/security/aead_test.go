package security

import (
	"bytes"
	"github.com/thxhix/passKeeper/internal/config"
	"go.uber.org/zap"
	"testing"
)

var (
	RightSecret = "12345678901234567890123456789012"
)

func TestAEAD_EncryptDecrypt(t *testing.T) {
	logger := zap.NewNop()
	cryptCfg := config.CryptConfig{
		CryptSecretByte: RightSecret,
	}
	cfg := &config.Config{
		CryptConfig: cryptCfg,
	}

	aead, err := NewAEAD(logger, cfg)
	if err != nil {
		t.Fatalf("failed to create AEAD: %v", err)
	}

	plaintext := []byte("secret data")
	nonce, ciphertext, err := aead.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	decrypted, err := aead.Decrypt(nonce, ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestAEAD_WrongKeyLength(t *testing.T) {
	logger := zap.NewNop()
	cryptCfg := config.CryptConfig{
		CryptSecretByte: "12345",
	}
	cfg := &config.Config{
		CryptConfig: cryptCfg,
	}
	cfg.CryptSecretByte = "12345"

	_, err := NewAEAD(logger, cfg)
	if err != ErrAEADWrongLength {
		t.Fatalf("expected ErrAEADWrongLength, got %v", err)
	}
}

func TestAEAD_DecryptTamperedCiphertext(t *testing.T) {
	logger := zap.NewNop()
	cryptCfg := config.CryptConfig{
		CryptSecretByte: RightSecret,
	}
	cfg := &config.Config{
		CryptConfig: cryptCfg,
	}

	aead, err := NewAEAD(logger, cfg)
	if err != nil {
		t.Fatalf("failed to create AEAD: %v", err)
	}

	plaintext := []byte("secret data")
	nonce, ciphertext, err := aead.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// tamper with ciphertext
	ciphertext[0] ^= 0xFF

	_, err = aead.Decrypt(nonce, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail for tampered ciphertext")
	}
}

func TestAEAD_DecryptWrongNonce(t *testing.T) {
	logger := zap.NewNop()
	cryptCfg := config.CryptConfig{
		CryptSecretByte: RightSecret,
	}
	cfg := &config.Config{
		CryptConfig: cryptCfg,
	}

	aead, err := NewAEAD(logger, cfg)
	if err != nil {
		t.Fatalf("failed to create AEAD: %v", err)
	}

	plaintext := []byte("secret data")
	nonce, ciphertext, err := aead.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// modify nonce
	nonce[0] ^= 0xFF

	_, err = aead.Decrypt(nonce, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong nonce")
	}
}
