package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath     string
	ServerPort string
	JWT        JWTConfig
}

type JWTConfig struct {
	Secret string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBPath:     getEnv("DB_PATH", "library.db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-change-me"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
