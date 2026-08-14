package config

import "os"

type Config struct {
	Host      string
	Port      string
	User      string
	Pass      string
	DBName    string
	AppPort   string
	JWTSecret []byte
}

func Load() *Config {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		Host:      os.Getenv("DB_HOST"),
		Port:      os.Getenv("DB_PORT"),
		User:      os.Getenv("DB_USER"),
		Pass:      os.Getenv("DB_PASSWORD"),
		DBName:    os.Getenv("DB_NAME"),
		AppPort:   appPort,
		JWTSecret: []byte(os.Getenv("JWT_SECRET")),
	}
}
