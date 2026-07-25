package config

import "os"

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	AppPort  string
}

func Load() *Config {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		Host:     os.Getenv("DBHOST"),
		Port:     os.Getenv("DBPORT"),
		Username: os.Getenv("DBUSER"),
		Password: os.Getenv("DBPASS"),
		DBName:   os.Getenv("DBNAME"),
		AppPort:  appPort,
	}
}
