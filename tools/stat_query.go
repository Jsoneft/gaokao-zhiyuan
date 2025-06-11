package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	// 连接数据库
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"43.248.188.28:26890"},
		Auth: clickhouse.Auth{
			Database: "gaokao",
			Username: "default",
			Password: "vfdeuiclgb",
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

	// 获取表结构列名
	fmt.Println("表结构信息")
	fmt.Println("==========")

	var columns []string
	err = conn.QueryRow(context.Background(),
		"SELECT column_name FROM system.columns WHERE database = 'gaokao' AND table = 'admission_data'").Scan(&columns)
	if err != nil {
		log.Printf("获取列名失败: %v", err)
	} else {
		fmt.Println("列名列表:")
		for i, col := range columns {
			fmt.Printf("%d. %s\n", i+1, col)
		}
	}

	// 查询学校名称列
	fmt.Println("\n列名是否包含学校名称的查询")
	fmt.Println("======================")

	rows1, err := conn.Query(context.Background(),
		"SELECT column_name FROM system.columns WHERE database = 'gaokao' AND table = 'admission_data' AND column_name LIKE '%college%' OR column_name LIKE '%school%'")
	if err != nil {
		log.Printf("搜索学校列名失败: %v", err)
	} else {
		for rows1.Next() {
			var colName string
			if err := rows1.Scan(&colName); err != nil {
				log.Printf("读取列名失败: %v", err)
				continue
			}
			fmt.Printf("发现可能的学校列名: %s\n", colName)
		}
		rows1.Close()
	}

	// 按省份分组统计
	fmt.Println("\n按省份统计数据条数")
	fmt.Println("==============")

	rows2, err := conn.Query(context.Background(), `
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

	// 打印表头
	fmt.Printf("%-15s %-10s\n", "省份", "数据条数")
	fmt.Println("---------------------------")

	// 处理查询结果
	var totalCount uint64 = 0
	for rows2.Next() {
		var province string
		var count uint64

		if err := rows2.Scan(&province, &count); err != nil {
			log.Fatalf("读取省份统计行失败: %v", err)
		}

		fmt.Printf("%-15s %-10d\n", province, count)
		totalCount += count
	}
	rows2.Close()

	// 打印总计
	fmt.Println("---------------------------")
	fmt.Printf("%-15s %-10d\n", "总计", totalCount)

	// 按年份统计
	fmt.Println("\n按年份统计数据条数")
	fmt.Println("==============")

	rows3, err := conn.Query(context.Background(), `
		SELECT 
			year, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY year 
		ORDER BY year DESC
	`)
	if err != nil {
		log.Fatalf("查询按年份统计失败: %v", err)
	}

	// 打印表头
	fmt.Printf("%-10s %-10s\n", "年份", "数据条数")
	fmt.Println("--------------------")

	// 处理查询结果
	for rows3.Next() {
		var year uint16
		var count uint64

		if err := rows3.Scan(&year, &count); err != nil {
			log.Fatalf("读取年份统计行失败: %v", err)
		}

		fmt.Printf("%-10d %-10d\n", year, count)
	}
	rows3.Close()

	// 按科类统计
	fmt.Println("\n按科类统计数据条数")
	fmt.Println("==============")

	rows4, err := conn.Query(context.Background(), `
		SELECT 
			subject_type, 
			COUNT(*) as count 
		FROM admission_data 
		GROUP BY subject_type 
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatalf("查询按科类统计失败: %v", err)
	}

	// 打印表头
	fmt.Printf("%-15s %-10s\n", "科类", "数据条数")
	fmt.Println("---------------------------")

	// 处理查询结果
	for rows4.Next() {
		var subjectType string
		var count uint64

		if err := rows4.Scan(&subjectType, &count); err != nil {
			log.Fatalf("读取科类统计行失败: %v", err)
		}

		fmt.Printf("%-15s %-10d\n", subjectType, count)
	}
	rows4.Close()

	// 湖北省院校数据
	fmt.Println("\n湖北省各院校录取数据Top10")
	fmt.Println("===================")

	rows5, err := conn.Query(context.Background(), `
		SELECT 
			college_name, 
			COUNT(*) as count 
		FROM admission_data 
		WHERE province = '湖北'
		GROUP BY college_name 
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		log.Printf("查询湖北省院校统计失败: %v", err)
	} else {
		// 打印表头
		fmt.Printf("%-30s %-10s\n", "院校名称", "录取人数")
		fmt.Println("------------------------------------------")

		// 处理查询结果
		for rows5.Next() {
			var collegeName string
			var count uint64

			if err := rows5.Scan(&collegeName, &count); err != nil {
				log.Printf("读取院校统计行失败: %v", err)
				continue
			}

			fmt.Printf("%-30s %-10d\n", collegeName, count)
		}
		rows5.Close()
	}
}
