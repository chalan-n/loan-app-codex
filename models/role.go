// models/role.go
package models

// Role constants
const (
	RoleOfficer = "officer" // พนักงานสินเชื่อ — เห็นแค่งานตัวเอง
	RoleManager = "manager" // ผู้จัดการ — เห็นงานทุกคน + dashboard
	RoleAdmin   = "admin"   // ผู้ดูแลระบบ — จัดการ users + audit log
)

var validRoles = map[string]struct{}{
	RoleOfficer: {},
	RoleManager: {},
	RoleAdmin:   {},
}

func IsValidRole(role string) bool {
	_, ok := validRoles[role]
	return ok
}

func NormalizeRole(role string) string {
	if !IsValidRole(role) {
		return RoleOfficer
	}
	return role
}

func IsManagerOrAbove(role string) bool {
	return role == RoleManager || role == RoleAdmin
}
