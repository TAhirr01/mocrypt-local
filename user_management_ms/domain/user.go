package domain

import (
	"time"
)

type User struct {
	Id                 uint       `gorm:"primaryKey" json:"id"`
	CreatedAt          *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"default:null" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"default:null" json:"deleted_at"`
	Email              string     `gorm:"size:100;not null" json:"email"`
	Phone              string     `gorm:"size:100;not null" json:"phone"`
	BirthDate          *time.Time `gorm:"default:NULL" json:"birth_date"`
	Password           string     `gorm:"size:100;not null" json:"password"`
	GoogleID           string     `gorm:"size:100;" json:"google_id"`
	EmailOtp           string     `gorm:"size:100;not null" json:"email_otp"`
	PhoneOtp           string     `gorm:"size:100;not null" json:"phone_otp"`
	EmailVerified      bool       `json:"email_verified"`
	PhoneVerified      bool       `json:"phone_verified"`
	EmailOtpExpireDate *time.Time `gorm:"default:NULL" json:"email_otp_expire_date"`
	PhoneOtpExpireDate *time.Time `gorm:"default:NULL" json:"phone_otp_expire_date"`
	//Credentials     []WebAuthnCredential `gorm:"foreignKey:UserID"`
	//LastFullAuth    *time.Time
	//LastPasskeyAuth *time.Time
	//WebAuthnUserHandle []byte
}

//func (u *User) WebAuthnID() []byte {
//	return u.WebAuthnUserHandle
//}

//func (u *User) WebAuthnName() string {
//	return u.Email
//}
//
//func (u *User) WebAuthnDisplayName() string {
//	return u.Email
//}
//
//func (u *User) WebAuthnIcon() string {
//	return ""
//}
//
//func (u *User) WebAuthnCredentials() []webauthn.Credential {
//	creds := make([]webauthn.Credential, 0, len(u.Credentials))
//	for _, c := range u.Credentials {
//		creds = append(creds, webauthn.Credential{
//			ID:        c.CredentialID,
//			PublicKey: c.PublicKey,
//			Authenticator: webauthn.Authenticator{
//				SignCount: c.SignCount,
//			},
//		})
//	}
//	return creds
//}
