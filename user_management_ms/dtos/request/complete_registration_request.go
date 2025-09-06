package request

import "time"

type CompleteRegisterRequest struct {
	Email     string     `json:"email" validate:"required,email"`
	BirthDate *time.Time `json:"birth_date" validate:"required"`
	Password  string     `json:"password" validate:"required"`
}

type CompleteGoogleRegistration struct {
	Email     string     `json:"email" validate:"required"`
	BirthDate *time.Time `json:"birth_date" validate:"required"`
	Password  string     `json:"password" validate:"required"`
}

type BirthDateAndPassword struct {
	BirthDate *time.Time `json:"birth_date" validate:"required"`
	Password  string     `json:"password" validate:"required"`
}
