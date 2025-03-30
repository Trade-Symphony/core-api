package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        string         `gorm:"primaryKey"`
	UserID    uint          `gorm:"index"`
	Expiry    time.Time     `gorm:"not null"`
	UserAgent string        `gorm:"not null"`
	IP        string        `gorm:"not null"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
