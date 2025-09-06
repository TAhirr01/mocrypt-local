package response

type OTPResponse struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type OTPResponsePhone struct {
	Phone   string `json:"phone"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type OTPResponseEmail struct {
	Email   string `json:"email"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GoogleResponse struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	Status        string `json:"status"`
}
