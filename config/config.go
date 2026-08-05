package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Get env. val
func getEnv(key, fallback string) string {

	if val := os.Getenv(key); val != "" {
		return val

	}

	return fallback
}

// config struct
type Config struct {
	Port               string
	DatabasePath       string
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpire time.Duration
	OTPExpiry          time.Duration
	GmailFrom          string
	GmailClientID      string
	GmailClientSecret  string
	GmailRefreshToken  string
	RateLimitRPS       int
	RateLimitBurst     int
}

// load config
func Load() *Config {

	//load env
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("cmd/.env"); err != nil {
			log.Println("no .env file found, reading from environment", err)
		}
	}

	rps, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "10"))
	burst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	secret := "GAFCGPWGECBW6A5SC646+AS4WIFUJCSABJK874JBMNZCBNMmasvkgkhkvsxlksJHkKVKgjJGkNL"

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabasePath:       getEnv("DATABASE_PATH", "./auth.db"),
		JWTSecret:          getEnv("JWT_SECRET", secret),
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpire: 7 * 24 * time.Hour,
		OTPExpiry:          10 * time.Minute,
		GmailFrom:          getEnv("GMAIL_FROM", ""),
		GmailClientID:      getEnv("GMAIL_CLIENT_ID", ""),
		GmailClientSecret:  getEnv("GMAIL_CLIENT_SECRET", ""),
		GmailRefreshToken:  getEnv("GMAIL_REFRESH_TOKEN", ""),
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
	}
}
