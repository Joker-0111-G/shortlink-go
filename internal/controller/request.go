
// 文件路径: internal/controller/request.go
package controller

// CreateShortLinkRequest 定义了创建短链接的请求体
type CreateShortLinkRequest struct {
	URL                 string `json:"url" validate:"required,url"`
	ExpirationInMinutes int    `json:"expiration_in_minutes,omitempty"` // omitempty表示如果值为0，json序列化时会忽略它
}