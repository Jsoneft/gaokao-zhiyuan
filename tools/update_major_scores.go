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

type UpdateRecord struct {
	ID                uint32  `json:"id"`
	MajorMinScore2024 float64 `json:"major_min_score_2024"`
	CollegeName       string  `json:"college_name"`
	MajorName         string  `json:"major_name"`
	SourceProvince    string  `json:"source_province"`
}

func main() {
	fmt.Println("更新专业最低分数据...")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 读取湖北分数数据
	fmt.Println("1. 读取湖北专业最低分数据...")
	records, err := readUpdateCSV("hubei_score_update_data.csv")
	if err != nil {
		log.Fatalf("读取CSV失败: %v", err)
	}
	fmt.Printf("读取到 %d 条湖北分数记录\n", len(records))

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
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 3. 验证样本匹配
	fmt.Println("\n3. 验证ID匹配...")
	sampleSize := 100
	if len(records) < sampleSize {
		sampleSize = len(records)
	}
	sampleRecords := records[:sampleSize]

	sampleIDs := make([]string, len(sampleRecords))
	for i, record := range sampleRecords {
		sampleIDs[i] = fmt.Sprintf("%d", record.ID)
	}
	idList := strings.Join(sampleIDs, ",")

	verifyQuery := fmt.Sprintf(`
		SELECT COUNT(*) as matched_count
		FROM default.admission_hubei_wide_2024 
		WHERE id IN (%s)
	`, idList)

	rows, err := conn.Query(ctx, verifyQuery)
	if err != nil {
		log.Fatalf("验证查询失败: %v", err)
	}
	defer rows.Close()

	var matchedCount uint64
	for rows.Next() {
		if err := rows.Scan(&matchedCount); err != nil {
			log.Fatalf("扫描失败: %v", err)
		}
	}

	matchRate := float64(matchedCount) / float64(len(sampleRecords)) * 100
	fmt.Printf("样本匹配率: %.1f%% (%d/%d)\n", matchRate, matchedCount, len(sampleRecords))

	if matchRate < 80 {
		log.Fatalf("匹配率过低 (%.1f%%)，停止更新", matchRate)
	}

	// 4. 检查是否需要添加字段
	fmt.Println("\n4. 检查专业最低分字段...")
	checkQuery := `
		SELECT name 
		FROM system.columns 
		WHERE table = 'admission_hubei_wide_2024' 
		AND database = 'default' 
		AND name = 'major_min_score_2024'
	`

	rows, err = conn.Query(ctx, checkQuery)
	if err != nil {
		log.Fatalf("检查字段失败: %v", err)
	}
	defer rows.Close()

	hasField := false
	for rows.Next() {
		var fieldName string
		if err := rows.Scan(&fieldName); err != nil {
			continue
		}
		if fieldName == "major_min_score_2024" {
			hasField = true
			break
		}
	}

	if !hasField {
		fmt.Println("添加专业最低分字段...")
		alterQuery := "ALTER TABLE default.admission_hubei_wide_2024 ADD COLUMN major_min_score_2024 Nullable(UInt16)"
		if err := conn.Exec(ctx, alterQuery); err != nil {
			log.Fatalf("添加字段失败: %v", err)
		}
		fmt.Println("✅ 字段添加成功")
	} else {
		fmt.Println("✅ 字段已存在")
	}

	// 5. 批量更新数据
	fmt.Println("\n5. 开始批量更新数据...")
	batchSize := 1000
	totalBatches := (len(records) + batchSize - 1) / batchSize
	updatedCount := 0

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		batchNum := (i / batchSize) + 1

		fmt.Printf("处理批次 %d/%d (%d 条记录)...\n", batchNum, totalBatches, len(batch))

		// 构建批量更新查询
		var updateCases []string
		var idList []string

		for _, record := range batch {
			updateCases = append(updateCases,
				fmt.Sprintf("WHEN %d THEN %d", record.ID, int(record.MajorMinScore2024)))
			idList = append(idList, fmt.Sprintf("%d", record.ID))
		}

		updateQuery := fmt.Sprintf(`
			ALTER TABLE default.admission_hubei_wide_2024
			UPDATE major_min_score_2024 = CASE id 
				%s
			END
			WHERE id IN (%s)
		`, strings.Join(updateCases, " "), strings.Join(idList, ","))

		if err := conn.Exec(ctx, updateQuery); err != nil {
			log.Printf("批次 %d 更新失败: %v", batchNum, err)
			continue
		}

		updatedCount += len(batch)

		// 显示进度
		if batchNum%5 == 0 || batchNum == totalBatches {
			fmt.Printf("已更新: %d/%d (%.1f%%)\n",
				updatedCount, len(records),
				float64(updatedCount)/float64(len(records))*100)
		}
	}

	// 6. 验证更新结果
	fmt.Println("\n6. 验证更新结果...")
	verifyUpdateQuery := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(major_min_score_2024) as updated_count,
			AVG(major_min_score_2024) as avg_score,
			MIN(major_min_score_2024) as min_score,
			MAX(major_min_score_2024) as max_score
		FROM default.admission_hubei_wide_2024
	`

	rows, err = conn.Query(ctx, verifyUpdateQuery)
	if err != nil {
		log.Printf("验证查询失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var totalCount, updatedCount uint64
			var avgScore, minScore, maxScore *float64

			if err := rows.Scan(&totalCount, &updatedCount, &avgScore, &minScore, &maxScore); err != nil {
				log.Printf("扫描验证结果失败: %v", err)
				continue
			}

			fmt.Printf("验证结果:\n")
			fmt.Printf("  总记录数: %d\n", totalCount)
			fmt.Printf("  已更新数: %d\n", updatedCount)
			if avgScore != nil {
				fmt.Printf("  平均分数: %.1f\n", *avgScore)
			}
			if minScore != nil && maxScore != nil {
				fmt.Printf("  分数范围: %.0f - %.0f\n", *minScore, *maxScore)
			}

			if updatedCount > 0 {
				updateRate := float64(updatedCount) / float64(totalCount) * 100
				fmt.Printf("  更新覆盖率: %.1f%%\n", updateRate)
			}
		}
	}

	// 7. 查看更新样本
	fmt.Println("\n7. 查看更新样本...")
	sampleQuery := `
		SELECT id, school_name, major_name, major_min_score_2024, min_score_2024
		FROM default.admission_hubei_wide_2024 
		WHERE major_min_score_2024 IS NOT NULL
		ORDER BY major_min_score_2024 DESC
		LIMIT 10
	`

	rows, err = conn.Query(ctx, sampleQuery)
	if err != nil {
		log.Printf("样本查询失败: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("高分专业样本:")
		for rows.Next() {
			var id uint32
			var schoolName, majorName string
			var majorMinScore, minScore *uint16

			if err := rows.Scan(&id, &schoolName, &majorName, &majorMinScore, &minScore); err != nil {
				log.Printf("扫描样本失败: %v", err)
				continue
			}

			majorScoreStr := "null"
			minScoreStr := "null"
			if majorMinScore != nil {
				majorScoreStr = fmt.Sprintf("%d", *majorMinScore)
			}
			if minScore != nil {
				minScoreStr = fmt.Sprintf("%d", *minScore)
			}

			fmt.Printf("  ID:%d %s-%s 专业最低分:%s 原最低分:%s\n",
				id, schoolName, majorName, majorScoreStr, minScoreStr)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✅ 专业最低分更新完成！更新了 %d 条记录\n", updatedCount)
}

func readUpdateCSV(filename string) ([]UpdateRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rawRecords, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rawRecords) == 0 {
		return nil, fmt.Errorf("CSV文件为空")
	}

	// 跳过标题行
	var records []UpdateRecord
	for i := 1; i < len(rawRecords); i++ {
		record := rawRecords[i]
		if len(record) < 5 {
			continue
		}

		id, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			continue
		}

		score, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		records = append(records, UpdateRecord{
			ID:                uint32(id),
			MajorMinScore2024: score,
			CollegeName:       record[2],
			MajorName:         record[3],
			SourceProvince:    record[4],
		})
	}

	return records, nil
}
