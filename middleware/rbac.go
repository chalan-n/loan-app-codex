package middleware

import (
	"loan-app/config"
	"loan-app/models"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission ตรวจสอบว่าผู้ใช้มีสิทธิ์ที่ต้องการหรือไม่
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึงข้อมูลผู้ใช้จาก context (จาก AuthMiddleware)
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: user not found",
			})
		}

		// ตรวจสอบสิทธิ์
		if !user.HasPermission(permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions",
				"required": permission,
			})
		}

		// ถ้ามีสิทธิ์ ให้ผ่าน
		return c.Next()
	}
}

// RequireRole ตรวจสอบว่าผู้ใช้มีบทบาทที่ต้องการหรือไม่
func RequireRole(roleName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: user not found",
			})
		}

		if !user.HasRole(roleName) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: insufficient role",
				"required": roleName,
			})
		}

		return c.Next()
	}
}

// RequireAnyRole ตรวจสอบว่าผู้ใช้มีบทบาทใดบทบาทหนึ่งที่ระบุหรือไม่
func RequireAnyRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: user not found",
			})
		}

		for _, roleName := range roles {
			if user.HasRole(roleName) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden: insufficient role",
			"required": strings.Join(roles, " or "),
		})
	}
}

// LoadUserMiddleware โหลดข้อมูลผู้ใช้ลงใน context หลังจากผ่านการตรวจสอบ JWT
func LoadUserMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึง username จาก token (จาก parseJWTUsername ใน helpers.go)
		username := c.Locals("username").(string)
		
		var user models.User
		if err := config.DB.
			Preload("Roles").
			Where("username = ? AND is_active = ?", username, true).
			First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: user not found or inactive",
			})
		}

		// เก็บข้อมูลผู้ใช้ไว้ใน context
		c.Locals("user", &user)
		return c.Next()
	}
}
