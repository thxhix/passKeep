package config

// CryptConfig holds the cryptographic configuration for the application.
//
// CryptSecretByte is the secret key used for symmetric encryption (e.g., AES-GCM).
// It should be exactly 32 bytes for AES-256.
type CryptConfig struct {
	CryptSecretByte string `envDefault:"12345678901234567890123456789012"`
}
