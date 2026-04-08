// config/database.go
package config

import (
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
}
