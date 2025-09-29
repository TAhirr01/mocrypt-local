package response

type Status string

const (
	VERIFIED             Status = "verified"
	UNVERIFIED           Status = "unverified"
	PHONE_MISMATCH       Status = "phone_mismatch"
	VERIFICATION_PENDING Status = "verification_pending"
	PHONE_EXISTS         Status = "user with this phone already exists"
	SET_PIN              Status = "set_pin"
	VERIFY_PIN           Status = "verify_pin"
)

type OTPResponse struct {
	UserId        uint   `json:"user_id"`
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
	UserId        uint   `json:"user_id"`
	UserType      string `json:"user_type"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	Completed     bool   `json:"completed"`
	Status        string `json:"status"`
	HasPin        bool   `json:"has_pin"`
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
	Completed     bool   `json:"completed"`
	HasPin        bool   `json:"has_pin"`
	UserId        uint   `json:"user_id"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	Status        Status `json:"status"`
}

type LoginResponse struct {
	UserId uint   `json:"user_id"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
}

type AfterRegisterPassword struct {
	UserId uint   `json:"user_id"`
	Status Status `json:"status"`
}

type AfterLoginVerification struct {
	UserId uint   `json:"user_id"`
	Status Status `json:"status"`
}
