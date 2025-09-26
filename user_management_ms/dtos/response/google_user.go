package response

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type CallBackResponse struct {
	UserId   uint   `json:"user_id"`
	Status   string `json:"status"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	GoogleId string `json:"google_id"`
}
