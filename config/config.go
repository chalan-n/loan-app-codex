// config/config.go
package config

import (
	"fmt"
	"log"
	"os"

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
	AppPort string
	AppEnv  string

	// JWT
	JWTSecret string
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

		AppPort: getEnv("APP_PORT", "3000"),
		AppEnv:  getEnv("APP_ENV", "development"),

		JWTSecret: getEnv("JWT_SECRET", "changeme-secret"),
	}

	log.Printf("[config] ✅ ENV=%s | DB=%s:%s/%s | Port=%s",
		cfg.AppEnv, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.AppPort)

	if cfg.GeminiAPIKey == "" {
		log.Println("[config] ⚠️  GEMINI_API_KEY ยังไม่ได้ตั้งค่า")
	}
	if cfg.IsProd() && cfg.JWTSecret == "changeme-secret" {
		log.Println("[config] ⚠️  JWT_SECRET กำลังใช้ค่า default ใน production — ควรเปลี่ยนด่วน!")
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
