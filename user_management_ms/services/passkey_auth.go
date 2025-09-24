package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"user_management_ms/dtos/request"
	"user_management_ms/dtos/response"
	"user_management_ms/repository"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/hashicorp/go-uuid"
	"gorm.io/gorm"
)

type IPasskeyService interface {
	RegisterStart(req *request.StartPasskeyRegistrationRequest) (*protocol.CredentialCreation, error)
	RegisterFinish(userID uint, r *http.Request) error
	LoginStart() (*protocol.CredentialAssertion, string, error)
	LoginFinish(sessionID string, r *http.Request) (*response.Tokens, error)
}

type PasskeyService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
	wa       *webauthn.WebAuthn
	jwt      IJWTService
	redis    IRedisService
}

func NewPasskeyService(wa *webauthn.WebAuthn, db *gorm.DB, userRepo repository.IUserRepository, redis IRedisService, jwt IJWTService) IPasskeyService {
	return &PasskeyService{wa: wa, db: db, userRepo: userRepo, redis: redis, jwt: jwt}
}

// RegisterStart start passkey registration stores temporary session inside redis
func (ps *PasskeyService) RegisterStart(req *request.StartPasskeyRegistrationRequest) (*protocol.CredentialCreation, error) {
	// 1. Fetch user + existing passkeys from DB
	user, err := ps.userRepo.GetByID(ps.db, req.UserId)

	if err != nil {
		return nil, err
	}
	if user.Password == "" || user.BirthDate == nil {
		return nil, errors.New("User doesn't completed registration ")
	}

	// 2. Begin registration (generates challenge)
	options, sessionData, err := ps.wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}

	// 3. Store session temporarily for FinishRegistration
	if err := ps.redis.StoreRegistrationSessionRedis(user.Id, sessionData); err != nil {
		return nil, err
	}

	return options, nil
}

// RegisterFinish finishes passkey registration it gets stored users passkey session from redis and validates registration if session is valid finishes it and stores
// credentials in user_passkey table
func (ps *PasskeyService) RegisterFinish(userID uint, r *http.Request) error {
	user, err := ps.userRepo.GetByID(ps.db, userID)
	if err != nil {
		return err
	}

	sessionData, err := ps.redis.GetRegistrationSessionRedis(userID)
	if err != nil {
		return err
	}

	cred, err := ps.wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		return err
	}
	authBytes, err := json.Marshal(cred.Authenticator)
	if err != nil {
		return err
	}

	if err := ps.userRepo.SavePasskey(ps.db, authBytes, user.Id, cred); err != nil {
		return err
	}

	if err := ps.redis.DeleteRegistrationSessionRedis(userID); err != nil {
		return err
	}
	return nil
}

func (ps *PasskeyService) LoginStart() (*protocol.CredentialAssertion, string, error) {
	// Generate a temporary session ID
	sessionID, _ := uuid.GenerateUUID() // implement a UUID generator
	assertion, sessionData, err := ps.wa.BeginDiscoverableLogin()

	if err != nil {
		return nil, "", err
	}
	// Store session in Redis for later finish
	if err := ps.redis.StoreSessionRedis(sessionID, sessionData); err != nil {
		return nil, "", err
	}

	return assertion, sessionID, nil
}

// LoginFinish validates the assertion response and updates signCount for the credential used
// Fixed LoginFinish method
func (ps *PasskeyService) LoginFinish(sessionID string, r *http.Request) (*response.Tokens, error) {
	// Retrieve session data from Redis
	sessionData, err := ps.redis.GetSessionRedis(sessionID)
	if err != nil {
		return nil, errors.New("failed to get session data")
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}

	// Reset r.Body so it can be read again
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Parse the assertion response from the request
	resp, err := protocol.ParseCredentialRequestResponse(r)
	if err != nil {
		return nil, errors.New("failed to parse credential response")
	}

	// Extract the credential ID from the response
	credentialID := resp.RawID
	if len(credentialID) == 0 {
		return nil, errors.New("missing credential ID in response")
	}

	// Find user by credential ID BEFORE calling FinishDiscoverableLogin
	user, err := ps.userRepo.FindUserByCredentialID(ps.db, credentialID)
	if err != nil {

		return nil, errors.New("user has no passkeys register one first")
	}

	sessionData.UserID = user.WebAuthnID()
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	credential, err := ps.wa.FinishLogin(user, *sessionData, r)
	if err != nil {
		return nil, errors.New("failed to finish login")
	}

	// Update the credential's sign count
	authBytes, _ := json.Marshal(credential.Authenticator)
	if err := ps.userRepo.UpdatePasskeyAfterLogin(ps.db, credential.ID, authBytes, credential.Authenticator.SignCount); err != nil {
		log.Printf("Warning: failed to update passkey after login: %v", err)
	}

	// Clean up: delete the temporary session
	if err := ps.redis.DeleteSessionRedis(sessionID); err != nil {
		log.Printf("Warning: failed to delete session: %v", err)
	}

	tokens, err := ps.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return &response.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
