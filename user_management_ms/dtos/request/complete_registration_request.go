package request

import "time"

type CompleteRegisterRequest struct {
	UserId    uint       `json:"user_id"`
	BirthDate *time.Time `json:"birth_date" validate:"required"`
	Password  string     `json:"password" validate:"required"`
}

type BirthDateAndPassword struct {
	BirthDate *time.Time `json:"birth_date" validate:"required"`
	Password  string     `json:"password" validate:"required,password"`
}
