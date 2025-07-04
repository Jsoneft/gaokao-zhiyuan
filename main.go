package main

import (
	"log"
	"net/http"

	"gaokao-zhiyuan/config"
	"gaokao-zhiyuan/database"
	"gaokao-zhiyuan/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 未找到.env文件或加载失败: %v", err)
	}

	// 加载配置
	cfg := config.LoadConfig()

	// 输出连接信息
	log.Printf("使用ClickHouse连接: %s:%d, 用户: %s, 数据库: %s",
		cfg.ClickHouseHost, cfg.ClickHousePort, cfg.ClickHouseUser, cfg.ClickHouseDatabase)

	// 设置Gin模式
	gin.SetMode(cfg.GinMode)

	// 连接数据库
	db, err := database.NewClickHouseDB(cfg)
	if err != nil {
		log.Fatalf("连接ClickHouse失败: %v", err)
	}
	defer db.Close()

	// 创建表（如果不存在）
	if err := db.CreateTable(); err != nil {
		log.Fatalf("创建表失败: %v", err)
	}

	// 创建处理器
	handler := handlers.NewHandler(db)

	// 创建路由
	router := setupRouter(handler)

	// 启动服务器
	log.Printf("服务器启动在端口 %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

func setupRouter(handler *handlers.Handler) *gin.Engine {
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API路由
	api := router.Group("/api")
	{
		// 健康检查
		api.GET("/health", handler.HealthCheck)

		// 位次查询接口
		api.GET("/rank/get", handler.GetRank)

		// 报表查询接口
		api.GET("/report/get", handler.GetReport)
	}

	// V1 API路由 - 新增接口
	v1 := router.Group("/api/v1")
	{
		// 高级查询位次接口
		v1.POST("/query_rank", handler.QueryRank)
	}

	return router
}
