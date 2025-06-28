package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	// 从环境变量或默认值获取连接信息
	host := getEnv("CLICKHOUSE_HOST", "43.248.188.28")
	port := getEnv("CLICKHOUSE_PORT", "42914")
	user := getEnv("CLICKHOUSE_USERNAME", "default")
	password := getEnv("CLICKHOUSE_PASSWORD", "")
	database := getEnv("CLICKHOUSE_DATABASE", "default")

	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s", user, password, host, port, database)
	fmt.Printf("测试ClickHouse连接: %s:%s, 用户: %s, 数据库: %s\n", host, port, user, database)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Ping失败: %v", err)
	}
	fmt.Println("✅ ClickHouse连接成功")

	// 测试查询版本
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("查询版本失败: %v", err)
	}
	fmt.Printf("✅ ClickHouse版本: %s\n", version)

	// 检查目标表
	var exists int
	err = db.QueryRow("SELECT count() FROM system.tables WHERE database = 'default' AND name = 'admission_hubei_wide_2024'").Scan(&exists)
	if err != nil {
		log.Printf("❌ 检查表失败: %v", err)
	} else if exists > 0 {
		fmt.Println("✅ 表 default.admission_hubei_wide_2024 存在")
		
		// 检查记录数
		var count int64
		err = db.QueryRow("SELECT COUNT(*) FROM default.admission_hubei_wide_2024").Scan(&count)
		if err != nil {
			log.Printf("❌ 查询记录数失败: %v", err)
		} else {
			fmt.Printf("✅ 表记录数: %d\n", count)
		}
		
		// 检查major_min_score_2024字段
		var fieldExists int
		err = db.QueryRow("SELECT count() FROM system.columns WHERE database = 'default' AND table = 'admission_hubei_wide_2024' AND name = 'major_min_score_2024'").Scan(&fieldExists)
		if err != nil {
			log.Printf("❌ 检查字段失败: %v", err)
		} else if fieldExists > 0 {
			fmt.Println("✅ major_min_score_2024 字段存在")
		} else {
			fmt.Println("❌ major_min_score_2024 字段不存在")
		}
	} else {
		fmt.Println("❌ 表 default.admission_hubei_wide_2024 不存在")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
