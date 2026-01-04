package config

import (
	"fmt"
	"os"
)

type Config struct {
	SFClientID     string
	SFClientSecret string
	SFLoginURL     string
	Port           string
}

func Load() (Config, error) {
	cfg := Config{
		SFClientID:     os.Getenv("SF_CLIENT_ID"),
		SFClientSecret: os.Getenv("SF_CLIENT_SECRET"),
		SFLoginURL:     os.Getenv("SF_LOGIN_URL"),
		Port:           getEnvOrDefault("PORT", "8080"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.SFClientID == "" {
		return fmt.Errorf("SF_CLIENT_ID is required")
	}
	if c.SFClientSecret == "" {
		return fmt.Errorf("SF_CLIENT_SECRET is required")
	}
	if c.SFLoginURL == "" {
		return fmt.Errorf("SF_LOGIN_URL is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
