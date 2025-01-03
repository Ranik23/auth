package mock

import (
	"auth/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveUser(userName string, hashedPassword []byte, age int32, email string) error {
	args := m.Called(userName, hashedPassword, age, email)
	return args.Error(0)
}

func (m *MockStorage) DeleteUser(userName string) error {
	args := m.Called(userName)
	return args.Error(0)
}

func (m *MockStorage) GetUserByUserName(userName string) (*entity.User, error) {
	args := m.Called(userName)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockStorage) GetUserByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockStorage) ChangePassword(email string, newPassword []byte) error {
	args := m.Called(email, newPassword)
	return args.Error(1)
}