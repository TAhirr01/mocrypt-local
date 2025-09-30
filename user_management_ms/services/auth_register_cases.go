package services

import (
	"errors"
	"user_management_ms/domain"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"

	"gorm.io/gorm"
)

const (
	CREATED              = "created"
	SET_PIN              = "set_pin"
	VERIFICATION_PENDING = "verification_pending"
)

type UserRegisterCases interface {
	Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error)
}
type HasntCompleted struct {
}

func (u HasntCompleted) Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error) {
	if user.EmailVerified && user.PhoneVerified && user.Password == "" {
		// User OTP verified amma registration tamamlanmayıb
		return &response.RegisterResponse{
			UserId:   user.Id,
			UserType: user.UserType,
			Email:    user.Email,
			Phone:    user.Phone,
			Status:   PASSWORD_PENDING,
		}, nil
	}
	return nil, nil
}

type SendLogin struct{}

func (u SendLogin) Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error) {
	if user.EmailVerified && user.PhoneVerified && user.Password != "" && user.PINHash != "" {
		return nil, errors.New("User already exists login ")
	}
	return nil, nil
}

type NeedsVerification struct {
	otp IOtp
}

func (u NeedsVerification) Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error) {
	if !(user.EmailVerified && user.PhoneVerified) {
		// User mövcuddur amma OTP verified deyil → OTP göndərilməlidir
		resp, err := u.otp.SendOTP(&request.OTPRequest{Email: req.Email, Phone: req.Phone})
		if err != nil {
			return nil, err
		}
		return &response.RegisterResponse{
			UserId:   user.Id,
			UserType: user.UserType,
			Email:    resp.Email,
			Phone:    resp.Phone,
			Status:   VERIFICATION_PENDING,
		}, nil
	}
	return nil, nil
}

type ExistingUser struct {
	query   query_repository.IUserQueryRepository
	command command_repository.IUserCommandRepository
	otp     IOtp
	db      *gorm.DB
}

func (u ExistingUser) Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error) {
	existingUser, err := u.query.GetUserByEmailOrPhone(u.db, req.Email, req.Phone)
	if existingUser != nil && err == nil {
		return nil, errors.New("user with this email or phone number already exists login")
	}
	newUser := &domain.User{
		UserType: req.UserType,
		Email:    req.Email,
		Phone:    req.Phone,
	}
	if _, err := u.command.Create(u.db, newUser); err != nil {
		return nil, err
	}
	sendOTP, err := u.otp.SendOTP(&request.OTPRequest{Email: req.Email, Phone: req.Phone})
	if err != nil {
		return nil, err
	}
	return &response.RegisterResponse{
		UserId:   newUser.Id,
		UserType: newUser.UserType,
		Email:    sendOTP.Email,
		Phone:    sendOTP.Phone,
		Status:   CREATED,
	}, nil
}

type SetPin struct{}

func (u SetPin) Handle(user *domain.User, req *request.StartRegistration) (*response.RegisterResponse, error) {
	if user.EmailVerified && user.PhoneVerified && user.Password != "" && user.PINHash == "" {
		return &response.RegisterResponse{
			UserId:   user.Id,
			UserType: user.UserType,
			Email:    user.Email,
			Phone:    user.Phone,
			Status:   SET_PIN,
		}, nil
	}
	return nil, nil
}
