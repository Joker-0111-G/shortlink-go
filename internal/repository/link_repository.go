// 文件路径: internal/repository/link_repository.go
package repository

import (
	"context"
	"errors"
	"shortlink-go/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// linkRepository 实现了 LinkRepository 接口
type linkRepository struct {
	db *gorm.DB
}

// NewLinkRepository 是 linkRepository 的构造函数
func NewLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepository{db: db}
}

func (r *linkRepository) Save(ctx context.Context, link *model.Link) error {
	return r.db.WithContext(ctx).Save(link).Error
}

func (r *linkRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	var link model.Link
	now := time.Now()
	// GORM的软删除会自动处理 "deleted_at IS NULL"
	err := r.db.WithContext(ctx).Where("short_code = ? AND (expires_at IS NULL OR expires_at > ?)", shortCode, now).First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &link, nil
}


// FindAll 按即将过期排序
func (r *linkRepository) FindAll(ctx context.Context) ([]model.Link, error) {
	var links []model.Link
	// GORM的软删除会自动处理 "deleted_at IS NULL"
	err := r.db.WithContext(ctx).Order("CASE WHEN expires_at IS NULL THEN 1 ELSE 0 END, expires_at ASC").Find(&links).Error
	return links, err
}

// SoftDeleteExpired 软删除过期链接
func (r *linkRepository) SoftDeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.Link{}).
		Where("expires_at <= ? AND deleted_at IS NULL", now).
		Update("deleted_at", now)
	return result.RowsAffected, result.Error
}

// FindAndLockOldestReusable 在事务中查找并锁定一个最旧的可复用记录
func (r *linkRepository) FindAndLockOldestReusable(ctx context.Context, tx *gorm.DB) (*model.Link, error) {
	var link model.Link
	// Unscoped() 让我们能查到被软删除的记录
	// Clauses(clause.Locking{Strength: "UPDATE"}) 会添加 "FOR UPDATE" 实现行锁
	err := tx.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL").
		Order("deleted_at asc").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&link).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有可复用的记录
		}
		return nil, err
	}
	return &link, nil
}

// SaveTx 在事务中保存
func (r *linkRepository) SaveTx(ctx context.Context, tx *gorm.DB, link *model.Link) error {
	return tx.WithContext(ctx).Save(link).Error
}

// FindByOriginalURL 查找一个当前有效的链接
func (r *linkRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) {
	var link model.Link
	now := time.Now()
	// GORM的软删除会自动处理 "deleted_at IS NULL"
	err := r.db.WithContext(ctx).Where("original_url = ? AND (expires_at IS NULL OR expires_at > ?)", originalURL, now).First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到是正常情况，不返回错误
		}
		return nil, err
	}
	return &link, nil
}