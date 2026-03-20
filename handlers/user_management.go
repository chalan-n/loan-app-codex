package handlers

import (
	"loan-app/config"
	"loan-app/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// UsersPage หน้าแสดงรายการผู้ใช้
func UsersPage(c *fiber.Ctx) error {
	var users []models.User
	config.DB.Preload("Roles").Find(&users)

	return c.Render("users", fiber.Map{
		"Users": users,
		"Title": "จัดการผู้ใช้งาน",
	})
}

// CreateUser สร้างผู้ใช้ใหม่
func CreateUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	phone := c.FormValue("phone")

	// อ่านค่า role_ids จาก form (checkboxes)
	c.Request().Header.Add("Content-Type", "multipart/form-data")
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "อ่านข้อมูลฟอร์มไม่สำเร็จ"})
	}
	roleIDs := form.Value["role_ids"]

	// เข้ารหัสรหัสผ่าน
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
	}

	// สร้างผู้ใช้
	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		FullName: fullName,
		Email:    email,
		Phone:    phone,
		IsActive: true,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "สร้างผู้ใช้ไม่สำเร็จ: " + err.Error()})
	}

	// กำหนดบทบาท
	for _, roleIDStr := range roleIDs {
		roleID, _ := strconv.ParseUint(roleIDStr, 10, 32)
		config.DB.Create(&models.UserRole{
			UserID: user.ID,
			RoleID: uint(roleID),
		})
	}

	return c.JSON(fiber.Map{"success": true, "message": "สร้างผู้ใช้สำเร็จ"})
}

// UpdateUser อัปเดตข้อมูลผู้ใช้
func UpdateUser(c *fiber.Ctx) error {
	userID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "ไม่พบผู้ใช้"})
	}

	// อัปเดตข้อมูลพื้นฐาน
	user.FullName = c.FormValue("full_name")
	user.Email = c.FormValue("email")
	user.Phone = c.FormValue("phone")
	user.IsActive = c.FormValue("is_active") == "true"

	// ถ้ามีการเปลี่ยนรหัสผ่าน
	newPassword := c.FormValue("password")
	if newPassword != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
		user.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "อัปเดตผู้ใช้ไม่สำเร็จ"})
	}

	// อัปเดตบทบาท (ลบเก่าแล้วเพิ่มใหม่)
	config.DB.Where("user_id = ?", user.ID).Delete(&models.UserRole{})

	// อ่านค่า role_ids จาก form
	form, err := c.MultipartForm()
	if err == nil {
		roleIDs := form.Value["role_ids"]
		for _, roleIDStr := range roleIDs {
			roleID, _ := strconv.ParseUint(roleIDStr, 10, 32)
			config.DB.Create(&models.UserRole{
				UserID: user.ID,
				RoleID: uint(roleID),
			})
		}
	}

	return c.JSON(fiber.Map{"success": true, "message": "อัปเดตผู้ใช้สำเร็จ"})
}

// DeleteUser ลบผู้ใช้
func DeleteUser(c *fiber.Ctx) error {
	userID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	// ลบความสัมพันธ์บทบาทก่อน
	config.DB.Where("user_id = ?", userID).Delete(&models.UserRole{})

	// ลบผู้ใช้
	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "ลบผู้ใช้ไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "ลบผู้ใช้สำเร็จ"})
}

// RolesPage หน้าแสดงรายการบทบาท
func RolesPage(c *fiber.Ctx) error {
	var roles []models.Role
	config.DB.Find(&roles)

	return c.Render("roles", fiber.Map{
		"Roles": roles,
		"Title": "จัดการบทบาท",
	})
}

// CreateRole สร้างบทบาทใหม่
func CreateRole(c *fiber.Ctx) error {
	name := c.FormValue("name")
	description := c.FormValue("description")

	// อ่านค่า permissions จาก form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "อ่านข้อมูลฟอร์มไม่สำเร็จ"})
	}
	permissions := form.Value["permissions"]

	role := models.Role{
		Name:        name,
		Description: description,
		Permissions: permissions,
	}

	if err := config.DB.Create(&role).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "สร้างบทบาทไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "สร้างบทบาทสำเร็จ"})
}

// UpdateRole อัปเดตบทบาท
func UpdateRole(c *fiber.Ctx) error {
	roleID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	var role models.Role
	if err := config.DB.First(&role, roleID).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "ไม่พบบทบาท"})
	}

	role.Name = c.FormValue("name")
	role.Description = c.FormValue("description")

	// อ่านค่า permissions จาก form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "อ่านข้อมูลฟอร์มไม่สำเร็จ"})
	}
	role.Permissions = form.Value["permissions"]

	if err := config.DB.Save(&role).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "อัปเดตบทบาทไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "อัปเดตบทบาทสำเร็จ"})
}

// DeleteRole ลบบทบาท
func DeleteRole(c *fiber.Ctx) error {
	roleID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	// ตรวจสอบว่ามีผู้ใช้ใช้บทบาทนี้อยู่หรือไม่
	var count int64
	config.DB.Table("user_roles").Where("role_id = ?", roleID).Count(&count)
	if count > 0 {
		return c.JSON(fiber.Map{"success": false, "message": "ไม่สามารถลบบทบาทที่มีผู้ใช้อยู่ได้"})
	}

	if err := config.DB.Delete(&models.Role{}, roleID).Error; err != nil {
		return c.JSON(fiber.Map{"success": false, "message": "ลบบทบาทไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "ลบบทบาทสำเร็จ"})
}
