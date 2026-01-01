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
