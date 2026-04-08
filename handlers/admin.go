// handlers/admin.go
package handlers

import (
	"loan-app/config"
	"loan-app/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ── Admin: Audit Log ─────────────────────────────────────────────────────────

func AuditLogPage(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	action := c.Query("action")
	username := c.Query("username")

	q := config.DB.Model(&models.AuditLog{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}

	var total int64
	q.Count(&total)

	var logs []models.AuditLog
	q.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs)

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.Render("admin_audit", fiber.Map{
		"Logs":           logs,
		"Page":           page,
		"TotalPages":     totalPages,
		"Total":          total,
		"FilterAction":   action,
		"FilterUsername": username,
		"CurrentRole":    GetCurrentUserRole(c),
	})
}

// ── Admin: User Management ────────────────────────────────────────────────────

func AdminUsersPage(c *fiber.Ctx) error {
	var users []models.User
	config.DB.Select("id, username, role").Find(&users)
	return c.Render("admin_users", fiber.Map{
		"Users":       users,
		"CurrentRole": GetCurrentUserRole(c),
	})
}

func AdminUpdateUserRole(c *fiber.Ctx) error {
	type Req struct {
		UserID uint   `json:"user_id"`
		Role   string `json:"role"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if !models.IsValidRole(req.Role) {
		return c.Status(400).JSON(fiber.Map{"error": "Role ไม่ถูกต้อง"})
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", req.UserID).
		Update("role", req.Role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกไม่สำเร็จ"})
	}

	WriteAudit(c, "update_user_role", "", "userID="+strconv.Itoa(int(req.UserID))+" role="+req.Role)
	return c.JSON(fiber.Map{"success": true})
}

func AdminCreateUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	role := c.FormValue("role")

	if username == "" || password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณากรอก username และ password"})
	}
	role = models.NormalizeRole(role)

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	user := models.User{Username: username, Password: string(hashed), Role: role}
	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง user ไม่สำเร็จ (อาจซ้ำ)"})
	}

	WriteAudit(c, "create_user", "", "username="+username+" role="+role)
	return c.JSON(fiber.Map{"success": true, "id": user.ID})
}

func AdminDeleteUser(c *fiber.Ctx) error {
	type Req struct {
		UserID uint `json:"user_id"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// ป้องกันลบตัวเอง
	selfUsername := parseJWTUsername(c.Cookies("token"))
	var self models.User
	config.DB.Select("id").Where("username = ?", selfUsername).First(&self)
	if self.ID == req.UserID {
		return c.Status(400).JSON(fiber.Map{"error": "ไม่สามารถลบบัญชีของตัวเองได้"})
	}

	var user models.User
	if err := config.DB.First(&user, req.UserID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบ user"})
	}
	if err := config.DB.Delete(&models.User{}, req.UserID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ลบ user ไม่สำเร็จ"})
	}
	WriteAudit(c, "delete_user", "", "username="+user.Username)
	return c.JSON(fiber.Map{"success": true})
}

// ── Manager: Dashboard ────────────────────────────────────────────────────────

func ManagerDashboard(c *fiber.Ctx) error {
	today := time.Now().Format("2006-01-02")
	firstDayOfMonth := time.Now().Format("2006-01") + "-01"

	// สถิติรวม
	var totalLoans, todayLoans, monthLoans int64
	var pendingLoans, approvedLoans, rejectedLoans int64

	config.DB.Model(&models.LoanApplication{}).Count(&totalLoans)
	config.DB.Model(&models.LoanApplication{}).Where("DATE(submitted_date) = ?", today).Count(&todayLoans)
	config.DB.Model(&models.LoanApplication{}).Where("submitted_date >= ?", firstDayOfMonth).Count(&monthLoans)
	config.DB.Model(&models.LoanApplication{}).Where("status = ?", models.LoanStatusPending).Count(&pendingLoans)
	config.DB.Model(&models.LoanApplication{}).Where("status = ?", models.LoanStatusApproved).Count(&approvedLoans)
	config.DB.Model(&models.LoanApplication{}).Where("status = ?", models.LoanStatusRejected).Count(&rejectedLoans)

	// ยอดสินเชื่อรวมเดือนนี้
	type SumResult struct {
		Total float64
	}
	var loanSum SumResult
	config.DB.Model(&models.LoanApplication{}).
		Where("submitted_date >= ?", firstDayOfMonth).
		Select("COALESCE(SUM(loan_amount), 0) as total").
		Scan(&loanSum)

	// Top 5 staff ที่ยื่นมากสุดเดือนนี้
	type StaffStat struct {
		StaffID string
		Count   int64
	}
	var topStaff []StaffStat
	config.DB.Model(&models.LoanApplication{}).
		Where("submitted_date >= ?", firstDayOfMonth).
		Select("staff_id, COUNT(*) as count").
		Group("staff_id").
		Order("count desc").
		Limit(5).
		Scan(&topStaff)

	// กราฟ: จำนวนสินเชื่อ 7 วันล่าสุด
	type DayStat struct {
		Day   string
		Count int64
	}
	var dailyStats []DayStat
	config.DB.Model(&models.LoanApplication{}).
		Where("submitted_date >= ?", time.Now().AddDate(0, 0, -6).Format("2006-01-02")).
		Select("DATE(submitted_date) as day, COUNT(*) as count").
		Group("day").
		Order("day asc").
		Scan(&dailyStats)

	// กิจกรรมล่าสุด (audit log)
	var recentLogs []models.AuditLog
	config.DB.Order("created_at desc").Limit(10).Find(&recentLogs)

	return c.Render("dashboard", fiber.Map{
		"TotalLoans":    totalLoans,
		"TodayLoans":    todayLoans,
		"MonthLoans":    monthLoans,
		"PendingLoans":  pendingLoans,
		"ApprovedLoans": approvedLoans,
		"RejectedLoans": rejectedLoans,
		"LoanSumMonth":  loanSum.Total,
		"TopStaff":      topStaff,
		"DailyStats":    dailyStats,
		"RecentLogs":    recentLogs,
		"CurrentRole":   GetCurrentUserRole(c),
	})
}
