// config/database.go
package config

import (
	"loan-app/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	cfg := GetConfig()
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		panic("เชื่อม DB ไม่ได้: " + err.Error())
	}
	DB = db

	// Auto migrate all models
	db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.LoanApplication{},
		&models.Guarantor{},
	)

	// Seed default roles
	seedRoles()
}

// seedRoles สร้างบทบาทเริ่มต้นถ้ายังไม่มี
func seedRoles() {
	for _, role := range models.DefaultRoles {
		var existing models.Role
		if err := DB.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			DB.Create(&role)
		}
	}
}
