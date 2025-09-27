package request

type VerifyOTPRequest struct {
	UserId   uint   `json:"user_id"`
	EmailOTP string `json:"email_otp" validate:"required"`
	PhoneOTP string `json:"phone_otp" validate:"required"`
}

type VerifyNumberOTPRequest struct {
	UserId   uint   `json:"user_id"`
	PhoneOTP string `json:"phone_otp" validate:"required"`
}

type VerifyEmailOTPRequest struct {
	UserId   uint   `json:"user_id"`
	EmailOTP string `json:"email_otp" validate:"required"`
}
