package mock

import (
	"auth/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveUser(userName string, hashedPassword string, age int32, email string) error {
	args := m.Called(userName, hashedPassword, age, email)
	return args.Error(0)
}

func (m *MockStorage) DeleteUser(userName string) error {
	args := m.Called(userName)
	return args.Error(0)
}

func (m *MockStorage) GetUser(userName string) (*entity.User, error) {
	args := m.Called(userName)
	return args.Get(0).(*entity.User), args.Error(1)
}