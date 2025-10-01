package request

type StartGoogleRegistration struct {
	UserId uint   `json:"user_id"`
	Phone  string `json:"phone" validate:"required,phone"`
}

type PhoneRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
}

type StartRegistration struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,phone"`
	UserType string `json:"user_type" validate:"required"`
}
