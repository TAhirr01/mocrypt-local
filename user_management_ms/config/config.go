package config

var Conf Config

type Config struct {
	Application Application `yaml:"application" json:"application"`
}

type Application struct {
	DisplayName string     `yaml:"display-name" json:"display_name"`
	Server      Server     `yaml:"server" json:"server"`
	Datasource  Datasource `yaml:"datasource" json:"datasource"`
	Migration   string     `yaml:"migration"`
	Security    Security   `yaml:"security" json:"security"`
	OPABaseURI  string     `yaml:"opa-base-uri" json:"opa_base_uri"`
	Redis       Redis      `yaml:"redis" json:"redis"`
	OAuth2      OAuth2     `yaml:"oauth2" json:"oauth2"`
	WebAuthn    WebAuthn   `yaml:"webauthn" json:"webauthn"`
}

type Server struct {
	ContextPath string `yaml:"context-path" json:"context_path"`
	ApiVersion  string `yaml:"api-version" json:"api_version"`
	Port        string `yaml:"port"`
}

type Datasource struct {
	PrimaryURL            string `yaml:"primary-url" json:"primary_url"`
	SecondaryURL          string `yaml:"secondary-url" json:" secondary_url"`
	MaxIdleConnections    int    `yaml:"max-idle-connections" json:"max_idle_connections"`
	MaxOpenConnections    int    `yaml:"max-open-connections" json:"max_open_connections"`
	ConnectionMaxLifetime int    `yaml:"connection-max-lifetime" json:"connection_max_lifetime"`
}
type Security struct {
	Secret                              string `yaml:"secret" json:"-"`
	Issuer                              string `yaml:"issuer" json:"issuer"`
	TokenValidityInSeconds              int    `yaml:"token-validity-in-seconds" json:"token_validity_in_seconds"`
	TokenValidityInSecondsForRememberMe int    `yaml:"token-validity-in-seconds-for-remember-me" json:"token_validity_in_seconds_for_remember_me"`
}

type Redis struct {
	Host string `yaml:"address" json:"address"`
}

type OAuth2 struct {
	RedirectUri  string `yaml:"redirect_uri" json:"redirect_uri"`
	ClientID     string `yaml:"client-id" json:"client_id"`
	ClientSecret string `yaml:"client-secret" json:"client_secret"`
	Scope        string `yaml:"scope" json:"scope"`
}

type WebAuthn struct {
	RpDisplayName string `yaml:"rp-display-name" json:"rp_display_name"`
	RpOrigin      string `yaml:"rp-origin" json:"rp_origin"`
	RpID          string `yaml:"rp-id" json:"rp_id"`
}
