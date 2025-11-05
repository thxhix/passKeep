package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/thxhix/passKeeper/internal/config"
	"go.uber.org/zap"
	"io"
)

// AEAD provides authenticated encryption and decryption using AES-GCM.
// It holds a 32-byte key and a logger for error reporting.
type AEAD struct {
	key    []byte
	logger *zap.Logger
}

// NewAEAD creates a new AEAD instance using the secret key from the configuration.
// The key must be exactly 32 bytes long; otherwise, ErrAEADWrongLength is returned.
// The logger is used to log errors during encryption or decryption.
func NewAEAD(logger *zap.Logger, cfg *config.Config) (*AEAD, error) {
	secretBytes := []byte(cfg.CryptSecretByte)

	if len(secretBytes) != 32 {
		return nil, ErrAEADWrongLength
	}
	return &AEAD{
		logger: logger,
		key:    secretBytes,
	}, nil
}

// Encrypt encrypts the given plaintext using AES-GCM and returns:
// - a randomly generated nonce,
// - the ciphertext,
// - and an error if encryption fails.
// The nonce is required for decryption.
func (a *AEAD) Encrypt(plaintext []byte) (nonce []byte, ciphertext []byte, err error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.logger.Error("Failed to create new AES cipher", zap.Error(err))
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		a.logger.Error("Failed to create new GCM", zap.Error(err))
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		a.logger.Error("Failed to generate nonce", zap.Error(err))
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts the given ciphertext using AES-GCM with the provided nonce.
// It returns the original plaintext or an error if decryption fails.
// An error is also returned if the ciphertext was tampered with or if the key is invalid.
func (a *AEAD) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		a.logger.Error("Failed to create new AES cipher", zap.Error(err))
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		a.logger.Error("Failed to create new GCM", zap.Error(err))
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		a.logger.Error("Failed to decrypt ciphertext", zap.Error(err))
		return nil, err
	}
	return plaintext, nil
}
