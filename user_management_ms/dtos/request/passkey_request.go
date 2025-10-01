package request

type StartPasskeyRegistrationRequest struct {
	UserId uint `json:"user_id" validate:"required"`
}
