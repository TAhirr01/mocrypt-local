package domain

import (
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
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
	Google2FASecret    string     // secret for TOTP
	Is2FAVerified      bool       `gorm:"default:false"`
	Passkeys           []Passkey  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user_passkeys"`
}

func (u User) WebAuthnID() []byte {
	return []byte(strconv.Itoa(int(u.Id)))
}
func (u User) WebAuthnName() string {
	return u.Email
}
func (u User) WebAuthnDisplayName() string {
	return u.Email
}
func (u User) WebAuthnCredentials() []webauthn.Credential {
	var creds []webauthn.Credential
	for _, p := range u.Passkeys {
		creds = append(creds, webauthn.Credential{
			ID:        p.CredentialID,
			PublicKey: p.PublicKey,
			Authenticator: webauthn.Authenticator{
				SignCount: p.SignCount,
			},
		})
	}
	return creds
}
