package domain

import (
	"encoding/json"
	"log"
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
	Loginable          bool       `gorm:"default:false" json:"loginable"`
	Password           string     `gorm:"size:100;not null" json:"password"`
	GoogleID           string     `gorm:"size:100;" json:"google_id"`
	EmailOtp           string     `gorm:"size:100;not null" json:"email_otp"`
	PhoneOtp           string     `gorm:"size:100;not null" json:"phone_otp"`
	EmailVerified      bool       `json:"email_verified"`
	PhoneVerified      bool       `json:"phone_verified"`
	EmailOtpExpireDate *time.Time `gorm:"default:NULL" json:"email_otp_expire_date"`
	PhoneOtpExpireDate *time.Time `gorm:"default:NULL" json:"phone_otp_expire_date"`
	PINHash            string     `gorm:"size:100;default:null" json:"pin_hash"`
	Is2FAVerified      bool       `gorm:"default:false"`
	UserType           string     `gorm:"size:100;default:null" json:"user_type"`
	Google2FASecret    string
	Passkeys           []Passkey `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user_passkeys"`
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
		var auth webauthn.Authenticator
		if len(p.Authenticator) > 0 {
			if err := json.Unmarshal(p.Authenticator, &auth); err != nil {
				log.Printf("failed to unmarshal authenticator: %v", err)
				continue
			}
		}

		creds = append(creds, webauthn.Credential{
			ID:              p.CredentialID,
			PublicKey:       p.PublicKey,
			Authenticator:   auth,
			AttestationType: p.AttestationType,
			Flags: webauthn.CredentialFlags{
				BackupEligible: p.BackupEligible,
				BackupState:    p.BackupState,
			},
		})
	}
	return creds
}
