package config

import "github.com/go-webauthn/webauthn/webauthn"

func InitWebAuthn() *webauthn.WebAuthn {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: Conf.Application.WebAuthn.RpDisplayName,
		RPID:          Conf.Application.WebAuthn.RpID,
		RPOrigins:     []string{Conf.Application.WebAuthn.RpOrigin, "http://127.0.0.1:5500"},
	})

	if err != nil {
		panic(err)
	}
	return wa
}
