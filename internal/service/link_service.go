// 文件路径: internal/service/link_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"shortlink-go/internal/model"
	"shortlink-go/internal/repository"
	"shortlink-go/pkg/util"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrLinkNotFound = errors.New("link not found")

type linkService struct {
	repo   repository.LinkRepository
	rdb    *redis.Client
	appURL string
	db     *gorm.DB // <--- 新增db实例，用于事务
}

// GetOriginalURL implements LinkService.
func (s *linkService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	panic("unimplemented")
}

// 文件路径: internal/service/link_service.go

// NewLinkService 构造函数需要传入 gorm.DB
func NewLinkService(repo repository.LinkRepository, rdb *redis.Client, db *gorm.DB, appURL string) LinkService {
	return &linkService{
		repo:   repo,
		rdb:    rdb,
		db:     db,
		appURL: appURL,
	}
}


// CreateShortLink 实现了“存在则刷新，不存在则创建/复用”的逻辑
func (s *linkService) CreateShortLink(ctx context.Context, originalURL string, expirationInMinutes int) (string, error) {
	// 1. 优先根据原始URL查找有效的、未过期的链接
	existingLink, err := s.repo.FindByOriginalURL(ctx, originalURL)
	if err != nil {
		return "", fmt.Errorf("error checking for existing link: %w", err)
	}

	// 2. 如果找到了，就更新它的过期时间并直接返回
	if existingLink != nil {
		fmt.Println("Found existing active link, refreshing expiration...")
		// 处理有效期刷新
		if expirationInMinutes == -1 {
			existingLink.ExpiresAt = nil // 设为永久
		} else {
			if expirationInMinutes <= 0 {
				expirationInMinutes = 60 // 默认60分钟
			}
			newExpiresAt := time.Now().Add(time.Duration(expirationInMinutes) * time.Minute)
			existingLink.ExpiresAt = &newExpiresAt
		}

		// 保存更新
		if err := s.repo.Save(ctx, existingLink); err != nil {
			return "", fmt.Errorf("failed to update expiration for existing link: %w", err)
		}

		// 更新缓存
		cacheKey := "shortlink:" + existingLink.ShortCode
		cacheDuration := 24 * time.Hour
		if existingLink.ExpiresAt != nil {
			cacheDuration = time.Until(*existingLink.ExpiresAt)
		}
		if cacheDuration > 0 {
			s.rdb.Set(ctx, cacheKey, existingLink.OriginalURL, cacheDuration)
		}

		return s.appURL + existingLink.ShortCode, nil
	}

	var finalLink *model.Link

	tx := s.db.Begin()
	// 开始事务
	if tx.Error != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	reusableLink, err := s.repo.FindAndLockOldestReusable(ctx, tx)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to find reusable link: %w", err)
	}

	if reusableLink != nil {
		// --- 找到了，复用它 ---
		reusableLink.OriginalURL = originalURL
		reusableLink.CreatedAt = time.Now()
		reusableLink.DeletedAt = gorm.DeletedAt{}
		if expirationInMinutes == -1 {
			reusableLink.ExpiresAt = nil
		} else {
			if expirationInMinutes <= 0 {
				expirationInMinutes = 60
			}
			expiresAt := time.Now().Add(time.Duration(expirationInMinutes) * time.Minute)
			reusableLink.ExpiresAt = &expiresAt
		}
		if err := s.repo.SaveTx(ctx, tx, reusableLink); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to resurrect link: %w", err)
		}
		finalLink = reusableLink
	} else {
		// --- 没找到，创建新的 ---
		newLink := &model.Link{
			OriginalURL: originalURL,
		}
		if expirationInMinutes == -1 {
			newLink.ExpiresAt = nil
		} else {
			if expirationInMinutes <= 0 {
				expirationInMinutes = 60
			}
			expiresAt := time.Now().Add(time.Duration(expirationInMinutes) * time.Minute)
			newLink.ExpiresAt = &expiresAt
		}
		if err := s.repo.SaveTx(ctx, tx, newLink); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to save new link: %w", err)
		}
		newLink.ShortCode = util.ToBase62(newLink.ID)
		if err := s.repo.SaveTx(ctx, tx, newLink); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to update new link with short code: %w", err)
		}
		finalLink = newLink
	}

	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	cacheKey := "shortlink:" + finalLink.ShortCode
	cacheDuration := 24 * time.Hour
	if finalLink.ExpiresAt != nil {
		cacheDuration = time.Until(*finalLink.ExpiresAt)
	}
	if cacheDuration > 0 {
		s.rdb.Set(ctx, cacheKey, finalLink.OriginalURL, cacheDuration)
	}

	return s.appURL + finalLink.ShortCode, nil
}

func (s *linkService) GetAllLinks(ctx context.Context) ([]model.Link, error) {
	return s.repo.FindAll(ctx)
}
