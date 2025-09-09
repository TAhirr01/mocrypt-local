package request

type StartPasskeyRegistrationRequest struct {
	UserId uint `json:"user_id" validate:"required"`
}

type FinishPasskeyRegistrationRequest struct {
	UserId uint `json:"user_id" validate:"required"`
}
