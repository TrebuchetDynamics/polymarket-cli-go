package config

import (
	"os"
)

type Config struct {
	PrivateKey string
	CLOBHost   string
	GammaURL   string
	ChainID    int
}

func Load() Config {
	cfg := Config{
		PrivateKey: os.Getenv("POLYMARKET_PRIVATE_KEY"),
		CLOBHost:   getEnv("POLYMARKET_HOST", "https://clob.polymarket.com"),
		GammaURL:   getEnv("POLYMARKET_GAMMA_URL", "https://gamma-api.polymarket.com"),
		ChainID:    137,
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
