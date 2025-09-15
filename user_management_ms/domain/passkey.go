package domain

import "time"

type Passkey struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"user_id"` // foreign key
	CredentialID    []byte     `gorm:"not null;unique" json:"credential_id"`
	PublicKey       []byte     `gorm:"not null" json:"public_key"`
	SignCount       uint32     `gorm:"not null" json:"sign_count"`
	CreatedAt       *time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"default:null" json:"updated_at"`
	AAGUID          []byte     `gorm:"not null" json:"aa_guid"`
	AttestationType string
	Authenticator   []byte `gorm:"type:json"`
	BackupEligible  bool   `gorm:"not null;default:false" json:"backup_eligible"`
	BackupState     bool   `gorm:"not null;default:false" json:"backup_state"`
}

func (Passkey) TableName() string {
	return "user_passkeys"
}
