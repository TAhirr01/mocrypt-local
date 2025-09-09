package services

import (
	"errors"
	"log"
	"time"
	"user_management_ms/config"
	"user_management_ms/domain"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository"
	"user_management_ms/util"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IUserService interface {
	RegisterRequestOTP(request *request.OTPRequest) (*response.RegisterResponse, error)
	VerifyRegisterOTP(otRequest *request.VerifyOTPRequest) (*response.OTPResponse, error)
	CompleteRegistration(registerRequest *request.CompleteRegisterRequest) (*response.Tokens, error)
	SendOTP(req *request.OTPRequest) (*response.SendOTPResponse, error)
	VerifyLoginOTP(otRequest *request.VerifyOTPRequest) (*response.Tokens, error)
	LoginLocal(req *request.LoginLocalRequest) (*response.LoginResponse, error)
	RefreshToken(req *request.RefreshTokenReq) (*response.Tokens, error)
	//ShouldForceFullAuth(req *request.CheckLogin, limit time.Duration) bool
}

type UserService struct {
	db    *gorm.DB
	redis IRedisService
	repo  repository.IUserRepository
	jwt   IJWTService
}

func NewUserService(db *gorm.DB, repo repository.IUserRepository, redis IRedisService, jwt IJWTService) IUserService {
	return &UserService{db: db, repo: repo, redis: redis, jwt: jwt}
}

func (u *UserService) RegisterRequestOTP(req *request.OTPRequest) (*response.RegisterResponse, error) {
	user, err := u.repo.GetUserWithEmailAndPhoneNumber(u.db, req.Email, req.Phone)
	emailOtp := util.GenerateOTP()
	phoneOtp := util.GenerateOTP()
	if err == nil {
		// User mövcuddur
		if user.EmailVerified && user.PhoneVerified && user.Password == "" {
			// User OTP verified amma registration tamamlanmayıb
			return &response.RegisterResponse{
				Email:         user.Email,
				Phone:         user.Phone,
				EmailVerified: user.EmailVerified,
				PhoneVerified: user.PhoneVerified,
				Completed:     false,
				Status:        "verified",
			}, nil
		} else if user.EmailVerified && user.PhoneVerified && user.Password != "" {
			// User OTP verified və password mövcuddur → login lazımdır
			return &response.RegisterResponse{
				Email:         user.Email,
				Phone:         user.Phone,
				Status:        "verified",
				EmailVerified: user.EmailVerified,
				PhoneVerified: user.PhoneVerified,
				Completed:     true,
			}, nil
		} else if !(user.EmailVerified && user.PhoneVerified) {
			// User mövcuddur amma OTP verified deyil → OTP göndərilməlidir
			if err := u.repo.SetUserEmailPhoneOtpAndExpireDates(u.db, user, emailOtp, phoneOtp); err != nil {
				return nil, err
			}
			if err := SendVerifyEmailAndPhoneNumberEvent(
				&request.VerifyEmailEvent{Email: user.Email},
				&request.VerifyPhoneEvent{Phone: user.Phone},
			); err != nil {
				return nil, err
			}
			return &response.RegisterResponse{
				Email:         user.Email,
				Phone:         user.Phone,
				EmailVerified: user.EmailVerified,
				PhoneVerified: user.PhoneVerified,
				Completed:     false,
				Status:        "verification_pending",
			}, nil
		}
	} else {
		existingUser, err := u.repo.GetUserByEmailOrPhone(u.db, req.Email, req.Phone)
		if existingUser != nil && err == nil {
			return &response.RegisterResponse{
				Email:  req.Email,
				Phone:  req.Phone,
				Status: "exists",
			}, errors.New("user with this email or phone number already exists")
		}
		// User yoxdur → yeni user yarat
		newUser := &domain.User{
			Email: req.Email,
			Phone: req.Phone,
		}
		if _, err := u.repo.Create(u.db, newUser); err != nil {
			return nil, err
		}
		if err := u.repo.SetUserEmailPhoneOtpAndExpireDates(u.db, newUser, emailOtp, phoneOtp); err != nil {
			return nil, err
		}
		if err := SendVerifyEmailAndPhoneNumberEvent(
			&request.VerifyEmailEvent{Email: newUser.Email},
			&request.VerifyPhoneEvent{Phone: newUser.Phone},
		); err != nil {
			return nil, err
		}

		return &response.RegisterResponse{
			Email:         req.Email,
			Phone:         req.Phone,
			Status:        "created",
			EmailVerified: false,
			PhoneVerified: false,
			Completed:     false,
		}, nil
	}

	// Heç bir şərt uyğun gəlmirsə (nəzəri olaraq mümkün deyil)
	return nil, errors.New("unable to process OTP request")
}

func (u *UserService) VerifyRegisterOTP(otRequest *request.VerifyOTPRequest) (*response.OTPResponse, error) {
	user, err := u.repo.GetUserWithEmailAndPhoneNumber(u.db, otRequest.Email, otRequest.Phone)
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
	if err := u.repo.DeteUserOtpAndExpireDate(u.db, user); err != nil {
		return nil, err
	}

	return &response.OTPResponse{
		Email:         user.Email,
		Phone:         user.Phone,
		Status:        "otp_verified",
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
	}, nil
}

func (u *UserService) CompleteRegistration(req *request.CompleteRegisterRequest) (*response.Tokens, error) {
	// 1. Check if user exists
	user, err := u.repo.GetUserByEmail(u.db, req.Email)
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
	user, err = u.repo.UpdateUserPasswordAndBirthDate(
		u.db,
		req.Email,
		string(hashedPassword),
		req.BirthDate,
	)
	if err != nil {
		return nil, err
	}

	// 5. Generate tokens
	tokens, err := u.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return &response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (u *UserService) SendOTP(req *request.OTPRequest) (*response.SendOTPResponse, error) {
	if err := u.repo.SaveUserOTPs(u.db, req.Email, req.Phone, 5*time.Minute); err != nil {
		return nil, err
	}
	return &response.SendOTPResponse{
		Email:  req.Email,
		Phone:  req.Phone,
		Status: "otp_sent",
	}, nil
}

func (u *UserService) LoginLocal(req *request.LoginLocalRequest) (*response.LoginResponse, error) {
	user, err := u.repo.GetUserWithEmailAndPhoneNumber(u.db, req.Email, req.Phone)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, errors.New("invalid password")
	}

	// OTP yenidən generate və Kafka event
	emailOTP := util.GenerateOTP()
	phoneOTP := util.GenerateOTP()
	if err := u.repo.SetUserEmailPhoneOtpAndExpireDates(u.db, user, emailOTP, phoneOTP); err != nil {
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
		Email: user.Email,
		Phone: user.Phone,
	}, nil
}

func (u *UserService) VerifyLoginOTP(otRequest *request.VerifyOTPRequest) (*response.Tokens, error) {
	user, err := u.repo.GetUserWithEmailAndPhoneNumber(u.db, otRequest.Email, otRequest.Phone)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if time.Now().After(*user.EmailOtpExpireDate) || user.EmailOtp != otRequest.EmailOTP {
		return nil, errors.New("email OTP invalid or expired")
	}

	if time.Now().After(*user.PhoneOtpExpireDate) || user.PhoneOtp != otRequest.PhoneOTP {
		return nil, errors.New("phone OTP invalid or expired")
	}

	if err := u.repo.DeteUserOtpAndExpireDate(u.db, user); err != nil {
		return nil, err
	}

	// OTP-lər doğru → token yarat
	tokens, err := u.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	// Refresh token Redis-ə set edilir
	if err := u.redis.SetRefreshToken(user.Id, tokens.RefreshToken); err != nil {
		log.Println("Failed to store refresh token:", err)
	}

	return &response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (u *UserService) RefreshToken(req *request.RefreshTokenReq) (*response.Tokens, error) {
	log.Println("Someone tries to refresh access token")
	if req.RefreshToken == "" {
		return nil, errors.New("empty refresh token")
	}

	token, err := u.jwt.ParseJWT(req.RefreshToken)
	if err != nil || token == nil {
		return nil, err
	}

	claims, err := u.jwt.GetClaims(token)
	if err != nil {
		return nil, err
	}
	userID := uint(claims["sub"].(float64))

	storedToken, err := u.redis.GetRefreshToken(userID)
	if err != nil {
		log.Println("Redis token not found:", err)
		return nil, errors.New("refresh token not found or expired")
	}

	if storedToken != req.RefreshToken {
		log.Println("Provided:", req.RefreshToken, "Stored:", storedToken)
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
		log.Println("Failed to store new refresh token")
		return nil, errors.New("could not store new refresh token")
	}

	log.Println("Successfully sent a new refresh token")
	return &response.Tokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

//func (u *UserService) HasUserCompletedOtpVerification(email string) (bool, error) {
//	user, _ := u.repo.GetUserByEmail(u.db, email)
//	if user == nil {
//		return false, errors.New("user not found")
//	}
//	if !user.OTPVerified {
//		return false, nil
//	}
//	return true, nil
//}
