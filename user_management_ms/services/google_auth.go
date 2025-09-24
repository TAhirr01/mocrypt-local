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
	"user_management_ms/repository"
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
}

type GoogleAuthService struct {
	db         *gorm.DB
	oauthConf  *oauth2.Config
	jwt        IJWTService
	googleRepo repository.IGoogleRepository
	redis      IRedisService
}

func NewGoogleAuthService(db *gorm.DB, oauthConf *oauth2.Config, googleRepo repository.IGoogleRepository, jwtService IJWTService, rdb IRedisService) IGoogleAuthService {
	return &GoogleAuthService{googleRepo: googleRepo, oauthConf: oauthConf, jwt: jwtService, db: db, redis: rdb}
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
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Phone == "" {
		return nil, errors.New("phone is required")
	}

	// Try to find user by email only (not email+phone) — this fixes the unreachable-case bug.
	user, err := g.googleRepo.FindUserByEmail(g.db, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// If user exists:
	if user != nil {
		// Case 1: If user has a phone and input phone is his and already verified
		// If both phone and email are already verified -> do nothing, return verified status.
		// NOTE: this code assumes you have both PhoneVerified and EmailVerified fields.
		if user.Phone != "" && user.Phone == req.Phone && user.PhoneVerified && user.EmailVerified {
			log.Println("User already fully verified")
			return &response.GoogleResponse{
				Email:         user.Email,
				Phone:         user.Phone,
				Status:        "verified",
				PhoneVerified: user.PhoneVerified,
			}, nil
		}
		// Case 2: If user has a phone but input phone is not his
		// If user has a phone, and it's the same as requested phone
		if user.Phone != "" && user.Phone == req.Phone {
			// If phone exists but not verified -> resend phone OTP

		}
		// Case 2: User don't have a phone yet attach a phone to user
		// If user exists, but they have no phone yet (we want to attach req.Phone)
		if user.Phone == "" {
			// check whether another user already uses requested phone
			isExists, err := g.googleRepo.IsUserWithPhoneExists(g.db, req.Phone)
			if err != nil {
				return nil, err
			}
			if isExists {
				return nil, errors.New("user with this phone already exists")
			}

			// attach phone and create phone OTP
			otp := util.GenerateOTP()
			expire := time.Now().Add(5 * time.Minute)

			updatedUser, err := g.googleRepo.UpdateGoogleUserPhone(g.db, req.Email, req.Phone, otp, expire)
			if err != nil {
				return nil, err
			}

			// send OTP via kafka
			if err := SendVerifyPhoneNumberEventToKafka(&request.VerifyPhoneEvent{
				Phone:    updatedUser.Phone,
				PhoneOTP: otp,
			}); err != nil {
				return nil, err
			}

			return &response.GoogleResponse{
				Email:         updatedUser.Email,
				Phone:         updatedUser.Phone,
				Status:        "phone_verification_pending",
				PhoneVerified: updatedUser.PhoneVerified,
			}, nil
		}

		// Case 3: Users phone is different that what user requested
		// If user exists but their phone is different from requested -> phone_mismatch
		if user.Phone != "" && user.Phone != req.Phone {
			log.Println("Case: User exists but requested phone is not user's")
			return &response.GoogleResponse{
				Email:  user.Email,
				Phone:  user.Phone,
				Status: "phone_mismatch",
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
	user, err := g.googleRepo.FindUserByGoogleId(g.db, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g *GoogleAuthService) VerifyPhoneOTP(req *request.VerifyNumberOTPRequest) (*response.OTPResponsePhone, error) {
	user, err := g.googleRepo.GetUserWithEmailAndPhone(g.db, req.Email, req.Phone)
	if err != nil {
		return nil, err
	}
	if user.PhoneOtp != req.PhoneOTP || time.Now().After(*user.PhoneOtpExpireDate) {
		return nil, errors.New("invalid or expired OTP")
	}
	user.PhoneVerified = true
	user.PhoneOtp = ""
	user.PhoneOtpExpireDate = nil

	if _, err := g.googleRepo.Update(g.db, user); err != nil {
		return nil, err
	}
	return &response.OTPResponsePhone{
		Phone:         req.Phone,
		PhoneVerified: user.PhoneVerified,
		Status:        "otp_verified",
		Message:       "Completion of registration is needed ",
	}, nil
}

func (g *GoogleAuthService) CompleteGoogleRegistration(req *request.CompleteGoogleRegistration) (*response.Tokens, error) {
	// 1. Check if user exists
	user, err := g.googleRepo.FindUserByEmail(g.db, req.Email)
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
	user, err = g.googleRepo.UpdateUserBirthdayAndPassword(
		g.db,
		req.Email,
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
	user, err := g.googleRepo.FindUserByEmail(g.db, req.Email)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.EmailOtp = otp
	user.EmailOtpExpireDate = &expire
	if _, err := g.googleRepo.Update(g.db, user); err != nil {
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
	user, err := g.googleRepo.FindUserByEmail(g.db, req.Email)
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
	if _, err := g.googleRepo.Update(g.db, user); err != nil {
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
	user, err := g.googleRepo.FindUserByEmail(g.db, email)
	if err == nil && user != nil {
		// User exists → link Google ID
		if user.GoogleID == "" {
			user.GoogleID = googleId
			if _, err := g.googleRepo.Update(g.db, user); err != nil {
				return nil, false, err
			}
		}
		// return (user, false=new user created?, nil)
		return user, false, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No user exists → create new user
		newUser, err := g.googleRepo.Create(g.db, &domain.User{Email: email, GoogleID: googleId})
		if err != nil {
			return nil, false, err
		}
		return newUser, true, nil
	}
	return nil, false, err
}

func (g *GoogleAuthService) SendPhoneVerificationOtp(req *request.OTPRequestPhone) (*response.OTPResponsePhone, error) {
	user, err := g.googleRepo.FindUserByPhoneNumber(g.db, req.Phone)
	if err != nil {
		return nil, err
	}
	otp := util.GenerateOTP()
	expire := time.Now().Add(5 * time.Minute)
	user.PhoneOtp = otp
	user.PhoneOtpExpireDate = &expire
	if _, err := g.googleRepo.Update(g.db, user); err != nil {
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
