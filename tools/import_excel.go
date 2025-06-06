package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"gaokao-zhiyuan/config"
	"gaokao-zhiyuan/database"
	"gaokao-zhiyuan/models"

	"github.com/tealeg/xlsx"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接数据库
	db, err := database.NewClickHouseDB(cfg)
	if err != nil {
		log.Fatalf("连接ClickHouse失败: %v", err)
	}
	defer db.Close()

	// 创建表
	if err := db.CreateTable(); err != nil {
		log.Fatalf("创建表失败: %v", err)
	}

	// 导入Excel数据
	excelFile := "21-24各省份录取数据(含专业组代码).xlsx"
	if err := importExcelData(db, excelFile); err != nil {
		log.Fatalf("导入数据失败: %v", err)
	}

	log.Println("数据导入完成！")
}

func importExcelData(db *database.ClickHouseDB, filename string) error {
	// 打开Excel文件
	xlFile, err := xlsx.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %v", err)
	}

	var allData []models.AdmissionData
	id := int64(1)

	// 遍历所有工作表
	for sheetIndex, sheet := range xlFile.Sheets {
		log.Printf("正在处理工作表 %d: %s", sheetIndex+1, sheet.Name)
		
		// 跳过表头行
		if len(sheet.Rows) <= 1 {
			continue
		}

		for rowIndex, row := range sheet.Rows {
			if rowIndex == 0 { // 跳过表头
				continue
			}

			// 确保行有足够的列
			if len(row.Cells) < 10 {
				continue
			}

			data := models.AdmissionData{
				ID: id,
			}
			id++

			// 解析每一列的数据
			for colIndex, cell := range row.Cells {
				cellValue := strings.TrimSpace(cell.String())
				
				switch colIndex {
				case 0: // 年份
					if year, err := strconv.Atoi(cellValue); err == nil {
						data.Year = year
					}
				case 1: // 省份
					data.Province = cellValue
				case 2: // 院校名称
					data.CollegeName = cellValue
				case 3: // 院校代码
					data.CollegeCode = cellValue
				case 4: // 专业组代码
					data.SpecialInterestGroupCode = cellValue
				case 5: // 专业名称
					data.ProfessionalName = cellValue
				case 6: // 选科要求
					data.ClassDemand = cellValue
				case 7: // 录取最低分
					if points, err := strconv.ParseInt(cellValue, 10, 64); err == nil {
						data.LowestPoints = points
					}
				case 8: // 录取最低位次
					if rank, err := strconv.ParseInt(cellValue, 10, 64); err == nil {
						data.LowestRank = rank
					}
				case 9: // 备注
					data.Description = cellValue
				}
			}

			// 验证必要字段
			if data.Year > 0 && data.CollegeName != "" && data.ProfessionalName != "" {
				allData = append(allData, data)
			}

			// 每1000条批量插入一次
			if len(allData) >= 1000 {
				if err := db.BatchInsert(allData); err != nil {
					return fmt.Errorf("批量插入数据失败: %v", err)
				}
				log.Printf("已插入 %d 条数据", len(allData))
				allData = []models.AdmissionData{}
			}
		}
	}

	// 插入剩余数据
	if len(allData) > 0 {
		if err := db.BatchInsert(allData); err != nil {
			return fmt.Errorf("插入剩余数据失败: %v", err)
		}
		log.Printf("已插入剩余 %d 条数据", len(allData))
	}

	return nil
} 