package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	fmt.Println("简化ID匹配验证...")
	fmt.Println(strings.Repeat("=", 40))

	// 1. 读取CSV中的ID样本
	fmt.Println("1. 读取湖北数据ID...")
	hubeiFile := "hubei_score_update_data.csv"
	file, err := os.Open(hubeiFile)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取CSV失败: %v", err)
	}

	// 提取前10个ID
	var excelIDs []uint32
	for i := 1; i <= 10 && i < len(records); i++ {
		if len(records[i]) > 0 {
			id, err := strconv.ParseUint(records[i][0], 10, 32)
			if err == nil {
				excelIDs = append(excelIDs, uint32(id))
			}
		}
	}

	fmt.Printf("Excel前10个ID: %v\n", excelIDs)

	// 2. 连接ClickHouse查看ID范围
	fmt.Println("\n2. 连接ClickHouse查看ID范围...")
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"43.248.188.28:42914"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 查询ClickHouse中的ID范围
	rangeQuery := "SELECT MIN(id), MAX(id), COUNT(*) FROM default.admission_hubei_wide_2024"
	rows, err := conn.Query(ctx, rangeQuery)
	if err != nil {
		log.Fatalf("查询ID范围失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var minID, maxID, count uint32
		if err := rows.Scan(&minID, &maxID, &count); err != nil {
			log.Printf("扫描失败: %v", err)
			continue
		}
		fmt.Printf("ClickHouse ID范围: %d - %d, 总数: %d\n", minID, maxID, count)
	}

	// 查询前10个ID
	topQuery := "SELECT id, school_name, major_name FROM default.admission_hubei_wide_2024 ORDER BY id LIMIT 10"
	rows, err = conn.Query(ctx, topQuery)
	if err != nil {
		log.Fatalf("查询前10个ID失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nClickHouse前10个ID:")
	for rows.Next() {
		var id uint32
		var schoolName, majorName string
		if err := rows.Scan(&id, &schoolName, &majorName); err != nil {
			log.Printf("扫描失败: %v", err)
			continue
		}
		fmt.Printf("  ID: %d, 学校: %s, 专业: %s\n", id, schoolName, majorName)
	}

	// 直接测试几个Excel ID是否存在
	if len(excelIDs) > 0 {
		fmt.Println("\n3. 测试Excel ID是否在ClickHouse中存在...")
		testIDs := excelIDs[:min(5, len(excelIDs))]
		idStrings := make([]string, len(testIDs))
		for i, id := range testIDs {
			idStrings[i] = fmt.Sprintf("%d", id)
		}
		idList := strings.Join(idStrings, ",")

		testQuery := fmt.Sprintf("SELECT id, school_name FROM default.admission_hubei_wide_2024 WHERE id IN (%s)", idList)
		rows, err = conn.Query(ctx, testQuery)
		if err != nil {
			log.Printf("测试查询失败: %v", err)
		} else {
			defer rows.Close()
			found := 0
			for rows.Next() {
				var id uint32
				var schoolName string
				if err := rows.Scan(&id, &schoolName); err != nil {
					log.Printf("扫描失败: %v", err)
					continue
				}
				fmt.Printf("  找到匹配: ID %d, 学校: %s\n", id, schoolName)
				found++
			}
			fmt.Printf("测试结果: 在%d个Excel ID中找到%d个匹配\n", len(testIDs), found)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
