package handlers

import (
	"loan-app/config"
	"loan-app/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *fiber.Ctx) error {
	if parseJWTUsername(c.Cookies("token")) != "" {
		return c.Redirect("/main")
	}

	errorMsg := c.Query("error")
	if errorMsg == "" && c.Query("status") == "error" {
		errorMsg = "invalid_credentials"
	}

	return c.Render("login", fiber.Map{
		"Error": errorMsg,
	})
}

func LoginPost(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	user := new(models.User)
	config.DB.Where("username = ?", username).First(user)

	if user.ID == 0 || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return c.Redirect("/login?error=invalid_credentials")
	}

	tokenStr, err := issueUserSession(user)
	if err != nil {
		return c.Status(500).SendString("ไม่สามารถสร้าง token ได้")
	}

	setAuthCookie(c, tokenStr)
	WriteAuditAs(c, user.Username, "login", "", "เข้าสู่ระบบสำเร็จ")
	return c.Redirect("/main")
}

func createJWTToken(username, sessionID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":   username,
		"session_id": sessionID,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(config.GetConfig().JWTSecret))
}

func Logout(c *fiber.Ctx) error {
	WriteAudit(c, "logout", "", "ออกจากระบบ")
	username := parseJWTUsername(c.Cookies("token"))
	if username != "" {
		config.DB.Model(&models.User{}).Where("username = ?", username).Update("current_session_id", "")
	}
	clearAuthCookie(c)
	return c.Redirect("/login")
}

func AuthMiddleware(c *fiber.Ctx) error {
	tokenStr := c.Cookies("token")
	if tokenStr == "" {
		return c.Redirect("/login")
	}

	username := parseJWTUsername(tokenStr)
	sessionID := parseJWTSessionID(tokenStr)
	if username == "" || sessionID == "" {
		clearAuthCookie(c)
		return c.Redirect("/login")
	}

	var user models.User
	if err := config.DB.Select("username, current_session_id").Where("username = ?", username).First(&user).Error; err != nil || user.CurrentSessionID == "" || user.CurrentSessionID != sessionID {
		clearAuthCookie(c)
		return c.Redirect("/login")
	}

	c.Locals("username", username)
	return c.Next()
}

func ChangePasswordPage(c *fiber.Ctx) error {
	return c.Render("change_password", nil)
}

func MobileAPIKeyMiddleware(c *fiber.Ctx) error {
	key := c.Get("X-API-Key")
	expected := config.GetConfig().MobileAPIKey
	if expected == "" || key != expected {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: invalid or missing API key"})
	}
	return c.Next()
}

func ChangePasswordPost(c *fiber.Ctx) error {
	oldPassword := c.FormValue("old_password")
	newPassword := c.FormValue("new_password")

	username := parseJWTUsername(c.Cookies("token"))
	if username == "" {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "รหัสผ่านเดิมไม่ถูกต้อง"})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	user.Password = string(hashedPassword)
	user.CurrentSessionID = ""

	if err := config.DB.Save(&user).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	clearAuthCookie(c)
	return c.JSON(fiber.Map{"success": true, "message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว"})
}
