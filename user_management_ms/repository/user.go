package repository

import (
	"time"
	"user_management_ms/domain"
	"user_management_ms/util"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetByID(db *gorm.DB, id uint) (*domain.User, error)
	Create(db *gorm.DB, entity *domain.User) (*domain.User, error)
	Update(db *gorm.DB, entity *domain.User) error
	Delete(db *gorm.DB, id uint) error
	IsUserWithEmailAndPhoneNumberExist(db *gorm.DB, email string, phone string) (bool, error)
	GetUserWithEmailAndPhoneNumber(db *gorm.DB, email string, phone string) (*domain.User, error)
	GetUserByEmail(db *gorm.DB, email string) (*domain.User, error)
	UpdateUserVerification(db *gorm.DB, email string, verify bool) error
	UpdateUserPasswordAndBirthDate(db *gorm.DB, email, hasPassword string, birthDate *time.Time) (*domain.User, error)
	GetUserByEmailOrPhone(db *gorm.DB, email, phone string) (*domain.User, error)
	SaveUserOTPs(db *gorm.DB, email string, phone string, duration time.Duration) error
	DeteUserOtpAndExpireDate(db *gorm.DB, user *domain.User) error
	SetUserEmailPhoneOtpAndExpireDates(db *gorm.DB, user *domain.User, emailOtp, phoneOtp string) error
}
type UserRepository struct {
}

func NewUserRepository() IUserRepository {
	return &UserRepository{}
}
func (u *UserRepository) GetByID(db *gorm.DB, id uint) (*domain.User, error) {
	var user domain.User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) Create(db *gorm.DB, entity *domain.User) (*domain.User, error) {
	return entity, db.Create(entity).Error
}

func (u *UserRepository) Update(db *gorm.DB, entity *domain.User) error {
	return db.Save(entity).Error
}

func (u *UserRepository) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&domain.User{}, id).Error
}

func (u *UserRepository) IsUserWithEmailAndPhoneNumberExist(db *gorm.DB, email string, phone string) (bool, error) {
	var user domain.User
	err := db.Where("email = ? or phone= ?", email, phone).First(&user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserRepository) GetUserWithEmailAndPhoneNumber(db *gorm.DB, email string, phone string) (*domain.User, error) {
	var user domain.User
	err := db.Where("email = ? and phone=?", email, phone).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (u *UserRepository) GetUserByEmail(db *gorm.DB, email string) (*domain.User, error) {
	var user domain.User
	err := db.Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (u *UserRepository) UpdateUserPasswordAndBirthDate(db *gorm.DB, email, hashPassword string, birthDate *time.Time) (*domain.User, error) {
	var user *domain.User
	err := db.Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}
	user.Password = hashPassword
	user.BirthDate = birthDate
	return user, db.Save(&user).Error
}

func (u *UserRepository) GetUserByEmailOrPhone(db *gorm.DB, email, phone string) (*domain.User, error) {
	var user domain.User
	if err := db.Where("email = ? OR phone = ?", email, phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) SaveUserOTPs(db *gorm.DB, email string, phone string, duration time.Duration) error {
	otpEmail := util.GenerateOTP()
	otpPhone := util.GenerateOTP()
	expire := time.Now().Add(duration)

	return db.Model(&domain.User{}).
		Where("email = ? AND phone = ?", email, phone).
		Updates(map[string]interface{}{
			"email_otp":                    otpEmail,
			"email_otp_expire_date":        expire,
			"phone_number_otp":             otpPhone,
			"phone_number_otp_expire_date": expire,
		}).Error
}

func (u *UserRepository) UpdateUserVerification(db *gorm.DB, email string, verified bool) error {
	return db.Model(&domain.User{}).
		Where("email = ?", email).
		Update("email_verified", verified).Error
}

func (u *UserRepository) DeteUserOtpAndExpireDate(db *gorm.DB, user *domain.User) error {
	user.PhoneOtp = ""
	user.EmailOtp = ""
	user.EmailOtpExpireDate = &time.Time{}
	user.PhoneOtpExpireDate = &time.Time{}
	return db.Save(user).Error
}

func (u *UserRepository) SetUserEmailPhoneOtpAndExpireDates(db *gorm.DB, user *domain.User, emailOtp, phoneOtp string) error {
	user.EmailOtp = emailOtp
	user.PhoneOtp = phoneOtp
	t := time.Now().Add(5 * time.Minute)
	user.EmailOtpExpireDate = &t
	user.PhoneOtpExpireDate = &t
	return db.Save(user).Error
}
