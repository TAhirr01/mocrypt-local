package domain

import "time"

type WebAuthnCredential struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"index"`
	CredentialID []byte `gorm:"unique"`
	PublicKey    []byte
	SignCount    uint32
	CreatedAt    *time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"default:null" json:"updated_at"`
}
