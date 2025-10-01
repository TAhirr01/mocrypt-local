package services

import (
	"errors"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"
	"user_management_ms/util"

	"gorm.io/gorm"
)

type IPinService interface {
	SetPIN(userId uint, request *request.PinReq) error
	VerifyPIN(userId uint, request *request.PinReq) (*response.Tokens, bool, error)
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

func (u *PinService) SetPIN(userId uint, req *request.PinReq) error {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return err
	}

	hashed, err := util.HashPIN(req.Pin)
	if err != nil {
		return err
	}

	user.PINHash = hashed
	return u.command.Update(u.db, user)
}

func (u *PinService) VerifyPIN(userId uint, req *request.PinReq) (*response.Tokens, bool, error) {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return nil, false, err
	}
	if !user.Loginable {
		return nil, false, errors.New("verify login first")
	}

	if user.PINHash == "" {
		return nil, false, errors.New("PIN not set")
	}
	valid := util.VerifyPIN(req.Pin, user.PINHash)
	if !valid {
		return nil, false, errors.New("invalid PIN")
	}

	token, err := u.jwt.GenerateTokens(user)

	return token, true, err
}
