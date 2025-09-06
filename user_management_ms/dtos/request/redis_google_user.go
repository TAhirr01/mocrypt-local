package request

type RedisGoogleUser struct {
	Email    string `json:"email"`
	GoogleId string `json:"google_id"`
}

type GoogleRegistration struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	GoogleID string `json:"googleId"`
}

type RegistrationData struct {
	Email    string `json:"email"`
	Phone    string `json:"phone_number"`
	Verified bool   `json:"verified"`
	Complete string `json:"complete"`
}
