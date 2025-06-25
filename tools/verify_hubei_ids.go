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

type HubeiRecord struct {
	ID                int64   `json:"id"`
	MajorMinScore2024 float64 `json:"major_min_score_2024"`
	CollegeNameExcel  string  `json:"college_name_excel"`
	MajorNameExcel    string  `json:"major_name_excel"`
	SourceProvince    string  `json:"source_province"`
}

type ClickHouseRecord struct {
	ID           int64   `json:"id"`
	SchoolName   string  `json:"school_name"`
	MajorName    string  `json:"major_name"`
	MinScore2024 *uint16 `json:"min_score_2024"`
	MinRank2024  *uint32 `json:"min_rank_2024"`
}

func main() {
	fmt.Println("验证湖北生源地ID映射关系...")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 读取CSV数据
	fmt.Println("1. 读取湖北分数更新数据...")
	hubeiRecords, err := readHubeiCSV("hubei_score_update_data.csv")
	if err != nil {
		log.Fatalf("读取CSV文件失败: %v", err)
	}
	fmt.Printf("湖北记录总数: %d\n", len(hubeiRecords))

	// 取样本进行验证
	sampleSize := 100
	if len(hubeiRecords) < sampleSize {
		sampleSize = len(hubeiRecords)
	}
	sampleRecords := hubeiRecords[:sampleSize]

	// 提取样本ID
	sampleIDs := make([]int64, len(sampleRecords))
	for i, record := range sampleRecords {
		sampleIDs[i] = record.ID
	}
	fmt.Printf("样本ID数量: %d\n", len(sampleIDs))

	// 2. 连接ClickHouse
	fmt.Println("\n2. 连接ClickHouse...")
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"43.248.188.28:42914"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		log.Fatalf("连接ClickHouse失败: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 3. 检查表结构
	fmt.Println("\n3. 检查目标表结构...")
	structureQuery := "DESCRIBE TABLE default.admission_hubei_wide_2024"
	rows, err := conn.Query(ctx, structureQuery)
	if err != nil {
		log.Fatalf("查询表结构失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("表结构:")
	fieldCount := 0
	for rows.Next() {
		var name, dataType, defaultType, defaultExpression, comment, codecExpression, ttlExpression string
		if err := rows.Scan(&name, &dataType, &defaultType, &defaultExpression, &comment, &codecExpression, &ttlExpression); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		if fieldCount < 10 { // 只显示前10个字段
			fmt.Printf("  %-25s %s\n", name, dataType)
		}
		fieldCount++
	}
	if fieldCount > 10 {
		fmt.Printf("  ... 还有 %d 个字段\n", fieldCount-10)
	}

	// 4. 查询匹配的ID
	fmt.Println("\n4. 查询ClickHouse中的匹配ID...")
	idStrings := make([]string, len(sampleIDs))
	for i, id := range sampleIDs {
		idStrings[i] = strconv.FormatInt(id, 10)
	}
	idList := strings.Join(idStrings, ",")

	query := fmt.Sprintf(`
		SELECT id, school_name, major_name, min_score_2024, min_rank_2024
		FROM default.admission_hubei_wide_2024 
		WHERE id IN (%s)
		ORDER BY id
		LIMIT 100
	`, idList)

	rows, err = conn.Query(ctx, query)
	if err != nil {
		log.Fatalf("查询匹配ID失败: %v", err)
	}
	defer rows.Close()

	var chRecords []ClickHouseRecord
	for rows.Next() {
		var record ClickHouseRecord
		if err := rows.Scan(&record.ID, &record.SchoolName, &record.MajorName, &record.MinScore2024, &record.MinRank2024); err != nil {
			log.Printf("扫描记录失败: %v", err)
			continue
		}
		chRecords = append(chRecords, record)
	}

	fmt.Printf("ClickHouse中匹配的记录数: %d\n", len(chRecords))

	if len(chRecords) > 0 {
		fmt.Println("\n前10条匹配记录:")
		for i, record := range chRecords {
			if i >= 10 {
				break
			}
			scoreStr := "null"
			if record.MinScore2024 != nil {
				scoreStr = fmt.Sprintf("%d", *record.MinScore2024)
			}
			fmt.Printf("  %2d. ID: %d, 院校: %s, 专业: %s, 分数: %s\n",
				i+1, record.ID, record.SchoolName, record.MajorName, scoreStr)
		}
	}

	// 5. 对比分析
	fmt.Println("\n5. 对比分析...")

	// 创建映射
	chIDMap := make(map[int64]ClickHouseRecord)
	for _, record := range chRecords {
		chIDMap[record.ID] = record
	}

	hubeiIDMap := make(map[int64]HubeiRecord)
	for _, record := range sampleRecords {
		hubeiIDMap[record.ID] = record
	}

	// 计算匹配情况
	var matchedIDs []int64
	for id := range chIDMap {
		if _, exists := hubeiIDMap[id]; exists {
			matchedIDs = append(matchedIDs, id)
		}
	}

	fmt.Printf("Excel湖北样本ID数: %d\n", len(sampleRecords))
	fmt.Printf("ClickHouse匹配ID数: %d\n", len(chRecords))
	fmt.Printf("成功匹配的ID数: %d\n", len(matchedIDs))
	fmt.Printf("Excel中未匹配的ID数: %d\n", len(sampleRecords)-len(matchedIDs))

	matchRate := 0.0
	if len(sampleRecords) > 0 {
		matchRate = float64(len(matchedIDs)) / float64(len(sampleRecords)) * 100
	}
	fmt.Printf("匹配率: %.1f%%\n", matchRate)

	// 6. 详细对比匹配的记录
	if len(matchedIDs) > 0 {
		fmt.Println("\n6. 详细对比前10条匹配记录...")

		for i, id := range matchedIDs {
			if i >= 10 {
				break
			}

			hubeiRecord := hubeiIDMap[id]
			chRecord := chIDMap[id]

			fmt.Printf("\n%d. ID: %d\n", i+1, id)
			fmt.Printf("   生源地:   %s\n", hubeiRecord.SourceProvince)
			fmt.Printf("   Excel院校: %s\n", hubeiRecord.CollegeNameExcel)
			fmt.Printf("   CH院校:   %s\n", chRecord.SchoolName)
			fmt.Printf("   Excel专业: %s\n", hubeiRecord.MajorNameExcel)
			fmt.Printf("   CH专业:   %s\n", chRecord.MajorName)
			fmt.Printf("   Excel分数: %.0f\n", hubeiRecord.MajorMinScore2024)

			chScoreStr := "null"
			if chRecord.MinScore2024 != nil {
				chScoreStr = fmt.Sprintf("%d", *chRecord.MinScore2024)
			}
			fmt.Printf("   CH分数:   %s\n", chScoreStr)
		}
	}

	// 7. 检查重复ID（笛卡尔积检查）
	fmt.Println("\n7. 笛卡尔积检查...")

	duplicateQuery := fmt.Sprintf(`
		SELECT id, COUNT(*) as count
		FROM default.admission_hubei_wide_2024 
		WHERE id IN (%s)
		GROUP BY id
		HAVING count > 1
		ORDER BY count DESC
	`, idList)

	rows, err = conn.Query(ctx, duplicateQuery)
	if err != nil {
		log.Printf("重复检查查询失败: %v", err)
	} else {
		defer rows.Close()

		var duplicateRecords []struct {
			ID    int64
			Count uint64
		}

		for rows.Next() {
			var id int64
			var count uint64
			if err := rows.Scan(&id, &count); err != nil {
				log.Printf("扫描重复记录失败: %v", err)
				continue
			}
			duplicateRecords = append(duplicateRecords, struct {
				ID    int64
				Count uint64
			}{id, count})
		}

		if len(duplicateRecords) > 0 {
			fmt.Printf("发现重复ID数量: %d\n", len(duplicateRecords))
			fmt.Println("重复ID详情:")
			for i, record := range duplicateRecords {
				if i >= 5 {
					break
				}
				fmt.Printf("  ID: %d, 重复次数: %d\n", record.ID, record.Count)
			}
		} else {
			fmt.Println("✅ 没有发现重复ID，无笛卡尔积问题")
		}
	}

	// 8. 总结
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("验证总结:")
	fmt.Printf("✅ Excel湖北数据: %d 条记录\n", len(hubeiRecords))
	fmt.Printf("✅ ID匹配率: %.1f%% (%d/%d)\n", matchRate, len(matchedIDs), len(sampleRecords))

	if len(matchedIDs) > 0 && matchRate > 50 {
		fmt.Println("✅ 匹配率可接受，可以进行数据更新")

		// 保存验证结果用于后续更新
		fmt.Println("\n正在保存验证后的完整湖北数据...")
		hubeiScoreFile := "hubei_score_update_verified.csv"
		updateData := make([][]string, len(hubeiRecords)+1)
		updateData[0] = []string{"id", "major_min_score_2024", "college_name", "major_name", "source_province"}

		for i, record := range hubeiRecords {
			updateData[i+1] = []string{
				strconv.FormatInt(record.ID, 10),
				fmt.Sprintf("%.0f", record.MajorMinScore2024),
				record.CollegeNameExcel,
				record.MajorNameExcel,
				record.SourceProvince,
			}
		}

		file, err := os.Create(hubeiScoreFile)
		if err != nil {
			log.Printf("创建验证文件失败: %v", err)
		} else {
			defer file.Close()
			writer := csv.NewWriter(file)
			defer writer.Flush()

			for _, record := range updateData {
				if err := writer.Write(record); err != nil {
					log.Printf("写入记录失败: %v", err)
				}
			}
			fmt.Printf("✅ 验证后的更新数据已保存到: %s\n", hubeiScoreFile)
		}
	} else {
		fmt.Printf("❌ 匹配率过低（%.1f%%），需要检查数据源\n", matchRate)
	}
}

func readHubeiCSV(filename string) ([]HubeiRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV文件为空")
	}

	// 跳过标题行
	var hubeiRecords []HubeiRecord
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 5 {
			continue
		}

		id, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue
		}

		score, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		hubeiRecords = append(hubeiRecords, HubeiRecord{
			ID:                id,
			MajorMinScore2024: score,
			CollegeNameExcel:  record[2],
			MajorNameExcel:    record[3],
			SourceProvince:    record[4],
		})
	}

	return hubeiRecords, nil
}
