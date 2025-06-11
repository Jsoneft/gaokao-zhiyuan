package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// 省份统计数据结构
type ProvinceData struct {
	Name         string
	Count        uint64
	AvgPoints    float64
	AvgRank      float64
	PhysicsCount uint64
	HistoryCount uint64
	CollegeCount uint64
	TopColleges  []string
	ScoreRanges  map[string]uint64
}

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

	// 获取所有省份列表
	rows, err := conn.Query(context.Background(), `
		SELECT 
			province,
			COUNT(*) as count,
			AVG(lowest_points) as avg_points,
			AVG(lowest_rank) as avg_rank
		FROM admission_data 
		GROUP BY province
		ORDER BY count DESC
	`)
	if err != nil {
		log.Fatalf("查询省份数据失败: %v", err)
	}
	defer rows.Close()

	// 收集省份数据
	var provinces []ProvinceData
	for rows.Next() {
		var province string
		var count uint64
		var avgPoints, avgRank float64

		if err := rows.Scan(&province, &count, &avgPoints, &avgRank); err != nil {
			log.Printf("读取省份数据失败: %v", err)
			continue
		}

		provinces = append(provinces, ProvinceData{
			Name:        province,
			Count:       count,
			AvgPoints:   avgPoints,
			AvgRank:     avgRank,
			ScoreRanges: make(map[string]uint64),
		})
	}

	// 获取每个省份的科类分布和院校数量
	for i, province := range provinces {
		// 科类分布
		rows1, err := conn.Query(context.Background(), `
			SELECT 
				subject_type,
				COUNT(*) as count
			FROM admission_data 
			WHERE province = ?
			GROUP BY subject_type
		`, province.Name)

		if err != nil {
			log.Printf("查询%s科类分布失败: %v", province.Name, err)
		} else {
			for rows1.Next() {
				var subjectType string
				var count uint64

				if err := rows1.Scan(&subjectType, &count); err != nil {
					continue
				}

				if subjectType == "物理" {
					provinces[i].PhysicsCount = count
				} else if subjectType == "历史" {
					provinces[i].HistoryCount = count
				}
			}
			rows1.Close()
		}

		// 院校数量
		err = conn.QueryRow(context.Background(), `
			SELECT 
				COUNT(DISTINCT college_name) as college_count
			FROM admission_data 
			WHERE province = ?
		`, province.Name).Scan(&provinces[i].CollegeCount)

		if err != nil {
			log.Printf("查询%s院校数量失败: %v", province.Name, err)
		}

		// 热门院校
		rows2, err := conn.Query(context.Background(), `
			SELECT 
				college_name,
				COUNT(*) as count
			FROM admission_data 
			WHERE province = ?
			GROUP BY college_name
			ORDER BY count DESC
			LIMIT 3
		`, province.Name)

		if err != nil {
			log.Printf("查询%s热门院校失败: %v", province.Name, err)
		} else {
			for rows2.Next() {
				var collegeName string
				var count uint64

				if err := rows2.Scan(&collegeName, &count); err != nil {
					continue
				}

				provinces[i].TopColleges = append(provinces[i].TopColleges, collegeName)
			}
			rows2.Close()
		}

		// 分数段分布
		rows3, err := conn.Query(context.Background(), `
			SELECT 
				CAST(FLOOR(lowest_points / 50) * 50 AS String) || '-' || CAST(FLOOR(lowest_points / 50) * 50 + 49 AS String) as score_range,
				COUNT(*) as count
			FROM admission_data 
			WHERE province = ?
			GROUP BY score_range
			ORDER BY MIN(lowest_points)
		`, province.Name)

		if err != nil {
			log.Printf("查询%s分数段分布失败: %v", province.Name, err)
		} else {
			for rows3.Next() {
				var scoreRange string
				var count uint64

				if err := rows3.Scan(&scoreRange, &count); err != nil {
					continue
				}

				provinces[i].ScoreRanges[scoreRange] = count
			}
			rows3.Close()
		}
	}

	// 打印省份数据
	fmt.Println("各省份录取数据对比")
	fmt.Println("==============")
	fmt.Printf("%-8s %-10s %-12s %-12s %-12s %-12s %-12s\n",
		"省份", "录取总数", "平均分", "平均位次", "物理人数", "历史人数", "院校数量")
	fmt.Println("---------------------------------------------------------------------------")

	for _, province := range provinces {
		fmt.Printf("%-8s %-10d %-12.2f %-12.2f %-12d %-12d %-12d\n",
			province.Name, province.Count, province.AvgPoints, province.AvgRank,
			province.PhysicsCount, province.HistoryCount, province.CollegeCount)
	}

	// 湖北省与周边省份对比
	fmt.Println("\n湖北省与周边省份对比")
	fmt.Println("================")

	// 寻找湖北省数据
	var hubei ProvinceData
	var centralProvinces []ProvinceData
	centralProvinceNames := []string{"湖北", "湖南", "河南", "安徽", "江西", "陕西", "重庆"}

	for _, province := range provinces {
		for _, name := range centralProvinceNames {
			if province.Name == name {
				centralProvinces = append(centralProvinces, province)
				if name == "湖北" {
					hubei = province
				}
				break
			}
		}
	}

	// 按录取总数排序
	sort.Slice(centralProvinces, func(i, j int) bool {
		return centralProvinces[i].Count > centralProvinces[j].Count
	})

	fmt.Printf("%-8s %-10s %-12s %-12s %-12s %-12s %-12s\n",
		"省份", "录取总数", "平均分", "平均位次", "物理人数", "历史人数", "院校数量")
	fmt.Println("---------------------------------------------------------------------------")

	for _, province := range centralProvinces {
		fmt.Printf("%-8s %-10d %-12.2f %-12.2f %-12d %-12d %-12d\n",
			province.Name, province.Count, province.AvgPoints, province.AvgRank,
			province.PhysicsCount, province.HistoryCount, province.CollegeCount)
	}

	// 湖北省热门院校
	fmt.Println("\n湖北省热门院校")
	fmt.Println("===========")
	for _, college := range hubei.TopColleges {
		fmt.Println(college)
	}
}
