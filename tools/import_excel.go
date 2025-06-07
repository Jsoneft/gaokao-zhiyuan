package main

import (
	"log"
	"os"
	"os/exec"

	"gaokao-zhiyuan/config"
	"gaokao-zhiyuan/database"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 检查Python脚本是否存在
	if _, err := os.Stat("scripts/process_excel.py"); os.IsNotExist(err) {
		log.Fatalf("无法找到数据处理脚本 scripts/process_excel.py: %v", err)
	}

	if _, err := os.Stat("scripts/setup_clickhouse.py"); os.IsNotExist(err) {
		log.Fatalf("无法找到数据库设置脚本 scripts/setup_clickhouse.py: %v", err)
	}

	// 执行ClickHouse设置脚本
	log.Println("正在设置ClickHouse数据库...")
	setupCmd := exec.Command("python3", "scripts/setup_clickhouse.py")
	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr
	if err := setupCmd.Run(); err != nil {
		log.Fatalf("运行数据库设置脚本失败: %v", err)
	}

	// 连接数据库
	db, err := database.NewClickHouseDB(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 验证数据
	count, err := db.GetDataCount()
	if err != nil {
		log.Fatalf("获取数据数量失败: %v", err)
	}

	log.Printf("数据导入完成，共导入 %d 条记录", count)
	log.Println("请运行 make run 启动服务器测试查询功能")
}
