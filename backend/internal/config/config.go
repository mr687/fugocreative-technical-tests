package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	AdminUsername     string
	AdminPassword     string
	JWTSecret         string
	JWTTTL            time.Duration
	LLMProvider       string
	LLMModel          string
	LLMAPIKey         string
	LLMEndpoint       string
	LLMTimeout        time.Duration
	ServerAddr        string
	CORSAllowedOrigin string
}

func Load() (Config, error) {
	// Load local .env when present; deployment environments supply variables directly.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AdminUsername:     strings.TrimSpace(os.Getenv("ADMIN_USERNAME")),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		LLMProvider:       strings.TrimSpace(os.Getenv("LLM_PROVIDER")),
		LLMModel:          strings.TrimSpace(os.Getenv("LLM_MODEL")),
		LLMAPIKey:         os.Getenv("LLM_API_KEY"),
		LLMEndpoint:       strings.TrimSpace(os.Getenv("LLM_ENDPOINT")),
		ServerAddr:        strings.TrimSpace(os.Getenv("SERVER_ADDR")),
		CORSAllowedOrigin: strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGIN")),
	}
	if err := required("DATABASE_URL", cfg.DatabaseURL); err != nil {
		return Config{}, err
	}
	databaseURL, err := url.Parse(cfg.DatabaseURL)
	if err != nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") || databaseURL.Host == "" || databaseURL.Path == "" || databaseURL.User == nil {
		return Config{}, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL with credentials and database name")
	}
	if err := required("ADMIN_USERNAME", cfg.AdminUsername); err != nil {
		return Config{}, err
	}
	if err := required("ADMIN_PASSWORD", cfg.AdminPassword); err != nil {
		return Config{}, err
	}
	if err := required("JWT_SECRET", cfg.JWTSecret); err != nil {
		return Config{}, err
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	cfg.JWTTTL, err = parseDuration("JWT_TTL", os.Getenv("JWT_TTL"), time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg.LLMTimeout, err = parseDuration("LLM_TIMEOUT", os.Getenv("LLM_TIMEOUT"), 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}
	_, port, err := net.SplitHostPort(cfg.ServerAddr)
	if err != nil {
		return Config{}, fmt.Errorf("SERVER_ADDR must be host:port: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("SERVER_ADDR must contain a valid port")
	}
	if err := validateOrigin(cfg.CORSAllowedOrigin); err != nil {
		return Config{}, err
	}
	if err := validateLLM(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func required(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func parseDuration(name, value string, fallback time.Duration) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	if strings.HasSuffix(value, "d") {
		days, err := strconv.ParseInt(strings.TrimSuffix(value, "d"), 10, 64)
		if err != nil || days <= 0 || days > int64((1<<63-1)/int64(24*time.Hour)) {
			return 0, fmt.Errorf("%s must be a positive duration", name)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return duration, nil
}

func validateOrigin(origin string) error {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("CORS_ALLOWED_ORIGIN must be an http(s) origin without path, query, or fragment")
	}
	return nil
}

func validateLLM(cfg Config) error {
	values := []struct {
		name  string
		value string
	}{
		{"LLM_PROVIDER", cfg.LLMProvider},
		{"LLM_MODEL", cfg.LLMModel},
		{"LLM_API_KEY", cfg.LLMAPIKey},
		{"LLM_ENDPOINT", cfg.LLMEndpoint},
	}
	configured := 0
	for _, value := range values {
		if value.value != "" {
			configured++
		}
	}
	if configured == 0 {
		return nil
	}
	for _, value := range values {
		if err := required(value.name, value.value); err != nil {
			return err
		}
	}
	endpoint, err := url.Parse(cfg.LLMEndpoint)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return fmt.Errorf("LLM_ENDPOINT must be an http(s) URL")
	}
	return nil
}
