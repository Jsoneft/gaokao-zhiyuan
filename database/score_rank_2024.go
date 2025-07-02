package database

import (
	"encoding/json"
	"gaokao-zhiyuan/models"
	"log"
	"os"
	"strconv"
	"strings"
)

// 一分一段表JSON数据结构
type ScoreRankEntry struct {
	Score      string `json:"score"`
	Num        int    `json:"num"`
	Accumulate int    `json:"accumulate"`
}

type ScoreRankJSON struct {
	Data []ScoreRankEntry `json:"data"`
}

// 2024年湖北省一分一段表数据（从官方JSON文件加载）
var scoreRankTable2024 models.ScoreRankTable2024

// 初始化函数，加载官方一分一段表数据
func init() {
	loadScoreRankData()
}

// 加载官方一分一段表数据
func loadScoreRankData() {
	// 加载物理类数据
	physicsData := loadJSONFile("hubei_data/ranking_score_hubei_physics.json")
	scoreRankTable2024.Physics = convertToScoreRankData(physicsData)

	// 加载历史类数据
	historyData := loadJSONFile("hubei_data/ranking_score_hubei_history.json")
	scoreRankTable2024.History = convertToScoreRankData(historyData)

	log.Printf("已加载2024年湖北省一分一段表数据：物理类 %d 条，历史类 %d 条",
		len(scoreRankTable2024.Physics), len(scoreRankTable2024.History))
}

// 从JSON文件加载数据
func loadJSONFile(filename string) []ScoreRankEntry {
	// 先尝试原路径
	file, err := os.Open(filename)
	if err != nil {
		// 如果失败，尝试相对于当前工作目录的路径
		log.Printf("尝试打开文件 %s 失败: %v，尝试其他路径", filename, err)

		// 获取当前工作目录
		pwd, _ := os.Getwd()
		log.Printf("当前工作目录: %s", pwd)

		// 检查文件是否存在
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Printf("文件 %s 不存在，将使用默认数据", filename)
			return []ScoreRankEntry{}
		}

		return []ScoreRankEntry{}
	}
	defer file.Close()

	var jsonData ScoreRankJSON
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&jsonData)
	if err != nil {
		log.Printf("解析JSON文件 %s 失败: %v，将使用默认数据", filename, err)
		return []ScoreRankEntry{}
	}

	log.Printf("成功加载JSON文件 %s，共 %d 条记录", filename, len(jsonData.Data))
	return jsonData.Data
}

// 将JSON数据转换为ScoreRankData格式
func convertToScoreRankData(entries []ScoreRankEntry) []models.ScoreRankData {
	var result []models.ScoreRankData

	for _, entry := range entries {
		// 处理分数字段，可能是单个分数或范围
		scores := parseScoreField(entry.Score)

		for _, score := range scores {
			result = append(result, models.ScoreRankData{
				Score: score,
				Rank:  entry.Accumulate, // 使用累计人数作为排名
			})
		}
	}

	// 按分数降序排列（高分在前）
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Score < result[j].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// 解析分数字段，处理单个分数和分数范围
func parseScoreField(scoreStr string) []int {
	var scores []int

	// 处理分数范围，如 "695-750"
	if strings.Contains(scoreStr, "-") {
		parts := strings.Split(scoreStr, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(parts[0])
			_, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				// 对于范围，我们使用起始分数
				scores = append(scores, start)
			}
		}
	} else {
		// 处理单个分数
		score, err := strconv.Atoi(scoreStr)
		if err == nil {
			scores = append(scores, score)
		}
	}

	return scores
}

// GetRankByScore2024 根据分数和首选科目查询2024年一分一段表排名
func GetRankByScore2024(score int, subjectType string) int {
	var data []models.ScoreRankData

	// 根据首选科目选择对应的一分一段表
	if subjectType == "物理" {
		data = scoreRankTable2024.Physics
	} else if subjectType == "历史" {
		data = scoreRankTable2024.History
	} else {
		// 默认使用物理类
		data = scoreRankTable2024.Physics
	}

	// 数据为空时的异常处理（理论上不应该发生）
	if len(data) == 0 {
		log.Printf("严重错误：一分一段表数据为空")
		return 1 // 返回最佳排名作为默认值
	}

	// 如果分数高于最高分，返回最高排名（最佳排名）
	if score >= data[0].Score {
		return data[0].Rank
	}

	// 如果分数低于最低分，返回最低排名（最差排名）
	if score <= data[len(data)-1].Score {
		return data[len(data)-1].Rank
	}

	// 线性插值查找对应排名
	for i := 0; i < len(data)-1; i++ {
		if score <= data[i].Score && score >= data[i+1].Score {
			// 线性插值计算排名
			scoreRange := data[i].Score - data[i+1].Score
			rankRange := data[i+1].Rank - data[i].Rank

			if scoreRange == 0 {
				return data[i].Rank
			}

			scoreDiff := score - data[i+1].Score
			interpolatedRank := data[i+1].Rank - (rankRange * scoreDiff / scoreRange)

			// 确保插值结果为正数
			return ensurePositiveRank(interpolatedRank)
		}
	}

	// 理论上不应该到达这里，但作为保险返回中位排名
	log.Printf("警告：分数 %d 未找到对应区间，返回中位排名", score)
	midIndex := len(data) / 2
	return data[midIndex].Rank
}

// ensurePositiveRank 确保排名为正数，最小值为1
func ensurePositiveRank(rank int) int {
	if rank <= 0 {
		return 1 // 最好的排名是第1名
	}
	return rank
}

// GetSubjectTypeFromClassDemand 从选科要求中推断首选科目
func GetSubjectTypeFromClassDemand(classDemand string) string {
	if classDemand == "" {
		return "物理" // 默认物理类
	}

	// 如果包含物理，认为是物理类
	if contains(classDemand, "物理") || contains(classDemand, "物") {
		return "物理"
	}

	// 如果包含历史，认为是历史类
	if contains(classDemand, "历史") || contains(classDemand, "史") {
		return "历史"
	}

	// 默认返回物理类
	return "物理"
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr))))
}

// containsInMiddle 检查字符串中间是否包含子字符串
func containsInMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
