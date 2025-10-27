package security

import (
	"golang.org/x/crypto/bcrypt"
)

// Hasher provides utilities for hashing and verifying passwords using bcrypt.
type Hasher struct{}

// NewHasher creates a new Hasher instance.
func NewHasher() Hasher {
	return Hasher{}
}

// HashPassword returns the bcrypt hash of the given plain-text password.
//
// It uses bcrypt.DefaultCost as the work factor. The returned string
// can be safely stored in the database and later verified using
// CheckPasswordHash.
//
// Example:
//
//	hash, _ := HashPassword("secret")
//	fmt.Println(hash) // $2a$10$...
func (h *Hasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPasswordHash compares a plain-text password with its bcrypt hash.
//
// It returns true if the password matches the hash, and false otherwise.
// This should be used during authentication to validate user credentials.
//
// Example:
//
//	match := CheckPasswordHash("secret", hash)
//	fmt.Println(match) // true
func (h *Hasher) CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
