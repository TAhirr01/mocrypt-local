package services

import (
	"errors"
	"fmt"
	"log"
	"time"
	"user_management_ms/config"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"
	"user_management_ms/util"

	"github.com/hashicorp/go-uuid"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IUserService interface {
	RegisterRequestOTP(request *request.StartRegistration) (*response.RegisterResponse, error)
	CompleteRegistration(registerRequest *request.CompleteRegisterRequest) (*response.AfterRegisterPassword, error)
	VerifyLoginOTP(otRequest *request.VerifyOTPRequest) (*response.AfterLoginVerification, error)
	LoginLocal(req *request.LoginLocalRequest) (*response.LoginResponse, error)
	RefreshToken(req *request.RefreshTokenReq) (*response.Tokens, error)
	Setup2FA(userId uint) (*response.TwoFASetupResponse, error)
	Verify2FA(userId uint, code string) (bool, error)
	RequestLoginQr() ([]byte, string, error)
	ApproveLoginQr(userId uint, sessionId string) error
	CheckLoginQr(sessionId string) (*response.QrLoginResponse, error)
}

type UserService struct {
	db      *gorm.DB
	redis   IRedisService
	command command_repository.IUserCommandRepository
	query   query_repository.IUserQueryRepository
	otp     IOtp
	jwt     IJWTService
}

func NewUserService(db *gorm.DB, redis IRedisService, otp IOtp, command command_repository.IUserCommandRepository, query query_repository.IUserQueryRepository, jwt IJWTService) IUserService {
	return &UserService{db: db, redis: redis, otp: otp, command: command, query: query, jwt: jwt}
}

func (u *UserService) RegisterRequestOTP(req *request.StartRegistration) (*response.RegisterResponse, error) {
	user, err := u.query.GetUserWithEmailAndPhone(u.db, req.Email, req.Phone)

	if err == nil {
		cases := []UserRegisterCases{
			HasntCompleted{},
			SendLogin{},
			NeedsVerification{otp: u.otp},
			SetPin{},
		}
		for _, c := range cases {
			if resp, err := c.Handle(user, req); resp != nil || err != nil {
				return resp, err
			}
		}
	} else {
		cases := []UserRegisterCases{
			ExistingUser{query: u.query, command: u.command, otp: u.otp, db: u.db},
		}
		for _, c := range cases {
			if resp, err := c.Handle(user, req); resp != nil || err != nil {
				return resp, err
			}
		}
	}
	return nil, errors.New("unable to process OTP request")
}

func (u *UserService) CompleteRegistration(req *request.CompleteRegisterRequest) (*response.AfterRegisterPassword, error) {
	// 1. Check if user exists
	user, err := u.query.GetByID(u.db, req.UserId)
	if err != nil {
		return nil, err
	}

	// 2. Ensure phone is verified
	if !user.PhoneVerified {
		return nil, errors.New("phone number not verified, complete phone verification first")
	}
	if user.Password != "" {
		return nil, errors.New("registration already completed")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. Update birthday + password
	user, err = u.command.UpdateUserPasswordAndBirthDateById(
		u.db,
		req.UserId,
		string(hashedPassword),
		req.BirthDate,
	)
	if err != nil {
		return nil, err
	}

	return &response.AfterRegisterPassword{
		UserId: user.Id,
		Status: response.SET_PIN,
	}, nil
}

func (u *UserService) LoginLocal(req *request.LoginLocalRequest) (*response.LoginResponse, error) {
	user, err := u.query.GetUserWithEmailAndPhone(u.db, req.Email, req.Phone)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, errors.New("invalid password")
	}

	// OTP yenidən generate və Kafka event
	emailOTP := util.GenerateOTP()
	phoneOTP := util.GenerateOTP()
	if err := u.command.SetUserEmailPhoneOtpAndExpireDates(u.db, user, emailOTP, phoneOTP); err != nil {
		return nil, err
	}
	if err := SendVerifyEmailEventToKafka(&request.VerifyEmailEvent{
		Email:    user.Email,
		EmailOTP: emailOTP,
	}); err != nil {
		log.Println("Failed to send email event:", err)
	}

	if err := SendVerifyPhoneNumberEventToKafka(&request.VerifyPhoneEvent{
		Phone:    user.Phone,
		PhoneOTP: phoneOTP,
	}); err != nil {
		log.Println("Failed to send phone event:", err)
	}

	return &response.LoginResponse{
		UserId: user.Id,
		Email:  user.Email,
		Phone:  user.Phone,
	}, nil
}

func (u *UserService) VerifyLoginOTP(otRequest *request.VerifyOTPRequest) (*response.AfterLoginVerification, error) {
	user, err := u.query.GetByID(u.db, otRequest.UserId)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	log.Println(user)

	if user.EmailOtpExpireDate != nil && user.PhoneOtpExpireDate != nil {
		if time.Now().After(*user.EmailOtpExpireDate) || user.EmailOtp != otRequest.EmailOTP {
			return nil, errors.New("email OTP invalid or expired")
		}

		if time.Now().After(*user.PhoneOtpExpireDate) || user.PhoneOtp != otRequest.PhoneOTP {
			return nil, errors.New("phone OTP invalid or expired")
		}
	} else if user.EmailOtpExpireDate == nil && user.PhoneOtpExpireDate == nil {
		if user.EmailVerified && user.PhoneVerified {
			return nil, errors.New("user already verified")
		} else {
			return nil, errors.New("user needs to be verified")
		}
	}

	if err := u.command.DeleteUserOtpAndExpireDate(u.db, user); err != nil {
		return nil, err
	}

	if user.PINHash != "" {
		return &response.AfterLoginVerification{
			UserId: user.Id,
			Status: response.VERIFY_PIN,
		}, nil
	} else {
		return &response.AfterLoginVerification{
			UserId: user.Id,
			Status: response.SET_PIN,
		}, nil
	}
}

func (u *UserService) RefreshToken(req *request.RefreshTokenReq) (*response.Tokens, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("empty refresh token")
	}

	token, err := u.jwt.ParseJWT(req.RefreshToken)
	if err != nil || token == nil {
		return nil, errors.New("invalid refresh token")
	}

	claims, err := u.jwt.GetClaims(token)
	if err != nil {
		return nil, err
	}
	userID := uint(claims["sub"].(float64))

	storedToken, err := u.redis.GetRefreshToken(userID)
	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	if storedToken != req.RefreshToken {
		return nil, errors.New("refresh token does not equal to stored token")
	}

	newAccessToken, err := u.jwt.GenerateToken(userID, time.Duration(config.Conf.Application.Security.TokenValidityInSeconds)*time.Second)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	newRefreshToken, err := u.jwt.GenerateToken(userID, time.Duration(config.Conf.Application.Security.TokenValidityInSecondsForRememberMe)*time.Second)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	u.redis.DelRefreshToken(userID)

	if err := u.redis.SetRefreshToken(userID, newRefreshToken); err != nil {
		return nil, errors.New("could not store new refresh token")
	}

	return &response.Tokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *UserService) Setup2FA(userId uint) (*response.TwoFASetupResponse, error) {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return nil, err
	}
	if user.Is2FAVerified {
		return nil, errors.New("user already has 2FA verified")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.Conf.Application.Security.Issuer,
		AccountName: user.Email,
	})
	if err != nil {
		return nil, err
	}

	user.Google2FASecret = key.Secret()
	if err := u.command.Update(u.db, user); err != nil {
		return nil, err
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return &response.TwoFASetupResponse{
		Secret: key.Secret(),
		QRCode: png,
	}, nil
}

func (u *UserService) Verify2FA(userId uint, code string) (bool, error) {
	user, err := u.query.GetByID(u.db, userId)
	if err != nil {
		return false, err
	}
	valid := totp.Validate(code, user.Google2FASecret)
	if valid {
		user.Is2FAVerified = true
		err := u.command.Update(u.db, user)
		if err != nil {
			log.Println("Failed to update user:", err)
		}
	}
	return valid, nil
}

func (u *UserService) RequestLoginQr() ([]byte, string, error) {
	sessionId, _ := uuid.GenerateUUID()
	err := u.redis.StoreLoginSessionRedis(sessionId)
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("https://mocadomain.com/qr-login?sessionId=%s", sessionId)
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, "", err
	}

	return png, sessionId, nil
}

func (u *UserService) ApproveLoginQr(userId uint, sessionId string) error {
	session, err := u.redis.GetLoginSessionRedis(sessionId)
	if err != nil {
		return errors.New("session not found or redis problem")
	}
	session.UserId = userId
	session.Status = "APPROVED"

	if err := u.redis.UpdateLoginSessionRedis(sessionId, session); err != nil {
		return err
	}
	return nil
}

func (u *UserService) CheckLoginQr(sessionId string) (*response.QrLoginResponse, error) {
	session, err := u.redis.GetLoginSessionRedis(sessionId)
	if err != nil {
		return &response.QrLoginResponse{Status: response.StatusExpired}, nil
	}

	switch session.Status {
	case "PENDING":
		return &response.QrLoginResponse{Status: response.StatusPending}, nil
	case "APPROVED":
		user, err := u.query.GetByID(u.db, session.UserId)
		if err != nil {
			return nil, err
		}
		tokens, err := u.jwt.GenerateTokens(user)
		if err != nil {
			return nil, err
		}

		// Once consumed, delete to avoid reuse
		_ = u.redis.DeleteLoginSessionRedis(sessionId)

		return &response.QrLoginResponse{
			Status: response.StatusApproved,
			Tokens: tokens,
		}, nil

	default:
		return &response.QrLoginResponse{Status: response.StatusExpired}, nil
	}
}
