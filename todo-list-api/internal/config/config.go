package config

import "os"

type Config struct {
	Host    string
	Port    string
	User    string
	Pass    string
	DBName  string
	AppPort string

	JWTSecret []byte
}

func Load() *Config {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		Host:    os.Getenv("DBHOST"),
		Port:    os.Getenv("DBPORT"),
		User:    os.Getenv("DBUSER"),
		Pass:    os.Getenv("DBPASS"),
		DBName:  os.Getenv("DBNAME"),
		AppPort: appPort,
	}
}

func SecretKey() *Config {
	return &Config{
		JWTSecret: []byte(os.Getenv("JWT_SECRET")),
	}
}
