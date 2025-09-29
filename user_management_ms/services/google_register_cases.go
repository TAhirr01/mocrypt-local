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

type RegistrationCase interface {
	Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error)
}

// AlreadyVerifiedCase  1.Already Verified
type AlreadyVerifiedCase struct {
}

func (c AlreadyVerifiedCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone != "" && user.Phone == req.Phone && user.PhoneVerified && user.EmailVerified && user.PINHash != "" {
		return &response.GoogleResponse{
			UserId:        user.Id,
			Email:         user.Email,
			Phone:         user.Phone,
			Status:        response.VERIFIED,
			PhoneVerified: user.PhoneVerified,
		}, nil
	}
	return nil, nil
}

// PhoneUnverifiedCase  2.Phone exists but unverified

type PhoneUnverifiedCase struct {
	otp IOtp
}

func (c PhoneUnverifiedCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone != "" && user.Phone == req.Phone && !user.PhoneVerified {
		_, _ = c.otp.SendPhoneOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone})
		return &response.GoogleResponse{
			UserId:        user.Id,
			Email:         user.Email,
			Phone:         user.Phone,
			Status:        response.UNVERIFIED,
			PhoneVerified: user.PhoneVerified,
		}, nil
	}
	return nil, nil
}

// 3. User has no phone attach new phone
type AttachPhoneCase struct {
	query   query_repository.IUserQueryRepository
	command command_repository.IUserCommandRepository
	db      *gorm.DB
	otp     IOtp
}

func (c AttachPhoneCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone == "" {
		isExists, err := c.query.IsUserWithPhoneExists(c.db, req.Phone)
		if err != nil {
			return nil, err
		}
		if isExists {
			return nil, errors.New(string(response.PHONE_EXISTS))
		}
		_, err = c.otp.SendPhoneOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone})
		if err != nil {
			return nil, err
		}
		return &response.GoogleResponse{
			UserId:        user.Id,
			Email:         user.Email,
			Phone:         req.Phone,
			Status:        response.VERIFICATION_PENDING,
			PhoneVerified: user.PhoneVerified,
		}, nil
	}
	return nil, nil
}

// 4. Phone mismatch
type PhoneMismatchCase struct{}

func (c PhoneMismatchCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone != "" && user.Phone != req.Phone {
		return &response.GoogleResponse{
			UserId: user.Id,
			Email:  user.Email,
			Phone:  user.Phone,
			Status: response.PHONE_MISMATCH,
		}, nil
	}
	return nil, nil
}
