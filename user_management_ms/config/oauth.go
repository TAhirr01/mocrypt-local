package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func InitOAuth() *oauth2.Config {
	OauthConf := &oauth2.Config{
		RedirectURL:  Conf.Application.OAuth2.RedirectUri,
		ClientID:     Conf.Application.OAuth2.ClientID,
		ClientSecret: Conf.Application.OAuth2.ClientSecret,
		Scopes:       []string{Conf.Application.OAuth2.Scope},
		Endpoint:     google.Endpoint,
	}

	return OauthConf
}
