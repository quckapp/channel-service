package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Environment  string
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers []string
	JWTSecret    string
	ServiceName  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	return &Config{
		Port:         getEnv("PORT", "3003"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		DatabaseURL:  getEnv("DATABASE_URL", "root:password@tcp(localhost:3306)/quckapp_channels?parseTime=true"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		ServiceName:  "channel-service",
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
