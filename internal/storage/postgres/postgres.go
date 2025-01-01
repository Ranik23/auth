package postgres

import (
	"auth/internal/config"
	"auth/internal/entity"
	"errors"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



type Storage interface {
	SaveUser(userName string, email string, age int32, hashedPassword []byte) error
	DeleteUser(userName string) error
	GetUser(userName string) (*entity.User, error)
}

type StorageImpl struct {
	db *gorm.DB
}

func NewStoragePostgres(cfg *config.Config) (*StorageImpl, error) {
	db, err := ConnectToDb(cfg)
	if err != nil {
		return nil, err
	}
	
	return &StorageImpl{db: db}, nil
}

func (s *StorageImpl) SaveUser(userName string, email string, age int32, hashedPassword []byte) error {
	user := &entity.User{
		UserName: userName,
		HashedPassword: hashedPassword,
		Age: age,
		Email: email,
	}

	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	log.Println("user saved successfully")

	return nil
}

func (s *StorageImpl) DeleteUser(userName string) error {
	user := &entity.User{
		UserName: userName,
	}
	if err := s.db.Delete(user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Println("user deleted successfully")
	return nil
}

func (s *StorageImpl) GetUser(userName string) (*entity.User, error) {
	var user entity.User
	if err := s.db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("user record not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.Printf("error fetching user: %v", err)
		return nil, err
	}

	return &user, nil
}


func ConnectToDb(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabasePort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
