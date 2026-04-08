// config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig เก็บค่าคอนฟิกทั้งหมดของระบบ
type AppConfig struct {
	// Gemini AI
	GeminiAPIKey string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Server
	AppPort          string
	AppEnv           string
	CORSAllowOrigins string

	// JWT
	JWTSecret                     string
	SessionIdleTimeoutMinutes     int
	SessionActivityRefreshSeconds int

	// Rate limiting
	LoginRateLimitMax               int
	LoginRateLimitWindowSeconds     int
	SearchRateLimitMax              int
	SearchRateLimitWindowSeconds    int
	InsuranceRateLimitMax           int
	InsuranceRateLimitWindowSeconds int
	UploadRateLimitMax              int
	UploadRateLimitWindowSeconds    int

	// Mobile API
	MobileAPIKey string

	// Cloudflare R2
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	R2Endpoint        string

	// Seed
	DefaultAdminPassword string

	// WebAuthn
	WebAuthnRPID   string
	WebAuthnOrigin string
}

// cfg เป็น singleton ที่เก็บค่าคอนฟิกไว้ภายใน package
var cfg *AppConfig

// GetConfig โหลดไฟล์ .env ที่เหมาะสม แล้วคืน *AppConfig (singleton)
//
// ลำดับการโหลด (priority สูง → ต่ำ):
//  1. ถ้า APP_ENV ถูก set ใน OS env อยู่แล้ว → โหลด .env.{APP_ENV}
//     เช่น:  APP_ENV=production ./loan-app  →  โหลด .env.production
//  2. ถ้ามีไฟล์ .env อยู่ → โหลด .env
//  3. fallback → ใช้ OS environment variables ตรงๆ (Docker / systemd)
func GetConfig() *AppConfig {
	if cfg != nil {
		return cfg
	}

	loadEnvFile()

	cfg = &AppConfig{
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "loan_db"),

		AppPort:          getEnv("APP_PORT", "3000"),
		AppEnv:           getEnv("APP_ENV", "development"),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),

		JWTSecret:                       getEnv("JWT_SECRET", "changeme-secret"),
		SessionIdleTimeoutMinutes:       getEnvInt("SESSION_IDLE_TIMEOUT_MINUTES", 30),
		SessionActivityRefreshSeconds:   getEnvInt("SESSION_ACTIVITY_REFRESH_SECONDS", 300),
		LoginRateLimitMax:               getEnvInt("LOGIN_RATE_LIMIT_MAX", 5),
		LoginRateLimitWindowSeconds:     getEnvInt("LOGIN_RATE_LIMIT_WINDOW_SECONDS", 300),
		SearchRateLimitMax:              getEnvInt("SEARCH_RATE_LIMIT_MAX", 60),
		SearchRateLimitWindowSeconds:    getEnvInt("SEARCH_RATE_LIMIT_WINDOW_SECONDS", 60),
		InsuranceRateLimitMax:           getEnvInt("INSURANCE_RATE_LIMIT_MAX", 30),
		InsuranceRateLimitWindowSeconds: getEnvInt("INSURANCE_RATE_LIMIT_WINDOW_SECONDS", 60),
		UploadRateLimitMax:              getEnvInt("UPLOAD_RATE_LIMIT_MAX", 20),
		UploadRateLimitWindowSeconds:    getEnvInt("UPLOAD_RATE_LIMIT_WINDOW_SECONDS", 300),

		MobileAPIKey: getEnv("MOBILE_API_KEY", ""),

		R2AccountId:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyId:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "cmo-loan-app"),
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),

		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", ""),

		WebAuthnRPID:   getEnv("WEBAUTHN_RPID", "localhost"),
		WebAuthnOrigin: getEnv("WEBAUTHN_ORIGIN", "https://localhost:3000"),
	}

	log.Printf("[config] ✅ ENV=%s | DB=%s:%s/%s | Port=%s",
		cfg.AppEnv, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.AppPort)

	if cfg.GeminiAPIKey == "" {
		log.Println("[config] ⚠️  GEMINI_API_KEY ยังไม่ได้ตั้งค่า")
	}
	if cfg.IsProd() && cfg.JWTSecret == "changeme-secret" {
		log.Println("[config] ⚠️  JWT_SECRET กำลังใช้ค่า default ใน production — ควรเปลี่ยนด่วน!")
	}
	if cfg.R2AccountId == "" {
		log.Println("[config] ⚠️  R2_ACCOUNT_ID ยังไม่ได้ตั้งค่า — ฟีเจอร์อัปโหลดไฟล์จะไม่ทำงาน")
	}
	if cfg.MobileAPIKey == "" {
		log.Println("[config] ⚠️  MOBILE_API_KEY ยังไม่ได้ตั้งค่า — /api/update-status จะปฏิเสธทุก request")
	}

	return cfg
}

// loadEnvFile เลือกไฟล์ .env ที่จะโหลดตาม priority
func loadEnvFile() {
	// Priority 1: APP_ENV set ใน OS → โหลด .env.{APP_ENV}
	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		envFile := ".env." + appEnv
		if err := godotenv.Load(envFile); err == nil {
			log.Printf("[config] 📄 โหลด %s", envFile)
			return
		}
		log.Printf("[config] ⚠️  ไม่พบ %s — ลองโหลด .env แทน", envFile)
	}

	// Priority 2: โหลด .env ทั่วไป
	if err := godotenv.Load(".env"); err == nil {
		log.Println("[config] 📄 โหลด .env")
		return
	}

	// Priority 3: ไม่มีไฟล์ → ใช้ OS env ตรงๆ (Docker / systemd / VPS export)
	log.Println("[config] 📄 ไม่พบไฟล์ .env — ใช้ OS environment variables")
}

// DSN สร้าง MySQL DSN string จากค่าคอนฟิก
func (c *AppConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// IsProd คืน true ถ้ากำลังรันใน production
func (c *AppConfig) IsProd() bool {
	return c.AppEnv == "production"
}

// getEnv คืนค่า env var หรือ fallback ถ้าไม่มี
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
