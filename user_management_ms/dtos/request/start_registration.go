package request

type StartGoogleRegistration struct {
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"required,phone"`
}

type PhoneRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
}
