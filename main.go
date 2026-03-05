package main

import (
	"fmt"
	"loan-app/config"
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

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Add CORS middleware for cross-domain requests
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Static("/static", "./static")

	config.ConnectDB()

	// === ย้ายมาอยู่ตรงนี้เลย ปลอดภัย 100% ไม่มี cycle ===
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), 12)
		config.DB.Create(&models.User{
			Username: "570639",
			Password: string(hashed),
		})
		log.Println("สร้างผู้ใช้สำเร็จ: 570639 / 123456")

		config.DB.Create(&models.User{
			Username: "580639",
			Password: string(hashed),
		})
		log.Println("สร้างผู้ใช้สำเร็จ: 580639 / 123456")
	}
	// ================================================

	routes.Setup(app)

	// ── TLS: สร้าง self-signed cert (ถ้ายังไม่มี) แล้ว listen HTTPS ────
	cert, key := config.TLSCertFiles()

	// แสดง URL ที่เข้าถึงได้ทั้งหมด
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
	fmt.Println("╚═══════════════════════════════════════════════════╝\n")

	log.Fatal(app.ListenTLS("0.0.0.0:3000", cert, key))
}
