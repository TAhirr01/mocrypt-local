package request

type PinReq struct {
	Pin string `json:"pin" validate:"required,pin"`
}
