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
type config struct {
	Port               string
	DatabasePath       string
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpire time.Duration
	OTPExpiry          time.Duration
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPass           string
	SMTPFrom           string
	RateLimitRPS       int
	RateLimitBurst     int
}

// load config
func load() *config {

	//load env
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("cmd/.env"); err != nil {
			log.Println("no .env file found, reading from environment", err)
		}
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	rps, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "10"))
	burst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	secret := "GAFCGPWGECBW6A5SC646+AS4WIFUJCSABJK874JBMNZCBNMmasvkgkhkvsxlksJHkKVKgjJGkNL"

	return &config{

		Port:               getEnv("PORT", "8080"),
		DatabasePath:       getEnv("DATABASE_PATH", "./auth.db"),
		JWTSecret:          getEnv("JWT_SECRET", secret),
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpire: 7 * 24 * time.Hour,
		OTPExpiry:          10 * time.Minute,
		SMTPHost:           getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:           smtpPort,
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPass:           getEnv("SMTP_PASS", ""),
		SMTPFrom:           getEnv("SMTP_FROM", "noreply@nexplay.com"),
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
	}

}
