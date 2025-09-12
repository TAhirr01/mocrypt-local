package response

type OTPResponse struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	Status        string `json:"status"`
}

type SendOTPResponse struct {
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type RegisterResponse struct {
	UserType      string `json:"user_type"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	Completed     bool   `json:"completed"`
	Status        string `json:"status"`
}

type OTPResponsePhone struct {
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type OTPResponseEmail struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type GoogleResponse struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	Status        string `json:"status"`
}

type LoginResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}
