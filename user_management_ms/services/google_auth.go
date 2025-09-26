package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	"user_management_ms/config"
	"user_management_ms/domain"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository/command_repository"
	"user_management_ms/repository/query_repository"
	"user_management_ms/util"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type IGoogleAuthService interface {
	LoginGoogle(state string) string
	ExchangeGoogleToken(code string) (*oauth2.Token, error)
	GetUserInfo(code string) (*response.GoogleUser, error)
	VerifyGoogleIDToken(idToken string) (*response.GoogleUser, error)
	FindUserByGoogleID(id string) (*domain.User, error)
	StartGoogleRegistration(req *request.StartGoogleRegistration) (*response.GoogleResponse, error)
	VerifyPhoneOTP(req *request.VerifyNumberOTPRequest) (*response.OTPResponsePhone, error)
	CompleteGoogleRegistration(req *request.CompleteGoogleRegistration) (*response.Tokens, error)
	SendEmailLoginOtp(req *request.OTPRequestEmail) (*response.OTPResponseEmail, error)
	VerifyGoogleLoginOtp(req *request.VerifyEmailOTPRequest) (*response.Tokens, error)
	CreteNewGoogleUser(email, googleId string) (*domain.User, bool, error)
	SendPhoneVerificationOtp(req *request.OTPRequestPhone) (*response.OTPResponsePhone, error)
	LoginOrRegister(isNew bool, user *domain.User) (*response.CallBackResponse, error)
}

type GoogleAuthService struct {
	db        *gorm.DB
	oauthConf *oauth2.Config
	jwt       IJWTService
	command   command_repository.IUserCommandRepository
	query     query_repository.IUserQueryRepository

	redis IRedisService
}

func NewGoogleAuthService(db *gorm.DB, oauthConf *oauth2.Config, command command_repository.IUserCommandRepository, query query_repository.IUserQueryRepository, jwtService IJWTService, rdb IRedisService) IGoogleAuthService {
	return &GoogleAuthService{oauthConf: oauthConf, query: query, command: command, jwt: jwtService, db: db, redis: rdb}
}
func (g *GoogleAuthService) LoginGoogle(state string) string {
	url := g.oauthConf.AuthCodeURL(state)
	return url
}

func (g *GoogleAuthService) ExchangeGoogleToken(code string) (*oauth2.Token, error) {
	token, err := g.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return token, err
}

func (g *GoogleAuthService) GetUserInfo(code string) (*response.GoogleUser, error) {

	token, err := g.ExchangeGoogleToken(code)
	if err != nil {
		return nil, err
	}

	client := g.oauthConf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, errors.New("failed to get user info")
	}
	defer resp.Body.Close()

	var gUser response.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		return nil, errors.New("failed to decode user info")
	}
	return &gUser, nil
}

func (g *GoogleAuthService) VerifyGoogleIDToken(idToken string) (*response.GoogleUser, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, config.Conf.Application.OAuth2.ClientID)
	if err != nil {
		return nil, err
	}
	user := &response.GoogleUser{
		ID:            payload.Claims["sub"].(string),
		Email:         payload.Claims["email"].(string),
		VerifiedEmail: payload.Claims["email_verified"].(bool),
	}

	return user, nil
}

func (g *GoogleAuthService) StartGoogleRegistration(req *request.StartGoogleRegistration) (*response.GoogleResponse, error) {
	// validate input quickly (optional but helpful)
	if req.Phone == "" {
		return nil, errors.New("phone is required")
	}

	// Try to find user by email only (not email+phone) — this fixes the unreachable-case bug.
	user, err := g.query.GetByID(g.db, req.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// If user exists:
	if user != nil {
		// Case 1: If user has a phone and input phone is his and already verified
		// If both phone and email are already verified -> do nothing, return verified status.
		// NOTE: this code assumes you have both PhoneVerified and EmailVerified fields.
		if user.Phone != "" && user.Phone == req.Phone && user.PhoneVerified && user.EmailVerified {
			return &response.GoogleResponse{
				UserId:        user.Id,
				Email:         user.Email,
				Phone:         user.Phone,
				Status:        response.VERIFIED,
				PhoneVerified: user.PhoneVerified,
			}, nil
		}
		// Case 2: If user has a phone but not verified
		// If user has a phone, and it's the same as requested phone
		if user.Phone != "" && user.Phone == req.Phone && !user.PhoneVerified {
			if _, err := g.SendPhoneVerificationOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone}); err != nil {
			}
			return &response.GoogleResponse{
				UserId:        user.Id,
				Email:         user.Email,
				Phone:         user.Phone,
				Status:        response.UNVERIFIED,
				PhoneVerified: user.PhoneVerified,
			}, nil

		}
		// Case 2: User don't have a phone yet attach a phone to user
		// If user exists, but they have no phone yet (we want to attach req.Phone)
		if user.Phone == "" {
			// check whether another user already uses requested phone
			isExists, err := g.query.IsUserWithPhoneExists(g.db, req.Phone)
			if err != nil {
				return nil, err
			}
			if isExists {
				return nil, errors.New("user with this phone already exists")
			}

			// attach phone and create phone OTP
			if _, err := g.SendPhoneVerificationOtp(&request.OTPRequestPhone{UserId: user.Id, Phone: req.Phone}); err != nil {
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
		// Case 3: Users phone is different that what user requested
		// If user exists but their phone is different from requested -> phone_mismatch
		if user.Phone != "" && user.Phone != req.Phone {
			log.Println("Case: User exists but requested phone is not user's")
			return &response.GoogleResponse{
				UserId: user.Id,
				Email:  user.Email,
				Phone:  user.Phone,
				Status: response.PHONE_MISMATCH,
			}, nil
		}
	}

	// If user not found by email, return explicit error (or create a new google-user here if you want)
	// Keep behavior explicit: currently we do not auto-create users in this flow.
	if user == nil {
		return nil, errors.New("user not found; please register first")
	}

	// default fallback (should not be reached)
	return nil, nil
}

func (g *GoogleAuthService) FindUserByGoogleID(id string) (*domain.User, error) {
	user, err := g.query.GetUserByGoogleId(g.db, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g *GoogleAuthService) VerifyPhoneOTP(req *request.VerifyNumberOTPRequest) (*response.OTPResponsePhone, error) {
	user, err := g.query.GetByID(g.db, req.UserId)
	if err != nil {
		return nil, err
	}
	if user.PhoneOtp != req.PhoneOTP || time.Now().After(*user.PhoneOtpExpireDate) {
		return nil, errors.New("invalid or expired OTP")
	}
	user.PhoneVerified = true
	user.PhoneOtp = ""
	user.PhoneOtpExpireDate = nil

	if err := g.command.Update(g.db, user); err != nil {
		return nil, err
	}
	return &response.OTPResponsePhone{
		Phone:         user.Phone,
		PhoneVerified: user.PhoneVerified,
		Status:        "otp_verified",
		Message:       "Completion of registration is needed ",
	}, nil
}

func (g *GoogleAuthService) CompleteGoogleRegistration(req *request.CompleteGoogleRegistration) (*response.Tokens, error) {
	// 1. Check if user exists
	user, err := g.query.GetByID(g.db, req.UserId)
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
	user, err = g.command.UpdateUserPasswordAndBirthDateById(
		g.db,
		req.UserId,
		string(hashedPassword),
		req.BirthDate,
	)
	if err != nil {
		return nil, err
	}

	// 5. Generate tokens
	tokens, err := g.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return &response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (g *GoogleAuthService) SendEmailLoginOtp(req *request.OTPRequestEmail) (*response.OTPResponseEmail, error) {
	user, err := g.query.GetUserByEmail(g.db, req.Email)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.EmailOtp = otp
	user.EmailOtpExpireDate = &expire
	if err := g.command.Update(g.db, user); err != nil {
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

func (g *GoogleAuthService) VerifyGoogleLoginOtp(req *request.VerifyEmailOTPRequest) (*response.Tokens, error) {
	user, err := g.query.GetUserByEmail(g.db, req.Email)
	if user != nil && (user.Password == "" || user.BirthDate == nil) {
		return nil, errors.New("user hasn't completed registration")
	}
	if err != nil {
		return nil, err
	}
	if user.EmailOtp != req.EmailOTP || time.Now().After(*user.EmailOtpExpireDate) {
		return nil, errors.New("invalid or expired OTP")
	}
	user.EmailOtp = ""
	user.EmailOtpExpireDate = nil
	if err := g.command.Update(g.db, user); err != nil {
		return nil, err
	}
	tokens, err := g.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}
	return &response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// CreteNewGoogleUser return:User,new user created,error
func (g *GoogleAuthService) CreteNewGoogleUser(email, googleId string) (*domain.User, bool, error) {
	// Check if a user already exists with this email
	user, err := g.query.GetUserByEmail(g.db, email)
	if err == nil && user != nil {
		// User exists → link Google ID
		if user.GoogleID == "" {
			user.GoogleID = googleId
			user.EmailOtp = ""
			user.EmailOtpExpireDate = nil
			user.EmailVerified = true
			if err := g.command.Update(g.db, user); err != nil {
				return nil, false, err
			}
		}
		// return (user, false=new user created?, nil)
		return user, false, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No user exists → create new user
		newUser, err := g.command.Create(g.db, &domain.User{Email: email, GoogleID: googleId, EmailVerified: true})
		if err != nil {
			return nil, false, err
		}
		return newUser, true, nil
	}
	return nil, false, err
}

func (g *GoogleAuthService) SendPhoneVerificationOtp(req *request.OTPRequestPhone) (*response.OTPResponsePhone, error) {
	user, err := g.query.GetByID(g.db, req.UserId)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.Phone = req.Phone
	user.PhoneOtp = otp
	user.PhoneOtpExpireDate = &expire
	if err := g.command.Update(g.db, user); err != nil {
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

func (g *GoogleAuthService) LoginOrRegister(isNew bool, user *domain.User) (*response.CallBackResponse, error) {
	if isNew {
		// New user → needs to complete registration
		return &response.CallBackResponse{
			UserId:   user.Id,
			Status:   "new_user",
			Phone:    user.Phone,
			Email:    user.Email,
			GoogleId: user.GoogleID,
		}, nil
	}

	//Phone hasn't verified send verification
	if isNew == false && !user.PhoneVerified {
		return &response.CallBackResponse{
			UserId:   user.Id,
			Status:   "send_request_phone",
			Phone:    user.Phone,
			Email:    user.Email,
			GoogleId: user.GoogleID,
		}, nil
	}
	if isNew == false && user.PhoneVerified && user.Password != "" {
		if _, err := g.SendEmailLoginOtp(&request.OTPRequestEmail{Email: user.Email}); err != nil {
			return nil, err
		}
		return &response.CallBackResponse{
			UserId:   user.Id,
			Status:   "send_login",
			Phone:    user.Phone,
			Email:    user.Email,
			GoogleId: user.GoogleID,
		}, nil
	}
	if isNew == false && user.PhoneVerified && user.Password == "" {
		return &response.CallBackResponse{
			UserId:   user.Id,
			Status:   "send_completion",
			Phone:    user.Phone,
			Email:    user.Email,
			GoogleId: user.GoogleID,
		}, nil
	}
	return nil, errors.New("shouldn't happen")
}
