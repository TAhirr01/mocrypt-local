package services

import (
	"encoding/json"

	"log"
	"net/http"

	"user_management_ms/dtos/request"
	"user_management_ms/repository"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type IPasskeyService interface {
	RegisterStart(req *request.StartPasskeyRegistrationRequest) (*protocol.CredentialCreation, error)
	RegisterFinish(userID uint, credResp *protocol.CredentialCreationResponse, r *http.Request) error
	LoginStart(userID uint) (*protocol.PublicKeyCredentialRequestOptions, error)
	//LoginFinish(userID uint, assertionResp *protocol.CredentialAssertionResponse, r *http.Request) error
}

type PasskeyService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
	wa       *webauthn.WebAuthn
	redis    IRedisService
}

func NewPasskeyService(wa *webauthn.WebAuthn, db *gorm.DB, userRepo repository.IUserRepository, redis IRedisService) IPasskeyService {
	return &PasskeyService{wa: wa, db: db, userRepo: userRepo, redis: redis}
}

// RegisterStart start passkey registration stores temporary session inside redis
func (ps *PasskeyService) RegisterStart(req *request.StartPasskeyRegistrationRequest) (*protocol.CredentialCreation, error) {
	// 1. Fetch user + existing passkeys from DB
	user, err := ps.userRepo.GetUserWithPasskeys(ps.db, req.UserId)
	if err != nil {
		return nil, err
	}

	// 2. Begin registration (generates challenge)
	options, sessionData, err := ps.wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}
	log.Println("sessionData:", sessionData)
	log.Println("options:", options)

	// 3. Store session temporarily for FinishRegistration
	if err := ps.redis.StoreSessionRedis(user.Id, sessionData); err != nil {
		return nil, err
	}

	return options, nil
}

// RegisterFinish finishes passkey registration it gets stored users passkey session from redis and validates registration if session is valid finishes it and stores
// credentials in user_passkey table
func (ps *PasskeyService) RegisterFinish(userID uint, credResp *protocol.CredentialCreationResponse, r *http.Request) error {
	user, err := ps.userRepo.GetUserWithPasskeys(ps.db, userID)
	if err != nil {
		return err
	}

	sessionData, err := ps.redis.GetSessionRedis(userID)
	if err != nil {
		return err
	}

	cred, err := ps.wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		return err
	}

	//We can restrict user to have exactly one passkey per device but this approach has its own drawbacks
	//if user deletes passkey for example from Google password manager server will never be noticed about this action and user will be locked out of registering again

	//for _, pk := range user.Passkeys {
	//	if bytes.Equal(pk.AAGUID, cred.Authenticator.AAGUID) {
	//		// Delete session since we've used it
	//		if delErr := ps.redis.DeleteSessionRedis(userID); delErr != nil {
	//			log.Printf("failed to delete session: %v", delErr)
	//		}
	//		return fmt.Errorf("this device already has a passkey")
	//	}
	//}
	//log.Println("cred:", cred)
	authBytes, err := json.Marshal(cred.Authenticator)
	if err != nil {
		return err
	}

	if err := ps.userRepo.SavePasskey(ps.db, authBytes, user.Id, cred); err != nil {
		return err
	}

	if err := ps.redis.DeleteSessionRedis(userID); err != nil {
		return err
	}
	return nil
}

func (ps *PasskeyService) LoginStart(userID uint) (*protocol.PublicKeyCredentialRequestOptions, error) {
	user, err := ps.userRepo.GetUserWithPasskeys(ps.db, userID)
	if err != nil {
		return nil, err
	}

	options, sessionData, err := ps.wa.BeginLogin(user)
	if err != nil {
		return nil, err
	}

	if err := ps.redis.StoreSessionRedis(userID, sessionData); err != nil {
		return nil, err
	}
	return &options.Response, nil
}

//func (ps *PasskeyService) LoginFinish(userID uint, assertionResp *protocol.CredentialAssertionResponse, r *http.Request) error {
//	user, err := ps.userRepo.GetUserWithPasskeys(ps.db, userID)
//	if err != nil {
//		return fmt.Errorf("failed to get user with passkeys: %w", err)
//	}
//
//	sessionData, err := ps.redis.GetSessionRedis(userID)
//	if err != nil {
//		return fmt.Errorf("failed to get session data: %w", err)
//	}
//
//	// Convert stored authenticator data to webauthn.Authenticator objects
//	credentials := make([]webauthn.Credential, len(user.Passkeys))
//	for i, pk := range user.Passkeys {
//		var auth webauthn.Authenticator
//		if err := json.Unmarshal(pk.Authenticator, &auth); err != nil {
//			return fmt.Errorf("failed to unmarshal authenticator: %w", err)
//		}
//
//		credentials[i] = webauthn.Credential{
//			ID:              pk.CredentialID,
//			PublicKey:       pk.PublicKey,
//			AttestationType: pk.AttestationType,
//			Authenticator:   auth,
//		}
//	}
//
//	// Create a temporary user with the proper credentials for webauthn
//	webAuthnUser := &struct {
//		*domain.User
//		credentials []webauthn.Credential
//	}{
//		User:        user,
//		credentials: credentials,
//	}
//
//	// Override the WebAuthnCredentials method to use our parsed credentials
//	webAuthnUser.WebAuthnCredentials = func() []webauthn.Credential {
//		return webAuthnUser.credentials
//	}
//
//	cred, err := ps.wa.FinishLogin(webAuthnUser, *sessionData, r)
//	if err != nil {
//		return fmt.Errorf("failed to finish login: %w", err)
//	}
//
//	// Update signCount for the specific credential used
//	if err := ps.userRepo.UpdateSignCountByCredentialID(ps.db, cred.ID, cred.Authenticator.SignCount); err != nil {
//		return fmt.Errorf("failed to update sign count: %w", err)
//	}
//
//	if err := ps.redis.DeleteSessionRedis(userID); err != nil {
//		log.Printf("failed to delete session: %v", err)
//		// We don't return this error as the login was successful
//	}
//
//	return nil
//}
