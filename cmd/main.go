// 文件路径: cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"shortlink-go/internal/controller"
	"shortlink-go/internal/repository"
	"shortlink-go/internal/service"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func main() {
	// --- 1. 初始化数据库 (MySQL) ---
	dsn := viper.GetString("database.mysql.dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	fmt.Println("Database connection successful.")

	// --- 2. 初始化缓存 (Redis) ---
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("cache.redis.addr"),
		Password: viper.GetString("cache.redis.password"),
		DB:       viper.GetInt("cache.redis.db"),
	})
	// 检查Redis连接
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	fmt.Println("Redis connection successful.")

	// --- 3. 初始化 Web 框架 (Echo) ---
	e := echo.New()

	// --- 4. 注册中间件 ---
	e.Use(middleware.Logger())  // 记录日志
	e.Use(middleware.Recover()) // 恢复 panic
	// **重要**: 添加CORS中间件以允许前端访问
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // 生产环境请指定前端域名
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))

	e.Static("/", "frontend") // 将根URL("/")映射到"frontend"目录

	// --- 5. 依赖注入 (从底层到高层) & 注册路由 ---
	linkRepo := repository.NewLinkRepository(db)
	appURL := viper.GetString("server.app_url")
	// 注意：构造函数现在需要传入 *gorm.DB
	linkSvc := service.NewLinkService(linkRepo, rdb, db, appURL) // <--- 修改
	controller.NewLinkController(e, linkSvc)
	fmt.Println("Controller and routes initialized.")

	// --- 启动后台清理任务 ---
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		// 使用 for range 的推荐写法，替代 for+select
		for range ticker.C {
			fmt.Println("Running cleanup job for expired links...")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			
			affected, err := linkSvc.CleanupExpiredLinks(ctx)
			if err != nil {
				fmt.Printf("Error cleaning up expired links: %v\n", err)
			}
			
			if affected > 0 {
				fmt.Printf("Cleaned up %d expired links.\n", affected)
			}
			
			cancel()
		}
	}()

	// --- 6. 启动服务 ---
	serverPort := viper.GetString("server.port")
	fmt.Printf("Starting server on %s\n", serverPort)
	if err := e.Start(serverPort); err != nil {
		e.Logger.Fatal(err)
	}


}