package command_repository

import (
	"errors"
	"time"
	"user_management_ms/domain"
	"user_management_ms/util"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type IUserCommandRepository interface {
	Create(db *gorm.DB, entity *domain.User) (*domain.User, error)
	Update(db *gorm.DB, entity *domain.User) error
	Delete(db *gorm.DB, id uint) error
	DeleteUserOtpAndExpireDate(db *gorm.DB, user *domain.User) error
	SetUserEmailPhoneOtpAndExpireDates(db *gorm.DB, user *domain.User, emailOtp, phoneOtp string) error
	SavePasskey(db *gorm.DB, authBytes []byte, userID uint, cred *webauthn.Credential) error
	SaveUserOTPs(db *gorm.DB, email string, phone string, duration time.Duration) (string, string, error)
	UpdatePasskeyAfterLogin(db *gorm.DB, credID []byte, auth []byte, signCount uint32) error
	UpdateUserPasswordAndBirthDateById(db *gorm.DB, userId uint, hashPassword string, birthDate *time.Time) (*domain.User, error)
	UpdateUserPhoneById(db *gorm.DB, userId uint, phone, phoneOtp string, expire time.Time) (*domain.User, error)
}
type UserCommandRepository struct {
}

func NewUserCommandRepository() IUserCommandRepository {
	return &UserCommandRepository{}
}

func (u *UserCommandRepository) Create(db *gorm.DB, entity *domain.User) (*domain.User, error) {
	return entity, db.Create(entity).Error
}
func (u *UserCommandRepository) Update(db *gorm.DB, entity *domain.User) error {
	return db.Save(entity).Error
}
func (u *UserCommandRepository) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&domain.User{}, id).Error
}
func (u *UserCommandRepository) DeleteUserOtpAndExpireDate(db *gorm.DB, user *domain.User) error {
	user.PhoneOtp = ""
	user.EmailOtp = ""
	user.EmailOtpExpireDate = nil
	user.PhoneOtpExpireDate = nil
	return db.Save(user).Error
}
func (u *UserCommandRepository) SetUserEmailPhoneOtpAndExpireDates(db *gorm.DB, user *domain.User, emailOtp, phoneOtp string) error {
	user.EmailOtp = emailOtp
	user.PhoneOtp = phoneOtp
	t := time.Now().Add(5 * time.Minute)
	user.EmailOtpExpireDate = &t
	user.PhoneOtpExpireDate = &t
	return db.Save(user).Error
}
func (u *UserCommandRepository) SavePasskey(db *gorm.DB, authBytes []byte, userID uint, cred *webauthn.Credential) error {
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
func (u *UserCommandRepository) SaveUserOTPs(db *gorm.DB, email string, phone string, duration time.Duration) (string, string, error) {
	otpEmail := util.GenerateOTP()
	otpPhone := util.GenerateOTP()
	expire := time.Now().Add(duration)

	return otpEmail, otpPhone, db.Model(&domain.User{}).
		Where("email = ? AND phone = ?", email, phone).
		Updates(map[string]interface{}{
			"email_otp":             otpEmail,
			"email_otp_expire_date": expire,
			"phone_otp":             otpPhone,
			"phone_otp_expire_date": expire,
		}).Error
}
func (u *UserCommandRepository) UpdatePasskeyAfterLogin(db *gorm.DB, credID []byte, auth []byte, signCount uint32) error {
	return db.Model(&domain.Passkey{}).
		Where("credential_id = ?", credID).
		Updates(map[string]interface{}{
			"authenticator": auth,
			"sign_count":    signCount,
		}).Error
}
func (u *UserCommandRepository) UpdateUserPasswordAndBirthDateById(db *gorm.DB, userId uint, hashPassword string, birthDate *time.Time) (*domain.User, error) {
	var user *domain.User
	err := db.First(&user, userId).Error
	if err != nil {
		return nil, errors.New("user not found")
	}
	user.Password = hashPassword
	user.BirthDate = birthDate
	return user, db.Save(&user).Error
}

func (u *UserCommandRepository) UpdateUserPhoneById(db *gorm.DB, userId uint, phone, phoneOtp string, expire time.Time) (*domain.User, error) {
	user := domain.User{}
	err := db.First(&user, userId).Error
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
