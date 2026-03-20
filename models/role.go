package models

import (
	"time"
)

// Role บทบาทของผู้ใช้งาน
type Role struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"unique;not null"`        // admin, manager, officer
	Description string    `gorm:"size:255"`               // คำอธิบายบทบาท
	Permissions []string  `gorm:"type:json"`              // ["view_all", "approve", "delete", "manage_users"]
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// UserRole การเชื่อมโยงผู้ใช้กับบทบาท (Many-to-Many)
type UserRole struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	RoleID    uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ค่าคงที่สำหรับสิทธิ์
const (
	// สิทธิ์ในการดูข้อมูล
	PermissionViewAll      = "view_all"        // ดูข้อมูลทั้งหมด
	PermissionViewOwn      = "view_own"        // ดูเฉพาะของตัวเอง
	PermissionViewReports  = "view_reports"    // ดูรายงาน
	
	// สิทธิ์ในการจัดการคำขอ
	PermissionCreateLoan   = "create_loan"     // สร้างคำขอใหม่
	PermissionEditLoan     = "edit_loan"       // แก้ไขคำขอ
	PermissionDeleteLoan   = "delete_loan"     // ลบคำขอ
	PermissionApprove      = "approve"         // อนุมัติคำขอ
	PermissionReject       = "reject"          // ปฏิเสธคำขอ
	
	// สิทธิ์ในการจัดการผู้ใช้
	PermissionManageUsers = "manage_users"    // จัดการผู้ใช้
	PermissionManageRoles  = "manage_roles"    // จัดการบทบาท
	PermissionViewAudit    = "view_audit"      // ดูประวัติการใช้งาน
)

// บทบาทเริ่มต้นที่สร้างอัตโนมัติ
var DefaultRoles = []Role{
	{
		Name:        "admin",
		Description: "ผู้ดูแลระบบ - มีสิทธิ์ทั้งหมด",
		Permissions: []string{
			PermissionViewAll, PermissionViewReports,
			PermissionCreateLoan, PermissionEditLoan, PermissionDeleteLoan,
			PermissionApprove, PermissionReject,
			PermissionManageUsers, PermissionManageRoles, PermissionViewAudit,
		},
	},
	{
		Name:        "manager",
		Description: "ผู้จัดการ - ดูและอนุมัติคำขอ",
		Permissions: []string{
			PermissionViewAll, PermissionViewReports,
			PermissionEditLoan, PermissionApprove, PermissionReject,
			PermissionViewAudit,
		},
	},
	{
		Name:        "officer",
		Description: "เจ้าหน้าที่ - สร้างและจัดการคำขอ",
		Permissions: []string{
			PermissionViewOwn, PermissionCreateLoan, PermissionEditLoan,
		},
	},
}
