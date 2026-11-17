package config

import (
	"fmt"
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		check   func() error
		wantErr bool
	}{
		{"valid server origin", func() error { return validateOrigin("https://admin.example.com") }, false},
		{"origin with path", func() error { return validateOrigin("https://admin.example.com/dashboard") }, true},
		{"positive duration", func() error { _, err := parseDuration("JWT_TTL", "1h", time.Hour); return err }, false},
		{"day duration", func() error {
			got, err := parseDuration("JWT_TTL", "1d", time.Hour)
			if err == nil && got != 24*time.Hour {
				return fmt.Errorf("got %s", got)
			}
			return err
		}, false},
		{"zero duration", func() error { _, err := parseDuration("JWT_TTL", "0s", time.Hour); return err }, true},
		{"incomplete LLM config", func() error { return validateLLM(Config{LLMProvider: "provider"}) }, true},
		{"complete LLM config", func() error {
			return validateLLM(Config{LLMProvider: "provider", LLMModel: "model", LLMAPIKey: "secret", LLMEndpoint: "https://llm.example.com/v1"})
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.check()
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}
