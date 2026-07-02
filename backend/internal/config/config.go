package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          string
	DataDir       string
	AnthropicKey  string
	Model         string
	AllowedOrigin string
	JWTSecret     []byte
	InviteCodes   map[string]bool // invite-only signup
	DailyLLMCap   int
}

func Load() Config {
	secret := env("JWT_SECRET", "dev-insecure-secret-change-me")
	codes := map[string]bool{}
	for _, c := range strings.Split(env("INVITE_CODES", "friends2026"), ",") {
		if c = strings.TrimSpace(c); c != "" {
			codes[c] = true
		}
	}
	cap, _ := strconv.Atoi(env("DAILY_LLM_CAP", "25"))
	return Config{
		Port:          env("PORT", "8080"),
		DataDir:       env("DATA_DIR", "./data"),
		AnthropicKey:  os.Getenv("ANTHROPIC_API_KEY"),
		Model:         env("JOBSCOUT_MODEL", "claude-haiku-4-5-20251001"),
		AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:3000"),
		JWTSecret:     []byte(secret),
		InviteCodes:   codes,
		DailyLLMCap:   cap,
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
