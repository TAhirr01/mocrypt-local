package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"user_management_ms/domain"
	"user_management_ms/dtos/request"
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
	LoginFinish(sessionID string, r *http.Request) (*domain.User, error)
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

	log.Println("SessionData:", sessionData)
	if err != nil {
		return nil, "", err
	}
	sessionData.AllowedCredentialIDs = [][]byte{}
	// Store session in Redis for later finish
	if err := ps.redis.StoreSessionRedis(sessionID, sessionData); err != nil {
		return nil, "", err
	}

	return assertion, sessionID, nil
}

// LoginFinish validates the assertion response and updates signCount for the credential used
// Fixed LoginFinish method
func (ps *PasskeyService) LoginFinish(sessionID string, r *http.Request) (*domain.User, error) {
	log.Printf("=== LoginFinish Debug Start ===")

	// Retrieve session data from Redis
	sessionData, err := ps.redis.GetSessionRedis(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	log.Printf("Retrieved session data: %+v", sessionData)
	log.Printf("Session challenge: %x", sessionData.Challenge)
	log.Printf("Session UserID: %s", string(sessionData.UserID))

	// Parse the assertion response from the request
	response, err := protocol.ParseCredentialRequestResponse(r)
	if err != nil {
		log.Printf("ParseCredentialRequestResponse error: %v", err)
		return nil, fmt.Errorf("failed to parse credential response: %w", err)
	}

	// Extract the credential ID from the response
	credentialID := response.RawID
	if len(credentialID) == 0 {
		return nil, fmt.Errorf("missing credential ID in response")
	}

	// Find user by credential ID BEFORE calling FinishDiscoverableLogin
	user, err := ps.userRepo.FindUserByCredentialID(ps.db, credentialID)
	if err != nil {
		log.Printf("FindUserByCredentialID error: %v", err)
		return nil, fmt.Errorf("failed to find user by credential ID: %w", err)
	}

	log.Printf("About to call FinishDiscoverableLogin...")
	log.Println("User:", user)
	credential, err := ps.wa.FinishLogin(user, *sessionData, r)

	//The actual call that's failing
	//credential, err := ps.wa.FinishDiscoverableLogin(
	//	func(rawID, userHandle []byte) (webauthn.User, error) {
	//		log.Printf("=== CALLBACK CALLED ===")
	//		log.Printf("Callback rawID: %v (hex: %x)", rawID, rawID)
	//		log.Printf("Callback userHandle: %v (string: %s)", userHandle, string(userHandle))
	//		log.Printf("Expected credentialID: %v (hex: %x)", credentialID, credentialID)
	//
	//		// Check if rawID matches any of the user's credentials
	//		var matchFound bool
	//		for _, passkey := range user.Passkeys {
	//			if bytes.Equal(rawID, passkey.CredentialID) {
	//				log.Printf("Found matching credential in user's passkeys")
	//				matchFound = true
	//				break
	//			}
	//		}
	//
	//		if !matchFound {
	//			log.Printf("No matching credential found in user's passkeys")
	//			// Don't return error yet, let's see what WebAuthn expects
	//		}
	//
	//		// Verify this is the same credential ID from the response
	//		if !bytes.Equal(rawID, credentialID) {
	//			log.Printf("WARNING: rawID from callback doesn't match credentialID from response")
	//			log.Printf("This might be normal - WebAuthn might be checking different credentials")
	//			// Don't return error here - let WebAuthn continue its process
	//		}
	//
	//		log.Printf("Returning user from callback: %+v", user)
	//		return user, nil
	//	},
	//	*sessionData,
	//	r,
	//)

	if err != nil {
		log.Printf("FinishDiscoverableLogin failed with error: %v", err)
		log.Printf("Error type: %T", err)
		return nil, fmt.Errorf("failed to finish discoverable login: %w", err)
	}

	log.Printf("FinishDiscoverableLogin successful!")
	log.Printf("Returned credential: ID=%v, SignCount=%d", credential.ID, credential.Authenticator.SignCount)

	// Update the credential's sign count
	authBytes, _ := json.Marshal(credential.Authenticator)
	if err := ps.userRepo.UpdatePasskeyAfterLogin(ps.db, credential.ID, authBytes, credential.Authenticator.SignCount); err != nil {
		log.Printf("Warning: failed to update passkey after login: %v", err)
	}

	// Clean up: delete the temporary session
	if err := ps.redis.DeleteSessionRedis(sessionID); err != nil {
		log.Printf("Warning: failed to delete session: %v", err)
	}

	log.Printf("=== LoginFinish Debug End - SUCCESS ===")
	return user, nil
}
