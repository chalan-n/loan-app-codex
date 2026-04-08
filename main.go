package main

import (
	"encoding/json"
	"fmt"
	"loan-app/config"
	"loan-app/handlers"
	"loan-app/models"
	"loan-app/routes"
	"log"
	"net"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
	"golang.org/x/crypto/bcrypt"
)

func thaiDate(dateStr string) string {
	if dateStr == "" {
		return "-"
	}

	// 1. Try ISO Format (YYYY-MM-DD) or DateTime
	formatsIso := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, f := range formatsIso {
		t, err := time.Parse(f, dateStr)
		if err == nil {
			// ISO is usually AD, convert to BE
			thaiMonths := []string{
				"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
				"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
			}
			year := t.Year() + 543
			return fmt.Sprintf("%d %s %d", t.Day(), thaiMonths[t.Month()], year)
		}
	}

	// 2. Try Thai Input Format (DD-MM-YYYY)
	// Example: 13-01-2568
	formatsThai := []string{"02-01-2006", "02/01/2006"}
	for _, f := range formatsThai {
		t, err := time.Parse(f, dateStr)
		if err == nil {
			thaiMonths := []string{
				"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
				"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
			}
			year := t.Year()
			// If Year is < 2400, assume AD and add 543.
			// If > 2400, assume already BE (e.g. 2568).
			if year < 2400 {
				year += 543
			}
			return fmt.Sprintf("%d %s %d", t.Day(), thaiMonths[t.Month()], year)
		}
	}

	return dateStr // Return original if all parses fail
}

func main() {
	engine := html.New("./templates", ".html")
	engine.AddFunc("thaiDate", thaiDate)
	engine.AddFunc("add", func(a, b int) int { return a + b })
	engine.AddFunc("jsonEncode", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})
	engine.AddFunc("seq", func(start, end int) []int {
		s := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			s = append(s, i)
		}
		return s
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Add CORS middleware for cross-domain requests
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.GetConfig().CORSAllowOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Idempotency-Key",
	}))

	app.Static("/static", "./static")

	config.ConnectDB()

	// === AutoMigrate — run once at startup ===
	config.DB.AutoMigrate(&models.Guarantor{}, &models.RefRunning{}, &models.LoanFile{})
	// =============================================

	// === WebAuthn Init ===
	waCfg := config.GetConfig()
	if err := handlers.InitWebAuthn(waCfg.WebAuthnRPID, waCfg.WebAuthnOrigin, "CMO APP"); err != nil {
		log.Fatalf("[webauthn] init failed: %v", err)
	}
	log.Printf("[webauthn] ✅ RPID=%s Origin=%s", waCfg.WebAuthnRPID, waCfg.WebAuthnOrigin)
	// =====================

	// === Seed default users with roles ===
	defaultPass := config.GetConfig().DefaultAdminPassword
	if defaultPass != "" {
		var count int64
		config.DB.Model(&models.User{}).Count(&count)
		if count == 0 {
			hashed, _ := bcrypt.GenerateFromPassword([]byte(defaultPass), 12)
			config.DB.Create(&models.User{Username: "570639", Password: string(hashed), Role: models.RoleAdmin})
			config.DB.Create(&models.User{Username: "580639", Password: string(hashed), Role: models.RoleOfficer})
			log.Println("[seed] สร้างผู้ใช้เริ่มต้นสำเร็จ")
		}
	} else {
		log.Println("[seed] DEFAULT_ADMIN_PASSWORD ไม่ได้ตั้งค่า — ข้ามการสร้างผู้ใช้เริ่มต้น")
	}
	// =============================================================

	routes.Setup(app)

	// ── TLS: สร้าง self-signed cert (ถ้ายังไม่มี) แล้ว listen HTTPS ────
	cert, key := config.TLSCertFiles()

	if config.GetConfig().IsProd() {
		// Production: รัน HTTP (ให้ nginx จัดการ TLS)
		log.Fatal(app.Listen("0.0.0.0:3000"))
	} else {
		// Development: รัน HTTPS สำหรับทดสอบ
		fmt.Println("\n╔═══════════════════════════════════════════════════╗")
		fmt.Println("║           🚀 HTTPS Server Ready                  ║")
		fmt.Println("╠═══════════════════════════════════════════════════╣")
		fmt.Printf("║  Local:   https://localhost:3000                  ║\n")
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
				fmt.Printf("║  LAN:     https://%-15s:3000       ║\n", ipNet.IP)
			}
		}
		fmt.Println("║                                                   ║")
		fmt.Println("║  ⚠️  Tablet: ยอมรับ cert แล้วเปิดกล้องได้          ║")
		fmt.Println("╚═══════════════════════════════════════════════════╝")
		fmt.Println()
		log.Fatal(app.ListenTLS("0.0.0.0:3000", cert, key))
	}
}
