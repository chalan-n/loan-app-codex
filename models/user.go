// models/user.go
package models

import "time"

type User struct {
	ID                    uint   `gorm:"primaryKey"`
	Username              string `gorm:"unique"`
	Password              string
	Role                  string     `gorm:"type:varchar(20);default:'officer'"`
	CurrentSessionID      string     `gorm:"type:varchar(36)"`
	SessionLastActivityAt *time.Time `gorm:"index"`
	SessionRevokedAt      *time.Time `gorm:"index"`
}
