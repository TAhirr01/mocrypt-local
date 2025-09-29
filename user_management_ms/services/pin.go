package services

import (
	"errors"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"
	"user_management_ms/util"

	"gorm.io/gorm"
)

type IPinService interface {
	SetPIN(userId uint, pin string) error
	VerifyPIN(userId uint, pin string) (*response.Tokens, bool, error)
}

type PinService struct {
	query   query_repository.IUserQueryRepository
	db      *gorm.DB
	command command_repository.IUserCommandRepository
	jwt     IJWTService
}

func NewPinService(query query_repository.IUserQueryRepository, command command_repository.IUserCommandRepository, db *gorm.DB, jwt IJWTService) IPinService {
	return &PinService{query: query, db: db, command: command, jwt: jwt}
}

func (u *PinService) SetPIN(userId uint, pin string) error {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return err
	}

	hashed, err := util.HashPIN(pin)
	if err != nil {
		return err
	}

	user.PINHash = hashed
	return u.command.Update(u.db, user)
}

func (u *PinService) VerifyPIN(userId uint, pin string) (*response.Tokens, bool, error) {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return nil, false, err
	}

	if user.PINHash == "" {
		return nil, false, errors.New("PIN not set")
	}
	valid := util.VerifyPIN(pin, user.PINHash)
	if !valid {
		return nil, false, errors.New("invalid PIN")
	}

	token, err := u.jwt.GenerateTokens(user)

	return token, true, err
}
