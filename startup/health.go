package startup

import (
	"errors"
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"strings"

	"gorm.io/gorm"
)

func Verify(cfg *config.AppConfig, db *gorm.DB) error {
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	if err := VerifyDatabase(db); err != nil {
		return err
	}
	return nil
}

func ValidateConfig(cfg *config.AppConfig) error {
	if cfg == nil {
		return errors.New("startup config is nil")
	}

	var issues []string

	requireString(&issues, cfg.DBHost, "DB_HOST")
	requireString(&issues, cfg.DBPort, "DB_PORT")
	requireString(&issues, cfg.DBUser, "DB_USER")
	requireString(&issues, cfg.DBName, "DB_NAME")
	requireString(&issues, cfg.AppPort, "APP_PORT")
	requireString(&issues, cfg.JWTSecret, "JWT_SECRET")
	requireString(&issues, cfg.WebAuthnRPID, "WEBAUTHN_RPID")
	requireString(&issues, cfg.WebAuthnOrigin, "WEBAUTHN_ORIGIN")

	requirePositive(&issues, cfg.SessionIdleTimeoutMinutes, "SESSION_IDLE_TIMEOUT_MINUTES")
	requirePositive(&issues, cfg.SessionActivityRefreshSeconds, "SESSION_ACTIVITY_REFRESH_SECONDS")
	requirePositive(&issues, cfg.LoginRateLimitMax, "LOGIN_RATE_LIMIT_MAX")
	requirePositive(&issues, cfg.LoginRateLimitWindowSeconds, "LOGIN_RATE_LIMIT_WINDOW_SECONDS")
	requirePositive(&issues, cfg.SearchRateLimitMax, "SEARCH_RATE_LIMIT_MAX")
	requirePositive(&issues, cfg.SearchRateLimitWindowSeconds, "SEARCH_RATE_LIMIT_WINDOW_SECONDS")
	requirePositive(&issues, cfg.InsuranceRateLimitMax, "INSURANCE_RATE_LIMIT_MAX")
	requirePositive(&issues, cfg.InsuranceRateLimitWindowSeconds, "INSURANCE_RATE_LIMIT_WINDOW_SECONDS")
	requirePositive(&issues, cfg.UploadRateLimitMax, "UPLOAD_RATE_LIMIT_MAX")
	requirePositive(&issues, cfg.UploadRateLimitWindowSeconds, "UPLOAD_RATE_LIMIT_WINDOW_SECONDS")

	validateR2Config(&issues, cfg)

	if cfg.IsProd() {
		if cfg.JWTSecret == "changeme-secret" {
			issues = append(issues, "JWT_SECRET must not use the default production value")
		}
		if cfg.MobileAPIKey == "" {
			issues = append(issues, "MOBILE_API_KEY is required in production")
		}
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(cfg.WebAuthnOrigin)), "https://") {
			issues = append(issues, "WEBAUTHN_ORIGIN must use https in production")
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("startup config validation failed: %s", strings.Join(issues, "; "))
	}

	return nil
}

func VerifyDatabase(db *gorm.DB) error {
	if db == nil {
		return errors.New("startup database verification failed: database is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("startup database verification failed: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("startup database verification failed: ping database: %w", err)
	}

	var issues []string

	for _, table := range requiredTables() {
		if !db.Migrator().HasTable(table.Model) {
			issues = append(issues, fmt.Sprintf("missing table %s", table.Name))
		}
	}

	for _, column := range requiredUserColumns() {
		if !db.Migrator().HasColumn(&models.User{}, column) {
			issues = append(issues, fmt.Sprintf("users.%s is missing", column))
		}
	}

	for _, column := range requiredLoanFileColumns() {
		if !db.Migrator().HasColumn(&models.LoanFile{}, column) {
			issues = append(issues, fmt.Sprintf("loan_application_files.%s is missing", column))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("startup database verification failed: %s", strings.Join(issues, "; "))
	}

	return nil
}

type requiredTable struct {
	Name  string
	Model interface{}
}

func requiredTables() []requiredTable {
	return []requiredTable{
		{Name: "schema_migrations", Model: &models.SchemaMigration{}},
		{Name: "users", Model: &models.User{}},
		{Name: "loan_applications", Model: &models.LoanApplication{}},
		{Name: "guarantors", Model: &models.Guarantor{}},
		{Name: "audit_logs", Model: &models.AuditLog{}},
		{Name: "webauthn_credentials", Model: &models.WebAuthnCredential{}},
		{Name: "ref_runnings", Model: &models.RefRunning{}},
		{Name: "loan_application_files", Model: &models.LoanFile{}},
	}
}

func requiredUserColumns() []string {
	return []string{
		"current_session_id",
		"session_last_activity_at",
		"session_revoked_at",
	}
}

func requiredLoanFileColumns() []string {
	return []string{
		"loan_id",
		"storage_key",
		"category",
		"uploaded_by",
	}
}

func requireString(issues *[]string, value, key string) {
	if strings.TrimSpace(value) == "" {
		*issues = append(*issues, key+" is required")
	}
}

func requirePositive(issues *[]string, value int, key string) {
	if value <= 0 {
		*issues = append(*issues, key+" must be greater than zero")
	}
}

func validateR2Config(issues *[]string, cfg *config.AppConfig) {
	fields := map[string]string{
		"R2_ACCOUNT_ID":        cfg.R2AccountId,
		"R2_ACCESS_KEY_ID":     cfg.R2AccessKeyId,
		"R2_SECRET_ACCESS_KEY": cfg.R2SecretAccessKey,
		"R2_BUCKET_NAME":       cfg.R2BucketName,
		"R2_ENDPOINT":          cfg.R2Endpoint,
	}

	missing := make([]string, 0, len(fields))
	configured := 0
	for key, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
			continue
		}
		configured++
	}

	if configured == 0 {
		if cfg.IsProd() {
			*issues = append(*issues, "R2 configuration is required in production")
		}
		return
	}

	if len(missing) > 0 {
		*issues = append(*issues, "R2 configuration is incomplete: missing "+strings.Join(missing, ", "))
	}
}
