package handlers

import (
	"fmt"
	"loan-app/config"
	"loan-app/models"
	"loan-app/startup"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type healthCheckItem struct {
	Name    string
	Status  string
	Message string
}

type healthMetricItem struct {
	Label string
	Value string
}

// AdminHealthPage shows a safe operational snapshot for administrators.
func AdminHealthPage(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	now := time.Now()

	checks := []healthCheckItem{
		checkConfigHealth(cfg),
		checkDatabaseHealth(),
		checkSchemaHealth(),
		checkStorageHealth(cfg),
	}

	overall := "ok"
	for _, check := range checks {
		if check.Status == "fail" {
			overall = "fail"
			break
		}
		if check.Status == "warn" && overall == "ok" {
			overall = "warn"
		}
	}

	return c.Render("admin_health", fiber.Map{
		"Overall":     overall,
		"Checks":      checks,
		"Metrics":     collectHealthMetrics(cfg),
		"GeneratedAt": now,
		"CurrentRole": GetCurrentUserRole(c),
	})
}

func checkConfigHealth(cfg *config.AppConfig) healthCheckItem {
	if err := startup.ValidateConfig(cfg); err != nil {
		return healthCheckItem{
			Name:    "Application config",
			Status:  "fail",
			Message: err.Error(),
		}
	}
	return healthCheckItem{Name: "Application config", Status: "ok", Message: "required environment values are present"}
}

func checkDatabaseHealth() healthCheckItem {
	if config.DB == nil {
		return healthCheckItem{Name: "Database", Status: "fail", Message: "database connection is nil"}
	}
	sqlDB, err := config.DB.DB()
	if err != nil {
		return healthCheckItem{Name: "Database", Status: "fail", Message: err.Error()}
	}
	if err := sqlDB.Ping(); err != nil {
		return healthCheckItem{Name: "Database", Status: "fail", Message: err.Error()}
	}
	return healthCheckItem{Name: "Database", Status: "ok", Message: "ping succeeded"}
}

func checkSchemaHealth() healthCheckItem {
	if config.DB == nil {
		return healthCheckItem{Name: "Database schema", Status: "fail", Message: "database connection is nil"}
	}
	if err := startup.VerifyDatabase(config.DB); err != nil {
		return healthCheckItem{Name: "Database schema", Status: "fail", Message: err.Error()}
	}

	var latest models.SchemaMigration
	if err := config.DB.Order("applied_at desc").First(&latest).Error; err != nil {
		return healthCheckItem{Name: "Database schema", Status: "warn", Message: "schema verified, but latest migration could not be read"}
	}

	return healthCheckItem{
		Name:    "Database schema",
		Status:  "ok",
		Message: fmt.Sprintf("latest migration: %s (%s)", latest.Version, latest.Name),
	}
}

func checkStorageHealth(cfg *config.AppConfig) healthCheckItem {
	missing := missingR2Fields(cfg)
	if len(missing) == 0 {
		return healthCheckItem{Name: "File storage", Status: "ok", Message: "R2 configuration is complete"}
	}
	if len(missing) == 5 {
		return healthCheckItem{Name: "File storage", Status: "warn", Message: "R2 is not configured; uploads will fail until env is set"}
	}
	return healthCheckItem{Name: "File storage", Status: "fail", Message: "missing " + strings.Join(missing, ", ")}
}

func missingR2Fields(cfg *config.AppConfig) []string {
	fields := map[string]string{
		"R2_ACCOUNT_ID":        cfg.R2AccountId,
		"R2_ACCESS_KEY_ID":     cfg.R2AccessKeyId,
		"R2_SECRET_ACCESS_KEY": cfg.R2SecretAccessKey,
		"R2_BUCKET_NAME":       cfg.R2BucketName,
		"R2_ENDPOINT":          cfg.R2Endpoint,
	}

	missing := make([]string, 0, len(fields))
	for key, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func collectHealthMetrics(cfg *config.AppConfig) []healthMetricItem {
	metrics := []healthMetricItem{
		{Label: "Environment", Value: cfg.AppEnv},
		{Label: "Port", Value: cfg.AppPort},
		{Label: "Database", Value: cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName},
		{Label: "Upload max size", Value: fmt.Sprintf("%.1f MB", float64(cfg.UploadMaxFileSizeBytes)/(1024*1024))},
		{Label: "Upload rate limit", Value: fmt.Sprintf("%d requests / %d seconds", cfg.UploadRateLimitMax, cfg.UploadRateLimitWindowSeconds)},
		{Label: "Session idle timeout", Value: fmt.Sprintf("%d minutes", cfg.SessionIdleTimeoutMinutes)},
		{Label: "WebAuthn origin", Value: cfg.WebAuthnOrigin},
	}

	if config.DB != nil {
		var loanCount int64
		var auditCount int64
		config.DB.Model(&models.LoanApplication{}).Count(&loanCount)
		config.DB.Model(&models.AuditLog{}).Count(&auditCount)
		metrics = append(metrics,
			healthMetricItem{Label: "Loan applications", Value: fmt.Sprintf("%d", loanCount)},
			healthMetricItem{Label: "Audit events", Value: fmt.Sprintf("%d", auditCount)},
		)
	}

	return metrics
}
