package query_repository

import (
	"errors"
	"user_management_ms/domain"

	"gorm.io/gorm"
)

type IUserQueryRepository interface {
	GetByID(db *gorm.DB, id uint) (*domain.User, error)
	GetUserByEmail(db *gorm.DB, email string) (*domain.User, error)
	GetUserByEmailOrPhone(db *gorm.DB, email, phone string) (*domain.User, error)
	GetCompletedUsersByEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error)
	GetUserByCredentialID(db *gorm.DB, credID []byte) (*domain.User, error)
	GetUserByGoogleId(db *gorm.DB, id string) (*domain.User, error)
	GetUserByPhoneNumber(db *gorm.DB, phone string) (*domain.User, error)
	GetUserWithEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error)
	IsUserWithPhoneExists(db *gorm.DB, phone string) (bool, error)
}

type UserQueryRepository struct{}

func NewUserQueryRepository() IUserQueryRepository {
	return &UserQueryRepository{}
}

func (u *UserQueryRepository) GetByID(db *gorm.DB, id uint) (*domain.User, error) {
	var user domain.User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserQueryRepository) GetUserByEmail(db *gorm.DB, email string) (*domain.User, error) {
	var user domain.User
	err := db.Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserQueryRepository) GetUserByEmailOrPhone(db *gorm.DB, email, phone string) (*domain.User, error) {
	var user domain.User
	if err := db.Where("email = ? OR phone = ?", email, phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserQueryRepository) GetCompletedUsersByEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error) {
	var user domain.User
	db.Where("email = ? AND phone = ? AND password is not null", email, phone).First(&user)
	if user.Password == "" {
		return nil, errors.New("user is not completed")
	}
	return &user, nil
}
func (u *UserQueryRepository) GetUserByCredentialID(db *gorm.DB, credentialID []byte) (*domain.User, error) {
	var user domain.User

	// Join with passkeys table to find user by credential ID
	err := db.Preload("Passkeys").
		Joins("JOIN user_passkeys ON users.id = user_passkeys.user_id").
		Where("user_passkeys.credential_id = ?", credentialID).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (u *UserQueryRepository) GetUserByGoogleId(db *gorm.DB, id string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("google_id=?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserQueryRepository) GetUserByPhoneNumber(db *gorm.DB, phone string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("phone=?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserQueryRepository) GetUserWithEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("email=? and phone=?", email, phone).First(&user).Error
	return &user, err
}
func (u *UserQueryRepository) IsUserWithPhoneExists(db *gorm.DB, phone string) (bool, error) {
	user := domain.User{}
	err := db.Where("phone=?", phone).First(&user).Error
	if err != nil {
		return false, nil
	}
	return true, nil
}
