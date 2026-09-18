// Package config loads LivePoll's configuration from environment variables
// (see .env.example).
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string // "development" | "production"
	MongoURI    string
	MongoDB     string
	RedisURL    string
	JWTSecret   string
	JWTTTL      time.Duration
	FrontendURL string // used for CORS + building shareable poll links
}

// Load reads configuration from the environment.
func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getenv("PORT", "8080"),
		Env:         getenv("APP_ENV", "development"),
		MongoURI:    getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:     getenv("MONGO_DATABASE", "livepoll"),
		RedisURL:    getenv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		FrontendURL: getenv("FRONTEND_URL", "http://localhost:5173"),
	}

	ttlMinutes, err := strconv.Atoi(getenv("JWT_TTL_MINUTES", "10080")) // 7 days
	if err != nil {
		log.Fatalf("invalid JWT_TTL_MINUTES: %v", err)
	}
	cfg.JWTTTL = time.Duration(ttlMinutes) * time.Minute

	if cfg.JWTSecret == "" && cfg.Env != "production" {
		// Development-only fallback so the server is easy to boot locally.
		cfg.JWTSecret = "dev-only-insecure-secret-change-me"
	}

	return cfg
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
