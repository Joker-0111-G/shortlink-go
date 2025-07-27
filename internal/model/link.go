// 文件路径: internal/model/link.go
package model

import (
	"gorm.io/gorm"
	"time"
)

type Link struct {
	ID          uint64         `gorm:"primaryKey"`
	OriginalURL string         `gorm:"column:original_url;type:varchar(2048);not null"`
	ShortCode   string         `gorm:"column:short_code;type:varchar(10);uniqueIndex;not null"`
	ExpiresAt   *time.Time     `gorm:"column:expires_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (Link) TableName() string {
	return "links"
}