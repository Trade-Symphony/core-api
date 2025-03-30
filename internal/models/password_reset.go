package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        string         `gorm:"primaryKey"`
	UserID    uint          `gorm:"index"`
	Token     string        `gorm:"not null"`
	ExpiresAt time.Time     `gorm:"not null"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
