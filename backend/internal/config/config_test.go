package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsWeakJWTSecretInProduction(t *testing.T) {
	for name, secret := range map[string]string{
		"dev default":          "dev-jwt-secret-change-me",
		"Coolify placeholder":  "missing JWT_SECRET",
		"too short":            "abc123",
		"31 characters":        strings.Repeat("a", 31),
		"empty":                "",
		"placeholder, padded ": "missing JWT_SECRET " + strings.Repeat("x", 40),
	} {
		cfg := &Config{AppSettings: AppSettings{Environment: "production"}, JWTSecret: secret}
		if err := cfg.validate(); err == nil {
			t.Errorf("%s: want error for JWT_SECRET %q in production", name, secret)
		}
	}
}

func TestValidateAcceptsStrongJWTSecretInProduction(t *testing.T) {
	cfg := &Config{AppSettings: AppSettings{Environment: "production"}, JWTSecret: strings.Repeat("ab", 32)}
	if err := cfg.validate(); err != nil {
		t.Fatalf("strong secret rejected: %v", err)
	}
}

func TestValidateAllowsDevDefaultOutsideProduction(t *testing.T) {
	cfg := &Config{AppSettings: AppSettings{Environment: "development"}, JWTSecret: "dev-jwt-secret-change-me"}
	if err := cfg.validate(); err != nil {
		t.Fatalf("development must keep working with the dev default: %v", err)
	}
}

func TestValidateErrorNeverContainsTheSecret(t *testing.T) {
	cfg := &Config{AppSettings: AppSettings{Environment: "production"}, JWTSecret: "short-secret-value"}
	err := cfg.validate()
	if err == nil || strings.Contains(err.Error(), "short-secret-value") {
		t.Fatalf("error must exist and not echo the secret: %v", err)
	}
}
