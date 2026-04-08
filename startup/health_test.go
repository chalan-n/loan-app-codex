package startup

import (
	"loan-app/config"
	"strings"
	"testing"
)

func TestValidateConfigAcceptsHealthyProductionConfig(t *testing.T) {
	cfg := &config.AppConfig{
		DBHost:                          "127.0.0.1",
		DBPort:                          "3306",
		DBUser:                          "root",
		DBName:                          "loan_db",
		AppPort:                         "3000",
		AppEnv:                          "production",
		JWTSecret:                       "super-secret",
		SessionIdleTimeoutMinutes:       30,
		SessionActivityRefreshSeconds:   300,
		LoginRateLimitMax:               5,
		LoginRateLimitWindowSeconds:     300,
		SearchRateLimitMax:              60,
		SearchRateLimitWindowSeconds:    60,
		InsuranceRateLimitMax:           30,
		InsuranceRateLimitWindowSeconds: 60,
		UploadRateLimitMax:              20,
		UploadRateLimitWindowSeconds:    300,
		MobileAPIKey:                    "mobile-key",
		R2AccountId:                     "account",
		R2AccessKeyId:                   "access",
		R2SecretAccessKey:               "secret",
		R2BucketName:                    "bucket",
		R2Endpoint:                      "https://example.r2.cloudflarestorage.com",
		WebAuthnRPID:                    "loan.example.com",
		WebAuthnOrigin:                  "https://loan.example.com",
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected config to be valid, got %v", err)
	}
}

func TestValidateConfigRejectsProductionDefaults(t *testing.T) {
	cfg := &config.AppConfig{
		DBHost:                          "127.0.0.1",
		DBPort:                          "3306",
		DBUser:                          "root",
		DBName:                          "loan_db",
		AppPort:                         "3000",
		AppEnv:                          "production",
		JWTSecret:                       "changeme-secret",
		SessionIdleTimeoutMinutes:       30,
		SessionActivityRefreshSeconds:   300,
		LoginRateLimitMax:               5,
		LoginRateLimitWindowSeconds:     300,
		SearchRateLimitMax:              60,
		SearchRateLimitWindowSeconds:    60,
		InsuranceRateLimitMax:           30,
		InsuranceRateLimitWindowSeconds: 60,
		UploadRateLimitMax:              20,
		UploadRateLimitWindowSeconds:    300,
		WebAuthnRPID:                    "loan.example.com",
		WebAuthnOrigin:                  "http://loan.example.com",
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected config validation error")
	}

	got := err.Error()
	for _, expected := range []string{
		"JWT_SECRET must not use the default production value",
		"MOBILE_API_KEY is required in production",
		"WEBAUTHN_ORIGIN must use https in production",
		"R2 configuration is required in production",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected error to contain %q, got %q", expected, got)
		}
	}
}

func TestValidateConfigRejectsIncompleteR2AndInvalidLimits(t *testing.T) {
	cfg := &config.AppConfig{
		DBHost:                          "127.0.0.1",
		DBPort:                          "3306",
		DBUser:                          "root",
		DBName:                          "loan_db",
		AppPort:                         "3000",
		AppEnv:                          "development",
		JWTSecret:                       "secret",
		SessionIdleTimeoutMinutes:       0,
		SessionActivityRefreshSeconds:   300,
		LoginRateLimitMax:               0,
		LoginRateLimitWindowSeconds:     300,
		SearchRateLimitMax:              60,
		SearchRateLimitWindowSeconds:    60,
		InsuranceRateLimitMax:           30,
		InsuranceRateLimitWindowSeconds: 60,
		UploadRateLimitMax:              20,
		UploadRateLimitWindowSeconds:    300,
		R2AccountId:                     "account",
		R2AccessKeyId:                   "access",
		WebAuthnRPID:                    "localhost",
		WebAuthnOrigin:                  "https://localhost:3000",
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected config validation error")
	}

	got := err.Error()
	for _, expected := range []string{
		"SESSION_IDLE_TIMEOUT_MINUTES must be greater than zero",
		"LOGIN_RATE_LIMIT_MAX must be greater than zero",
		"R2 configuration is incomplete",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected error to contain %q, got %q", expected, got)
		}
	}
}

func TestVerifyDatabaseRejectsNilDB(t *testing.T) {
	err := VerifyDatabase(nil)
	if err == nil {
		t.Fatal("expected nil database to fail verification")
	}
	if !strings.Contains(err.Error(), "database is nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}
