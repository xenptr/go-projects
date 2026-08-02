package config

import "os"

type Config struct {
	Host    string
	Port    string
	User    string
	Pass    string
	DBName  string
	AppPort string
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
