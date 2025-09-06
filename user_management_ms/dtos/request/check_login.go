package request

type CheckLogin struct {
	Email string `json:"email" validate:"required,email"`
}
