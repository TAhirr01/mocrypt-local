package request

type PINRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	PIN   string `json:"pin"`
}
