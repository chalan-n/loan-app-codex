// handlers/rbac.go
package handlers

import (
	"loan-app/models"

	"github.com/gofiber/fiber/v2"
)

// RequireRole สร้าง middleware ที่อนุญาตเฉพาะ role ที่ระบุ
// ใช้งาน: app.Get("/admin", RequireRole(models.RoleAdmin), handler)
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := parseJWTUsername(c.Cookies("token"))
		role := getUserRole(username)

		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		// ถ้าเป็น API request ส่ง JSON
		if c.Get("Accept") == "application/json" || len(c.Get("X-API-Key")) > 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "ไม่มีสิทธิ์เข้าถึง",
			})
		}

		// ถ้าเป็น HTML request ส่งกลับหน้า main พร้อม error
		return c.Redirect("/main?error=forbidden")
	}
}

// RequireManagerOrAbove อนุญาต manager และ admin
func RequireManagerOrAbove() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := getUserRole(parseJWTUsername(c.Cookies("token")))
		if models.IsManagerOrAbove(role) {
			return c.Next()
		}

		if c.Get("Accept") == "application/json" || len(c.Get("X-API-Key")) > 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "ไม่มีสิทธิ์เข้าถึง",
			})
		}

		return c.Redirect("/main?error=forbidden")
	}
}

// RequireAdmin อนุญาตเฉพาะ admin
func RequireAdmin() fiber.Handler {
	return RequireRole(models.RoleAdmin)
}

// GetCurrentUserRole helper สำหรับ template rendering
func GetCurrentUserRole(c *fiber.Ctx) string {
	username := parseJWTUsername(c.Cookies("token"))
	return getUserRole(username)
}
