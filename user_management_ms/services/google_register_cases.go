package services

import (
	"errors"
	"log"
	"user_management_ms/domain"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
)

type RegistrationCase interface {
	Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error)
}

// AlreadyVerifiedCase  1.Already Verified
type AlreadyVerifiedCase struct {
}

func (c AlreadyVerifiedCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone != "" && user.Phone == req.Phone && user.PhoneVerified && user.EmailVerified {
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
	svc *GoogleAuthService
}

func (c PhoneUnverifiedCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone != "" && user.Phone == req.Phone && !user.PhoneVerified {
		_, _ = c.svc.SendPhoneVerificationOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone})
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

// 3. User has no phone → attach new phone
type AttachPhoneCase struct {
	svc *GoogleAuthService
}

func (c AttachPhoneCase) Handle(user *domain.User, req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	if user.Phone == "" {
		isExists, err := c.svc.query.IsUserWithPhoneExists(c.svc.db, req.Phone)
		if err != nil {
			return nil, err
		}
		if isExists {
			return nil, errors.New("user with this phone already exists")
		}
		_, err = c.svc.SendPhoneVerificationOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone})
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
		log.Println("Case: User exists but requested phone is not user's")
		return &response.GoogleResponse{
			UserId: user.Id,
			Email:  user.Email,
			Phone:  user.Phone,
			Status: response.PHONE_MISMATCH,
		}, nil
	}
	return nil, nil
}
