// models/user.go
package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	FullName  string    `gorm:"size:100"`     // ชื่อ-นามสกุล
	Email     string    `gorm:"size:100"`     // อีเมล
	Phone     string    `gorm:"size:20"`      // เบอร์โทรศัพท์
	IsActive  bool      `gorm:"default:true"` // สถานะการใช้งาน
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	Roles []Role `gorm:"many2many:user_roles;"`
}

// HasPermission ตรวจสอบว่าผู้ใช้มีสิทธิ์ที่ต้องการหรือไม่
func (u *User) HasPermission(permission string) bool {
	for _, role := range u.Roles {
		for _, p := range role.Permissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

// HasRole ตรวจสอบว่าผู้ใช้มีบทบาทที่ต้องการหรือไม่
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
