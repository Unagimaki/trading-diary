package config

import "os"

type Config struct {
	Environment     string
	HTTPAddress     string
	DatabaseURL     string
	FrontendOrigin  string
	UploadDirectory string
}

func FromEnv() Config {
	return Config{
		Environment:     envOrDefault("APP_ENV", "development"),
		HTTPAddress:     ":" + envOrDefault("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		FrontendOrigin:  envOrDefault("FRONTEND_ORIGIN", "http://localhost:5173"),
		UploadDirectory: envOrDefault("UPLOAD_DIRECTORY", "./data/uploads"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
