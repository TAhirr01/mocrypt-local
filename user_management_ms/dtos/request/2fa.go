package request

type VerifyTwoFa struct {
	Code string `json:"code" validate:"required"`
}
