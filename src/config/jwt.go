package config

import (
	"time"
)

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

func GetJWTConfig() JWTConfig {
	secret := "jwt-secret"

	return JWTConfig{
		Secret: secret,
		ExpiresIn: time.Hour,
	}
}