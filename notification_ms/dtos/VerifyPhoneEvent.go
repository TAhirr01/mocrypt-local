package dtos

type VerifyPhoneEvent struct {
	Phone string `yaml:"phone" json:"phone"`
	Otp   string `yaml:"otp" json:"otp"`
}
