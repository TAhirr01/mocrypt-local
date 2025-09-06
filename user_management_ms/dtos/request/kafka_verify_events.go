package request

type VerifyEmailEvent struct {
	Email    string `json:"email"`
	EmailOTP string `json:"email_otp"`
}

type VerifyPhoneEvent struct {
	Phone    string `json:"phone"`
	PhoneOTP string `json:"phone_otp"`
}
