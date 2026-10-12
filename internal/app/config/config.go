package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type RedisConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Config struct {
	JWT   JWTConfig
	Redis RedisConfig
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func New() (*Config, error) {
	expiresIn, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "1h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRES_IN: %w", err)
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_PORT must be int: %w", err)
	}

	return &Config{
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "tco-secret"),
			ExpiresIn: expiresIn,
		},
		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        redisPort,
			User:        os.Getenv("REDIS_USER"),
			Password:    getEnv("REDIS_PASSWORD", "password"),
			DialTimeout: 10 * time.Second,
			ReadTimeout: 10 * time.Second,
		},
	}, nil
}
