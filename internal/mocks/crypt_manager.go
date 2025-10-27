package mocks

import "github.com/stretchr/testify/mock"

type CryptManager struct {
	mock.Mock
}

func (m *CryptManager) Encrypt(plaintext []byte) (nonce []byte, ciphertext []byte, err error) {
	args := m.Called(plaintext)
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *CryptManager) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	args := m.Called(nonce, ciphertext)
	return args.Get(0).([]byte), args.Error(1)
}
