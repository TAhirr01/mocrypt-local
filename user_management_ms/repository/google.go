package repository

import (
	"time"
	"user_management_ms/domain"

	"gorm.io/gorm"
)

type IGoogleRepository interface {
	FindUserByGoogleId(db *gorm.DB, id string) (*domain.User, error)
	FindUserByPhoneNumber(db *gorm.DB, phone string) (*domain.User, error)
	Create(db *gorm.DB, entity *domain.User) (*domain.User, error)
	FindUserByEmail(db *gorm.DB, email string) (*domain.User, error)
	UpdateGoogleUserPhone(db *gorm.DB, email, phone, phoneOtp string, expire time.Time) (*domain.User, error)
	GetUserWithEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error)
	UpdateUserBirthdayAndPassword(db *gorm.DB, email, password string, birthDay *time.Time) (*domain.User, error)
	UpdateUserVerifyStatus(db *gorm.DB, email string, verify bool) (*domain.User, error)
	Update(db *gorm.DB, entity *domain.User) (*domain.User, error)
	IsUserWithPhoneExists(db *gorm.DB, phone string) (bool, error)
}

type GoogleRepository struct{}

func NewGoogleRepository() *GoogleRepository {
	return &GoogleRepository{}
}

func (s *GoogleRepository) FindUserByGoogleId(db *gorm.DB, id string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("google_id=?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GoogleRepository) FindUserByPhoneNumber(db *gorm.DB, phone string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("phone=?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GoogleRepository) Create(db *gorm.DB, entity *domain.User) (*domain.User, error) {
	entity.EmailVerified = true
	return entity, db.Create(entity).Error
}

func (s *GoogleRepository) FindUserByEmail(db *gorm.DB, email string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GoogleRepository) UpdateGoogleUserPhone(db *gorm.DB, email, phone, phoneOtp string, expire time.Time) (*domain.User, error) {
	user := domain.User{}
	err := db.Model(&user).Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	if phone != "" {
		user.Phone = phone
	}
	user.PhoneOtp = phoneOtp
	user.PhoneOtpExpireDate = &expire
	return &user, db.Save(&user).Error
}

func (s *GoogleRepository) GetUserWithEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error) {
	user := domain.User{}
	err := db.Where("email=? and phone=?", email, phone).First(&user).Error
	return &user, err
}

func (s *GoogleRepository) UpdateUserBirthdayAndPassword(db *gorm.DB, email, password string, birthDay *time.Time) (*domain.User, error) {
	user := domain.User{}
	err := db.Model(&user).Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	user.Password = password
	user.BirthDate = birthDay
	return &user, db.Save(&user).Error
}

func (s *GoogleRepository) UpdateUserVerifyStatus(db *gorm.DB, phone string, verify bool) (*domain.User, error) {
	user := domain.User{}
	err := db.Model(&user).Where("phone=?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	user.PhoneVerified = verify
	return &user, db.Save(&user).Error
}

func (s *GoogleRepository) Update(db *gorm.DB, entity *domain.User) (*domain.User, error) {
	return entity, db.Save(entity).Error
}

func (s *GoogleRepository) IsUserWithPhoneExists(db *gorm.DB, phone string) (bool, error) {
	user := domain.User{}
	err := db.Where("phone=?", phone).First(&user).Error
	if err != nil {
		return false, nil
	}
	return true, nil
}
