package controller

//
//
//import (
//	"bytes"
//	"encoding/json"
//	"net/http"
//	"time"
//	"user-service/service"
//	"user_management_ms/services"
//
//	"github.com/go-webauthn/webauthn/webauthn"
//	"github.com/gofiber/fiber/v2"
//)
//
//type IWebAuthnController interface {
//	WebAuthnBeginRegistration(c *fiber.Ctx) error
//}
//type WebAuthnController struct {
//	userService services.IUserService
//}
//
//func NewWebAuthnController(userService services.IUserService) IWebAuthnController {
//	return &WebAuthnController{userService: userService}
//}
//func (ac *WebAuthnController) WebAuthnBeginRegistration(c *fiber.Ctx) error {
//	email := gjson.GetBytes(c.Body(), "email").String()
//	user, err := ac.repo.FindByEmailWithCreds(email) // preload Credentials
//	if err != nil {
//		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
//	}
//
//	// ensure user has a stable handle
//	if len(user.WebAuthnUserHandle) == 0 {
//		user.WebAuthnUserHandle = mustRand32()
//		ac.repo.UpdateUserHandle(user.ID, user.WebAuthnUserHandle)
//	}
//
//	opts := []webauthn.RegistrationOption{
//		webauthn.WithAuthenticatorSelection(webauthn.AuthenticatorSelection{
//			AuthenticatorAttachment: webauthn.Platform,
//			ResidentKey:             webauthn.PreferResidentKey,
//			UserVerification:        webauthn.VerificationRequired,
//		}),
//		webauthn.WithConveyancePreference(webauthn.PreferNoAttestation),
//	}
//
//	sessionData, creationOpts, err := webAuthn.BeginRegistration(user, opts...)
//	if err != nil {
//		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
//	}
//
//	_ = ac.redis.Set(c.Context(), "webauthn:reg:"+email, mustJSON(map[string]any{
//		"session": sessionData, "user_id": user.ID,
//	}), 10*time.Minute).Err()
//
//	return c.JSON(creationOpts) // send to navigator.credentials.create
//}
//
//func (ac *AuthController) WebAuthnFinishRegistration(c *fiber.Ctx) error {
//	email := gjson.GetBytes(c.Body(), "email").String()
//
//	// load session
//	raw, err := ac.redis.Get(c.Context(), "webauthn:reg:"+email).Bytes()
//	if err != nil {
//		return c.Status(400).JSON(fiber.Map{"error": "session expired"})
//	}
//	var tmp struct {
//		Session webauthn.SessionData `json:"session"`
//		UserID  uint                 `json:"user_id"`
//	}
//	_ = json.Unmarshal(raw, &tmp)
//
//	user, err := ac.repo.FindByIDWithCreds(tmp.UserID)
//	if err != nil {
//		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
//	}
//
//	// Build a minimal *http.Request for the lib
//	httpReq, _ := http.NewRequest("POST", "", bytes.NewReader(c.Body()))
//	for k, v := range c.GetReqHeaders() {
//		httpReq.Header.Set(k, v)
//	}
//
//	cred, err := webAuthn.FinishRegistration(user, tmp.Session, httpReq)
//	if err != nil {
//		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
//	}
//
//	// persist
//	crec := WebAuthnCredential{
//		UserID:       user.ID,
//		CredentialID: cred.ID,
//		PublicKey:    cred.PublicKey,
//		SignCount:    cred.Authenticator.SignCount,
//		Transports:   cred.Transport,
//		AAGUID:       cred.Authenticator.AAGUID,
//	}
//	if err := ac.repo.CreateCredential(&crec); err != nil {
//		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
//	}
//
//	ac.redis.Del(c.Context(), "webauthn:reg:"+email)
//
//	// mark full auth completion if this is part of first-time flow
//	now := time.Now()
//	ac.repo.UpdateLastFullAuth(user.ID, &now)
//
//	return c.JSON(fiber.Map{"status": "ok"})
//}
