// 文件路径: internal/controller/link_controller.go
package controller

import (
	"net/http"
	"shortlink-go/internal/service"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// LinkController 封装了与链接相关的 HTTP handlers
type LinkController struct {
	linkService service.LinkService
}

// NewLinkController 是 LinkController 的构造函数，并在此注册路由
func NewLinkController(e *echo.Echo, linkService service.LinkService) *LinkController {
	controller := &LinkController{
		linkService: linkService,
	}

	// 使用组来管理API版本
	apiGroup := e.Group("/api/v1")
	apiGroup.POST("/shorten", controller.Create)
	apiGroup.GET("/links", controller.GetAll) // <--- 新增这行路由

	// 重定向路由放在根路径
	e.GET("/:shortCode", controller.Redirect)

	return controller
}


// Redirect 处理短链接重定向
func (c *LinkController) Redirect(ctx echo.Context) error {
	shortCode := ctx.Param("shortCode")
	if shortCode == "" {
		return ctx.String(http.StatusBadRequest, "Short code cannot be empty.")
	}

	originalURL, err := c.linkService.GetOriginalURL(ctx.Request().Context(), shortCode)
	if err != nil {
		if err == service.ErrLinkNotFound {
			return ctx.String(http.StatusNotFound, "URL not found.")
		}
		return ctx.String(http.StatusInternalServerError, "Internal server error.")
	}

	return ctx.Redirect(http.StatusMovedPermanently, originalURL)
}


// Create 创建短链接
func (c *LinkController) Create(ctx echo.Context) error {
	var req CreateShortLinkRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	if req.URL == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "URL cannot be empty"})
	}

	shortURL, err := c.linkService.CreateShortLink(ctx.Request().Context(), req.URL, req.ExpirationInMinutes) // <--- 修改
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create short link"})
	}
	return ctx.JSON(http.StatusCreated, map[string]string{"short_url": shortURL})
}

func (c *LinkController) GetAll(ctx echo.Context) error {
	links, err := c.linkService.GetAllLinks(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve links"})
	}
	
	type linkResponse struct {
		ShortURL    string     `json:"short_url"`
		OriginalURL string     `json:"original_url"`
		ExpiresAt   *time.Time `json:"expires_at"` // <--- 新增
	}

	appURL := viper.GetString("server.app_url")
	
	var responses []linkResponse
	for _, link := range links {
		if link.ShortCode != "" {
			responses = append(responses, linkResponse{
				ShortURL:    appURL + link.ShortCode,
				OriginalURL: link.OriginalURL,
				ExpiresAt:   link.ExpiresAt, // <--- 新增
			})
		}
	}

	return ctx.JSON(http.StatusOK, responses)
}