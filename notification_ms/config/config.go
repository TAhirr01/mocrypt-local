package config

var Conf Config

type Config struct {
	Application Application `yaml:"application" json:"application"`
}

type Application struct {
	DisplayName string `yaml:"display-name" json:"display_name"`
	Server      Server `yaml:"server" json:"server"`
	Smtp        Smtp   `yaml:"smtp" json:"smtp"`
}

type Server struct {
	Port string `yaml:"port"`
}

type Smtp struct {
	From     string `yaml:"from"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}
