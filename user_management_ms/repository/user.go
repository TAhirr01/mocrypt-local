package repository

import (
	"errors"
	"time"
	"user_management_ms/domain"
	"user_management_ms/util"

	"github.com/go-webauthn/webauthn/webauthn"
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
	GetUserWithPasskeys(db *gorm.DB, userId uint) (*domain.User, error)
	SavePasskey(db *gorm.DB, authBytes []byte, userID uint, cred *webauthn.Credential) error
	UpdateSignCount(db *gorm.DB, userID uint, signCount uint32) error
	UpdateSignCountByCredentialID(db *gorm.DB, credentialID []byte, signCount uint32) error
	GetCompletedUsersByEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error)
	UpdatePasskeyAfterLogin(db *gorm.DB, credID []byte, auth []byte, signCount uint32) error
	FindUserByCredentialID(db *gorm.DB, credID []byte) (*domain.User, error)
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
	user.EmailOtpExpireDate = nil
	user.PhoneOtpExpireDate = nil
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

func (u *UserRepository) GetUserWithPasskeys(db *gorm.DB, userId uint) (*domain.User, error) {
	var user domain.User
	err := db.Preload("Passkeys").First(&user, userId).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) SavePasskey(db *gorm.DB, authBytes []byte, userID uint, cred *webauthn.Credential) error {
	passkey := domain.Passkey{
		UserID:          userID,
		CredentialID:    cred.ID,
		PublicKey:       cred.PublicKey,
		SignCount:       cred.Authenticator.SignCount,
		AAGUID:          cred.Authenticator.AAGUID,
		AttestationType: cred.AttestationType,
		BackupState:     cred.Flags.BackupState,
		BackupEligible:  cred.Flags.BackupEligible,
		Authenticator:   authBytes,
	}

	if err := db.Create(&passkey).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserRepository) UpdateSignCount(db *gorm.DB, userID uint, signCount uint32) error {
	return db.Model(&domain.Passkey{}).
		Where("user_id = ?", userID).
		Update("sign_count", signCount).Error
}

func (u *UserRepository) UpdateSignCountByCredentialID(db *gorm.DB, credentialID []byte, signCount uint32) error {
	return db.Model(&domain.Passkey{}).
		Where("credential_id = ?", credentialID).
		Update("sign_count", signCount).Error
}

func (u *UserRepository) GetCompletedUsersByEmailAndPhone(db *gorm.DB, email string, phone string) (*domain.User, error) {
	var user domain.User
	db.Where("email = ? AND phone = ? AND password is not null", email, phone).First(&user)
	if user.Password == "" {
		return nil, errors.New("user is not completed")
	}
	return &user, nil
}

func (u *UserRepository) UpdatePasskeyAfterLogin(db *gorm.DB, credID []byte, auth []byte, signCount uint32) error {
	return db.Model(&domain.Passkey{}).
		Where("credential_id = ?", credID).
		Updates(map[string]interface{}{
			"authenticator": auth,
			"sign_count":    signCount,
		}).Error
}
func (u *UserRepository) FindUserByCredentialID(db *gorm.DB, credentialID []byte) (*domain.User, error) {
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
