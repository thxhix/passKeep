package mocks

import "github.com/stretchr/testify/mock"

type PasswordHasherMock struct {
	mock.Mock
}

func (m *PasswordHasherMock) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.Get(0).(string), args.Error(1)
}

func (m *PasswordHasherMock) CheckPasswordHash(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}
