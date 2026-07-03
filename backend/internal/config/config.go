package config

import (
	"bufio"
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
	loadDotEnv(".env")
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

// loadDotEnv reads KEY=VALUE lines from path and populates any unset vars.
// Missing file is not an error — .env is optional.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:eq])
		v := strings.TrimSpace(line[eq+1:])
		if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
			v = v[1 : len(v)-1]
		}
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
