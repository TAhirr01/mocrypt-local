package request

type OTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"required"`
}

type OTPRequestPhone struct {
	UserId uint   `json:"userId" validate:"required,gte=1"`
	Phone  string `json:"phone" validate:"required"`
}

type OTPRequestEmail struct {
	UserId uint   `json:"userId" validate:"required,gte=1"`
	Email  string `json:"email" validate:"required,email"`
}

type EmailAndPhoneOTP struct {
	EmailOTP string `json:"email_otp" validate:"required"`
	PhoneOTP string `json:"phone_otp" validate:"required"`
}

type PhoneOTP struct {
	PhoneOTP string `json:"phone_otp" validate:"required"`
}

type EmailOTP struct {
	EmailOTP string `json:"email_otp" validate:"required"`
}
