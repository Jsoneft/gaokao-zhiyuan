package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// 显示表的前10条数据
	fmt.Println("表数据示例（前5条）")
	fmt.Println("===================")

	dataRows, err := conn.Query(context.Background(), `
		SELECT * FROM admission_data LIMIT 5
	`)
	if err != nil {
		log.Fatalf("查询数据失败: %v", err)
	}
	defer dataRows.Close()

	// 获取列信息
	columnTypes := dataRows.ColumnTypes()
	var columns []string
	for _, ct := range columnTypes {
		columns = append(columns, ct.Name())
	}

	// 打印列名
	fmt.Println("表结构信息")
	fmt.Println("==========")
	fmt.Println("列名列表:")
	for i, col := range columns {
		fmt.Printf("%d. %s\n", i+1, col)
	}

	fmt.Println("\n表数据示例（前5条）")
	fmt.Println("===================")

	// 打印表格头部
	for _, col := range columns {
		fmt.Printf("%-15s", col)
	}
	fmt.Println()

	// 打印分隔线
	for range columns {
		fmt.Printf("%-15s", "-------------")
	}
	fmt.Println()

	// 打印数据
	for dataRows.Next() {
		// 创建与列数相同的接口切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			log.Fatalf("扫描行失败: %v", err)
		}

		// 打印每列的值
		for _, val := range values {
			fmt.Printf("%-15v", val)
		}
		fmt.Println()
	}

	// 尝试获取学校名称相关的列
	fmt.Println("\n学校名称相关列的前10个值")
	fmt.Println("===================")

	// 尝试查询可能的学校名称列
	for _, possibleCol := range []string{"school", "school_name", "university", "university_name", "institution", "institution_name"} {
		// 检查列是否存在
		found := false
		for _, col := range columns {
			if col == possibleCol {
				found = true
				break
			}
		}

		if found {
			schoolRows, err := conn.Query(context.Background(), fmt.Sprintf(`
				SELECT %s, COUNT(*) as count
				FROM admission_data
				WHERE province = '湖北'
				GROUP BY %s
				ORDER BY count DESC
				LIMIT 10
			`, possibleCol, possibleCol))

			if err != nil {
				log.Printf("查询学校列 %s 失败: %v", possibleCol, err)
				continue
			}

			fmt.Printf("使用列 '%s' 统计湖北省的学校录取数据:\n", possibleCol)
			fmt.Printf("%-30s %-10s\n", "学校名称", "录取人数")
			fmt.Println("------------------------------------------")

			for schoolRows.Next() {
				var schoolName string
				var count uint64

				if err := schoolRows.Scan(&schoolName, &count); err != nil {
					log.Printf("读取学校数据失败: %v", err)
					break
				}

				fmt.Printf("%-30s %-10d\n", schoolName, count)
			}

			schoolRows.Close()
		}
	}
}
