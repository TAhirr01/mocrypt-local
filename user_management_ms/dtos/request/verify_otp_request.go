package request

type VerifyOTPRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	EmailOTP string `json:"email_otp" validate:"required"`
	PhoneOTP string `json:"phone_otp" validate:"required"`
}

type VerifyNumberOTPRequest struct {
	Email    string `json:"email" validate:"required,email"`
	PhoneOTP string `json:"phone_otp" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

type VerifyEmailOTPRequest struct {
	EmailOTP string `json:"email_otp" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}
