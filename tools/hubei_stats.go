package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type CollegeStats struct {
	Name      string
	Count     uint64
	AvgPoints float64
	MinPoints float64
	MaxPoints float64
	Subjects  map[string]uint64
}

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

	fmt.Println("湖北省高考录取数据分析")
	fmt.Println("===================")

	// 获取湖北省录取总数和平均分
	var hubeiCount uint64
	var avgPoints float64
	err = conn.QueryRow(context.Background(), `
		SELECT 
			COUNT(*) as count,
			AVG(lowest_points) as avg_points
		FROM admission_data 
		WHERE province = '湖北'
	`).Scan(&hubeiCount, &avgPoints)

	if err != nil {
		log.Printf("查询湖北省总数据失败: %v", err)
	} else {
		fmt.Printf("湖北省录取总数: %d\n", hubeiCount)
		fmt.Printf("湖北省平均录取分数: %.2f\n", avgPoints)
	}

	// 按科类统计湖北省数据
	fmt.Println("\n湖北省各科类录取情况")
	fmt.Println("=================")

	rows1, err := conn.Query(context.Background(), `
		SELECT 
			subject_type, 
			COUNT(*) as count,
			AVG(lowest_points) as avg_points,
			MIN(lowest_points) as min_points,
			MAX(lowest_points) as max_points
		FROM admission_data 
		WHERE province = '湖北'
		GROUP BY subject_type 
		ORDER BY count DESC
	`)

	if err != nil {
		log.Printf("查询湖北省科类统计失败: %v", err)
	} else {
		fmt.Printf("%-10s %-10s %-15s %-15s %-15s\n", "科类", "录取人数", "平均分", "最低分", "最高分")
		fmt.Println("--------------------------------------------------------------------")

		for rows1.Next() {
			var subjectType string
			var count uint64
			var avgPoints, minPoints, maxPoints float64

			if err := rows1.Scan(&subjectType, &count, &avgPoints, &minPoints, &maxPoints); err != nil {
				log.Printf("读取科类统计行失败: %v", err)
				continue
			}

			fmt.Printf("%-10s %-10d %-15.2f %-15.2f %-15.2f\n",
				subjectType, count, avgPoints, minPoints, maxPoints)
		}
		rows1.Close()
	}

	// 湖北省热门院校录取数据
	fmt.Println("\n湖北省热门院校录取数据")
	fmt.Println("=================")

	// 获取院校录取数据
	rows2, err := conn.Query(context.Background(), `
		SELECT 
			college_name, 
			COUNT(*) as count,
			AVG(lowest_points) as avg_points,
			MIN(lowest_points) as min_points,
			MAX(lowest_points) as max_points
		FROM admission_data 
		WHERE province = '湖北'
		GROUP BY college_name 
		ORDER BY count DESC
		LIMIT 20
	`)

	if err != nil {
		log.Printf("查询湖北省院校统计失败: %v", err)
	} else {
		fmt.Printf("%-25s %-10s %-15s %-15s %-15s\n", "院校名称", "录取人数", "平均分", "最低分", "最高分")
		fmt.Println("-------------------------------------------------------------------------")

		var colleges []CollegeStats

		for rows2.Next() {
			var collegeName string
			var count uint64
			var avgPoints, minPoints, maxPoints float64

			if err := rows2.Scan(&collegeName, &count, &avgPoints, &minPoints, &maxPoints); err != nil {
				log.Printf("读取院校统计行失败: %v", err)
				continue
			}

			colleges = append(colleges, CollegeStats{
				Name:      collegeName,
				Count:     count,
				AvgPoints: avgPoints,
				MinPoints: minPoints,
				MaxPoints: maxPoints,
				Subjects:  make(map[string]uint64),
			})

			fmt.Printf("%-25s %-10d %-15.2f %-15.2f %-15.2f\n",
				collegeName, count, avgPoints, minPoints, maxPoints)
		}
		rows2.Close()

		// 获取每个院校的科类分布
		for i, college := range colleges {
			rows3, err := conn.Query(context.Background(), `
				SELECT 
					subject_type, 
					COUNT(*) as count
				FROM admission_data 
				WHERE province = '湖北' AND college_name = ?
				GROUP BY subject_type 
				ORDER BY count DESC
			`, college.Name)

			if err != nil {
				log.Printf("查询院校科类分布失败: %v", err)
				continue
			}

			for rows3.Next() {
				var subjectType string
				var count uint64

				if err := rows3.Scan(&subjectType, &count); err != nil {
					log.Printf("读取院校科类分布行失败: %v", err)
					continue
				}

				colleges[i].Subjects[subjectType] = count
			}
			rows3.Close()
		}

		// 打印院校的科类分布情况
		fmt.Println("\n热门院校科类分布情况 (前5所院校)")
		fmt.Println("==========================")

		for i, college := range colleges {
			if i >= 5 {
				break
			}

			fmt.Printf("\n%s (共%d个专业):\n", college.Name, college.Count)
			fmt.Printf("%-10s %-10s %-10s\n", "科类", "录取人数", "占比(%)")
			fmt.Println("--------------------------------")

			// 将科类排序
			type SubjectCount struct {
				Name  string
				Count uint64
			}

			var subjectCounts []SubjectCount
			for subject, count := range college.Subjects {
				subjectCounts = append(subjectCounts, SubjectCount{subject, count})
			}

			sort.Slice(subjectCounts, func(i, j int) bool {
				return subjectCounts[i].Count > subjectCounts[j].Count
			})

			for _, sc := range subjectCounts {
				percentage := float64(sc.Count) / float64(college.Count) * 100
				fmt.Printf("%-10s %-10d %.2f%%\n", sc.Name, sc.Count, percentage)
			}
		}
	}

	// 湖北省分数段分布
	fmt.Println("\n湖北省录取分数段分布")
	fmt.Println("=================")

	rows4, err := conn.Query(context.Background(), `
		SELECT 
			CAST(FLOOR(lowest_points / 50) * 50 AS String) || '-' || CAST(FLOOR(lowest_points / 50) * 50 + 49 AS String) as score_range,
			COUNT(*) as count
		FROM admission_data 
		WHERE province = '湖北'
		GROUP BY score_range
		ORDER BY MIN(lowest_points)
	`)

	if err != nil {
		log.Printf("查询分数段分布失败: %v", err)
	} else {
		fmt.Printf("%-15s %-10s\n", "分数段", "录取人数")
		fmt.Println("-------------------------")

		for rows4.Next() {
			var scoreRange string
			var count uint64

			if err := rows4.Scan(&scoreRange, &count); err != nil {
				log.Printf("读取分数段分布行失败: %v", err)
				continue
			}

			fmt.Printf("%-15s %-10d\n", scoreRange, count)
		}
		rows4.Close()
	}
}
