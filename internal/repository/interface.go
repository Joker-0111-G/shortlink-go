package repository

import (
	"context"
	"gorm.io/gorm" // 导入 gorm
	"shortlink-go/internal/model"
)


type LinkRepository interface {
	Save(ctx context.Context, link *model.Link) error
	FindByShortCode(ctx context.Context, shortCode string) (*model.Link, error)
	FindAll(ctx context.Context) ([]model.Link, error)
	SoftDeleteExpired(ctx context.Context) (int64, error)
	FindAndLockOldestReusable(ctx context.Context, tx *gorm.DB) (*model.Link, error)
	SaveTx(ctx context.Context, tx *gorm.DB, link *model.Link) error
	FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) // <--- 新增这行
}