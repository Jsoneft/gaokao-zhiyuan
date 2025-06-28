package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	// 连接数据库
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"43.248.188.28:26890"},
		Auth: clickhouse.Auth{
			Database: "gaokao",
			Username: "default",
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
	})
	if err != nil {
		log.Fatalf("连接ClickHouse失败: %v", err)
	}
	defer conn.Close()

	// 检查连接
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Ping失败: %v", err)
	}

	// 按省份(生源地)分组统计
	fmt.Println("按生源地(province)统计数据分布")
	fmt.Println("==========================")

	rows, err := conn.Query(context.Background(), `
		SELECT 
			province, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY province 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatalf("查询按省份统计失败: %v", err)
	}
	defer rows.Close()

	// 创建结果存储结构
	type ProvinceStats struct {
		Province string
		Count    uint64
	}

	var stats []ProvinceStats
	var totalCount uint64 = 0

	// 处理查询结果
	for rows.Next() {
		var province string
		var count uint64

		if err := rows.Scan(&province, &count); err != nil {
			log.Fatalf("读取行数据失败: %v", err)
		}

		stats = append(stats, ProvinceStats{
			Province: province,
			Count:    count,
		})
		totalCount += count
	}

	// 排序结果（按数量降序）
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// 打印表头
	fmt.Printf("%-15s %-10s %-10s %-15s\n", "省份(生源地)", "数据条数", "占比(%)", "占比(条形图)")
	fmt.Println("-----------------------------------------------------------------------")

	// 打印结果
	for _, stat := range stats {
		percentage := float64(stat.Count) / float64(totalCount) * 100
		barLength := int(percentage / 2) // 每2%显示一个字符
		bar := ""
		for i := 0; i < barLength; i++ {
			bar += "█"
		}
		fmt.Printf("%-15s %-10d %-10.2f%% %-15s\n", stat.Province, stat.Count, percentage, bar)
	}

	// 打印总计
	fmt.Println("-----------------------------------------------------------------------")
	fmt.Printf("%-15s %-10d %-10s\n", "总计", totalCount, "100.00%")
}
