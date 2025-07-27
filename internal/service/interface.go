// 文件路径: internal/service/interface.go

package service

import (
	"context"
	"shortlink-go/internal/model"
)

// LinkService 定义了与链接相关的业务逻辑
// 这里只应该有方法签名，不能有方法体
type LinkService interface {
	CreateShortLink(ctx context.Context, originalURL string, expirationInMinutes int) (string, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetAllLinks(ctx context.Context) ([]model.Link, error)
	CleanupExpiredLinks(ctx context.Context) (int64, error)
}

// CleanupExpiredLinks 调用仓库层来软删除过期的链接
func (s *linkService) CleanupExpiredLinks(ctx context.Context) (int64, error) {
	return s.repo.SoftDeleteExpired(ctx)
}
