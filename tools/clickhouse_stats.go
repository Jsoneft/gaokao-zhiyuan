package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// 远程ClickHouse服务器参数
const (
	ClickHouseHost     = "43.248.188.28"
	ClickHousePort     = 26890
	ClickHouseUser     = "default"
	ClickHousePassword = os.Getenv("CLICKHOUSE_PASSWORD")
	ClickHouseDatabase = "gaokao"
)

func main() {
	// 连接数据库
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", ClickHouseHost, ClickHousePort)},
		Auth: clickhouse.Auth{
			Database: ClickHouseDatabase,
			Username: ClickHouseUser,
			Password: ClickHousePassword,
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

	fmt.Println("1. 按省份统计数据条数")
	fmt.Println("=====================")
	printProvinceStats(conn)

	fmt.Println("\n2. 按年份统计数据条数")
	fmt.Println("=====================")
	printYearStats(conn)

	fmt.Println("\n3. 按科类统计数据条数")
	fmt.Println("=====================")
	printSubjectTypeStats(conn)

	fmt.Println("\n4. 湖北省各院校录取数据Top10")
	fmt.Println("============================")
	printTopSchoolsInHubei(conn)
}

func printProvinceStats(conn clickhouse.Conn) {
	// 执行SQL统计
	rows, err := conn.Query(context.Background(), `
		SELECT 
			province, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY province 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
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
	fmt.Printf("%-15s %-10s %-10s\n", "省份", "数据条数", "占比(%)")
	fmt.Println("------------------------------------------")

	// 打印结果
	for _, stat := range stats {
		percentage := float64(stat.Count) / float64(totalCount) * 100
		fmt.Printf("%-15s %-10d %.2f%%\n", stat.Province, stat.Count, percentage)
	}

	// 打印总计
	fmt.Println("------------------------------------------")
	fmt.Printf("%-15s %-10d 100.00%%\n", "总计", totalCount)
}

func printYearStats(conn clickhouse.Conn) {
	// 执行SQL统计
	rows, err := conn.Query(context.Background(), `
		SELECT 
			year, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY year 
		ORDER BY year DESC
	`)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	defer rows.Close()

	// 打印表头
	fmt.Printf("%-10s %-10s\n", "年份", "数据条数")
	fmt.Println("--------------------")

	// 处理查询结果
	var totalCount uint64 = 0
	for rows.Next() {
		var year uint16
		var count uint64

		if err := rows.Scan(&year, &count); err != nil {
			log.Fatalf("读取行数据失败: %v", err)
		}

		fmt.Printf("%-10d %-10d\n", year, count)
		totalCount += count
	}

	// 打印总计
	fmt.Println("--------------------")
	fmt.Printf("%-10s %-10d\n", "总计", totalCount)
}

func printSubjectTypeStats(conn clickhouse.Conn) {
	// 执行SQL统计
	rows, err := conn.Query(context.Background(), `
		SELECT 
			subject_type, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY subject_type 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	defer rows.Close()

	// 打印表头
	fmt.Printf("%-15s %-10s\n", "科类", "数据条数")
	fmt.Println("------------------------")

	// 处理查询结果
	var totalCount uint64 = 0
	for rows.Next() {
		var subjectType string
		var count uint64

		if err := rows.Scan(&subjectType, &count); err != nil {
			log.Fatalf("读取行数据失败: %v", err)
		}

		fmt.Printf("%-15s %-10d\n", subjectType, count)
		totalCount += count
	}

	// 打印总计
	fmt.Println("------------------------")
	fmt.Printf("%-15s %-10d\n", "总计", totalCount)
}

func printTopSchoolsInHubei(conn clickhouse.Conn) {
	// 执行SQL统计
	rows, err := conn.Query(context.Background(), `
		SELECT 
			school_name, 
			COUNT(*) as count 
		FROM admission_data 
		WHERE province = '湖北'
		GROUP BY school_name 
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	defer rows.Close()

	// 打印表头
	fmt.Printf("%-30s %-10s\n", "院校名称", "录取人数")
	fmt.Println("------------------------------------------")

	// 处理查询结果
	for rows.Next() {
		var schoolName string
		var count uint64

		if err := rows.Scan(&schoolName, &count); err != nil {
			log.Fatalf("读取行数据失败: %v", err)
		}

		fmt.Printf("%-30s %-10d\n", schoolName, count)
	}
}
