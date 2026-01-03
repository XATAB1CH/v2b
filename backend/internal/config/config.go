package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	OpenAIKey   string
	OpenAIModel string
	OpenAIBase  string

	HTTPTimeout time.Duration

	CORSAllowOrigins []string

	// Postgres
	DatabaseURL string

	// JWT
	JWTAccessSecret  string
	JWTAccessTTLMins int

	// OTP
	OTPTTLMins     int
	OTPCodeLength  int
	OTPMaxAttempts int

	FreeAttemptsLimit        int
	SubscriptionDurationDays int
}

func Load() Config {
	loadDotEnv()

	return Config{
		Port: getenv("PORT", "8080"),
		Env:  getenv("ENV", "local"),

		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIModel: getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBase:  getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		HTTPTimeout: time.Duration(getenvInt("HTTP_TIMEOUT_SECONDS", 45)) * time.Second,

		CORSAllowOrigins: parseCSV(getenv("CORS_ALLOW_ORIGINS", "*")),

		DatabaseURL: os.Getenv("DATABASE_URL"),

		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTAccessTTLMins: getenvInt("JWT_ACCESS_TTL_MINUTES", 15),

		OTPTTLMins:     getenvInt("OTP_TTL_MINUTES", 10),
		OTPCodeLength:  getenvInt("OTP_CODE_LENGTH", 6),
		OTPMaxAttempts: getenvInt("OTP_MAX_ATTEMPTS", 5),

		FreeAttemptsLimit:        getenvInt("FREE_ATTEMPTS_LIMIT", 10),
		SubscriptionDurationDays: getenvInt("SUBSCRIPTION_DURATION_DAYS", 30),
	}
}

func loadDotEnv() {
	// Пробуем несколько путей — удобно при запуске из разных директорий
	paths := []string{
		".env",
		"../.env",
		"../../.env",
		"../../../.env",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err != nil {
				log.Printf("failed to load %s: %v", p, err)
			}
			return
		}
	}

	log.Print("No .env file found (searched: .env, ../.env, ../../.env, ../../../.env)")
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
