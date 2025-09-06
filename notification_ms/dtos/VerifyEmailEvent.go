package dtos

type VerifyEmailEvent struct {
	Email string `json:"email"`
	Otp   string `json:"email_otp"`
}
