package services

import (
	"errors"
	"time"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"
	"user_management_ms/util"

	"gorm.io/gorm"
)

type IOtp interface {
	VerifyPhoneOTP(req *request.VerifyNumberOTPRequest) (*response.OTPResponsePhone, error)
	SendEmailOtp(req *request.OTPRequestEmail) (*response.OTPResponseEmail, error)
	SendPhoneOtp(req *request.OTPRequestPhone) (*response.OTPResponsePhone, error)
	SendOTP(req *request.OTPRequest) (*response.SendOTPResponse, error)
	VerifyOTPs(otRequest *request.VerifyOTPRequest) (*response.OTPResponse, error)
}

type Otp struct {
	db      *gorm.DB
	query   query_repository.IUserQueryRepository
	command command_repository.IUserCommandRepository
}

func NewOtpService(db *gorm.DB, query query_repository.IUserQueryRepository, command command_repository.IUserCommandRepository) IOtp {
	return &Otp{db: db, query: query, command: command}
}

func (o *Otp) VerifyPhoneOTP(req *request.VerifyNumberOTPRequest) (*response.OTPResponsePhone, error) {
	user, err := o.query.GetByID(o.db, req.UserId)
	if err != nil {
		return nil, err
	}
	if user.PhoneOtp != req.PhoneOTP || time.Now().After(*user.PhoneOtpExpireDate) {
		return nil, errors.New("invalid or expired OTP")
	}
	user.PhoneVerified = true
	user.PhoneOtp = ""
	user.PhoneOtpExpireDate = nil

	if err := o.command.Update(o.db, user); err != nil {
		return nil, err
	}
	return &response.OTPResponsePhone{
		Phone:         user.Phone,
		PhoneVerified: user.PhoneVerified,
		Status:        "otp_verified",
		Message:       "Completion of registration is needed ",
	}, nil
}

func (o *Otp) SendEmailOtp(req *request.OTPRequestEmail) (*response.OTPResponseEmail, error) {
	user, err := o.query.GetByID(o.db, req.UserId)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.EmailOtp = otp
	user.EmailOtpExpireDate = &expire
	if err := o.command.Update(o.db, user); err != nil {
		return nil, err
	}
	if err := SendVerifyEmailEventToKafka(&request.VerifyEmailEvent{Email: req.Email, EmailOTP: otp}); err != nil {
		return nil, err
	}
	return &response.OTPResponseEmail{
		Email:   req.Email,
		Status:  "otp_sent",
		Message: "Email OTP sent",
	}, nil
}

func (o *Otp) SendOTP(req *request.OTPRequest) (*response.SendOTPResponse, error) {
	if err := o.command.SaveUserOTPs(o.db, req.Email, req.Phone, 5*time.Minute); err != nil {
		return nil, err
	}
	return &response.SendOTPResponse{
		Email:  req.Email,
		Phone:  req.Phone,
		Status: "otp_sent",
	}, nil
}

func (o *Otp) SendPhoneOtp(req *request.OTPRequestPhone) (*response.OTPResponsePhone, error) {
	user, err := o.query.GetByID(o.db, req.UserId)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.Phone = req.Phone
	user.PhoneOtp = otp
	user.PhoneOtpExpireDate = &expire
	if err := o.command.Update(o.db, user); err != nil {
		return nil, err
	}
	if err := SendVerifyPhoneNumberEventToKafka(&request.VerifyPhoneEvent{Phone: req.Phone, PhoneOTP: otp}); err != nil {
		return nil, err
	}
	return &response.OTPResponsePhone{
		Phone:   req.Phone,
		Status:  "otp_sent",
		Message: "Email OTP sent",
	}, nil
}

func (o *Otp) VerifyOTPs(otRequest *request.VerifyOTPRequest) (*response.OTPResponse, error) {
	user, err := o.query.GetByID(o.db, otRequest.UserId)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if time.Now().After(*user.EmailOtpExpireDate) || user.EmailOtp != otRequest.EmailOTP {
		return nil, errors.New("email OTP invalid or expired")
	}

	if time.Now().After(*user.PhoneOtpExpireDate) || user.PhoneOtp != otRequest.PhoneOTP {
		return nil, errors.New("phone OTP invalid or expired")
	}

	user.PhoneVerified = true
	user.EmailVerified = true
	if err := o.command.DeleteUserOtpAndExpireDate(o.db, user); err != nil {
		return nil, err
	}

	return &response.OTPResponse{
		UserId:        user.Id,
		Email:         user.Email,
		Phone:         user.Phone,
		Status:        "otp_verified",
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
	}, nil
}
