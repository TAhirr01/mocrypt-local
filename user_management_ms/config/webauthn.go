package config

import "github.com/go-webauthn/webauthn/webauthn"

func InitWebAuthn() *webauthn.WebAuthn {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: Conf.Application.WebAuthn.RpDisplayName,
		RPID:          Conf.Application.WebAuthn.RpID,
		RPOrigins:     []string{Conf.Application.WebAuthn.RpOrigin},
	})
	if err != nil {
		panic(err)
	}
	return wa
}
