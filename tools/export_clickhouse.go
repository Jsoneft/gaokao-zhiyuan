package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gaokao-zhiyuan/config"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 未找到.env文件或加载失败: %v", err)
	}

	// 加载配置
	cfg := config.LoadConfig()

	// 连接数据库
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouseHost, cfg.ClickHousePort)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDatabase,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePassword,
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

	// 导出表结构
	if err := exportTableSchema(conn); err != nil {
		log.Fatalf("导出表结构失败: %v", err)
	}

	// 导出数据
	if err := exportTableData(conn); err != nil {
		log.Fatalf("导出数据失败: %v", err)
	}

	log.Println("导出完成！")
}

func exportTableSchema(conn driver.Conn) error {
	var createTableSQL string
	row := conn.QueryRow(context.Background(), "SHOW CREATE TABLE gaokao.admission_data")
	if err := row.Scan(&createTableSQL); err != nil {
		return err
	}

	// 创建文件
	file, err := os.Create("setup_clickhouse.sql")
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入数据库创建语句
	file.WriteString("-- 创建数据库\n")
	file.WriteString("CREATE DATABASE IF NOT EXISTS gaokao;\n\n")

	// 写入表创建语句
	file.WriteString("-- 创建表\n")
	file.WriteString(createTableSQL + ";\n\n")

	log.Println("表结构导出成功")
	return nil
}

func exportTableData(conn driver.Conn) error {
	// 查询数据总量
	var count uint64
	row := conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM gaokao.admission_data")
	if err := row.Scan(&count); err != nil {
		return err
	}
	log.Printf("共有 %d 条数据需要导出", count)

	// 打开数据文件
	file, err := os.OpenFile("setup_clickhouse.sql", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 分批查询数据
	batchSize := 10000
	totalBatches := (count + uint64(batchSize) - 1) / uint64(batchSize)
	log.Printf("将分 %d 批导出数据", totalBatches)

	for batchNum := uint64(0); batchNum < totalBatches; batchNum++ {
		offset := batchNum * uint64(batchSize)
		limit := uint64(batchSize)

		log.Printf("正在导出第 %d/%d 批数据...", batchNum+1, totalBatches)

		// 查询一批数据
		rows, err := conn.Query(context.Background(), `
			SELECT id, year, province, batch, subject_type, class_demand, 
			       college_code, special_interest_group_code, college_name, 
			       professional_code, professional_name, lowest_points, 
			       lowest_rank, description
			FROM gaokao.admission_data
			ORDER BY id
			LIMIT ? OFFSET ?
		`, limit, offset)
		if err != nil {
			return err
		}

		// 开始一个批量插入语句
		if batchNum == 0 {
			file.WriteString("-- 插入数据\n")
			file.WriteString("INSERT INTO gaokao.admission_data (id, year, province, batch, subject_type, class_demand, college_code, special_interest_group_code, college_name, professional_code, professional_name, lowest_points, lowest_rank, description) VALUES\n")
		}

		// 处理每一行数据
		rowCount := 0
		for rows.Next() {
			var (
				id                       uint64
				year                     uint16
				province                 string
				batch                    string
				subjectType              string
				classDemand              string
				collegeCode              string
				specialInterestGroupCode string
				collegeName              string
				professionalCode         string
				professionalName         string
				lowestPoints             int64
				lowestRank               int64
				description              string
			)

			if err := rows.Scan(
				&id, &year, &province, &batch, &subjectType, &classDemand,
				&collegeCode, &specialInterestGroupCode, &collegeName,
				&professionalCode, &professionalName, &lowestPoints,
				&lowestRank, &description,
			); err != nil {
				rows.Close()
				return err
			}

			// 转义字符串
			province = escapeString(province)
			batch = escapeString(batch)
			subjectType = escapeString(subjectType)
			classDemand = escapeString(classDemand)
			collegeCode = escapeString(collegeCode)
			specialInterestGroupCode = escapeString(specialInterestGroupCode)
			collegeName = escapeString(collegeName)
			professionalCode = escapeString(professionalCode)
			professionalName = escapeString(professionalName)
			description = escapeString(description)

			// 写入一行数据
			if rowCount > 0 || batchNum > 0 {
				file.WriteString(",\n")
			}
			file.WriteString(fmt.Sprintf("(%d, %d, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %d, %d, '%s')",
				id, year, province, batch, subjectType, classDemand, collegeCode, specialInterestGroupCode,
				collegeName, professionalCode, professionalName, lowestPoints, lowestRank, description))

			rowCount++
		}
		rows.Close()

		// 每批次完成后保存文件
		if batchNum == totalBatches-1 {
			file.WriteString(";\n")
		}

		log.Printf("已导出 %d 条数据", (batchNum+1)*uint64(batchSize))
		time.Sleep(100 * time.Millisecond) // 避免请求过快
	}

	return nil
}

func escapeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "'", "\\'"), "\"", "\\\"")
}
