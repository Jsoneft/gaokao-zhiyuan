package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gaokao-zhiyuan/config"
	"gaokao-zhiyuan/models"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseDB struct {
	conn driver.Conn
}

func NewClickHouseDB(cfg *config.Config) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouseHost, cfg.ClickHousePort)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDatabase,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePassword,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &ClickHouseDB{conn: conn}, nil
}

func (db *ClickHouseDB) Close() error {
	return db.conn.Close()
}

// 创建表
func (db *ClickHouseDB) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS admission_data (
		id UInt64,
		year UInt32,
		province String,
		college_name String,
		college_code String,
		special_interest_group_code String,
		professional_name String,
		class_demand String,
		lowest_points Int64,
		lowest_rank Int64,
		description String
	) ENGINE = MergeTree()
	ORDER BY (year, lowest_points, lowest_rank)
	`
	return db.conn.Exec(context.Background(), query)
}

// 批量插入数据
func (db *ClickHouseDB) BatchInsert(data []models.AdmissionData) error {
	batch, err := db.conn.PrepareBatch(context.Background(), 
		"INSERT INTO admission_data (id, year, province, college_name, college_code, special_interest_group_code, professional_name, class_demand, lowest_points, lowest_rank, description)")
	if err != nil {
		return err
	}

	for _, item := range data {
		err := batch.Append(
			item.ID,
			item.Year,
			item.Province,
			item.CollegeName,
			item.CollegeCode,
			item.SpecialInterestGroupCode,
			item.ProfessionalName,
			item.ClassDemand,
			item.LowestPoints,
			item.LowestRank,
			item.Description,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

// 根据分数查询位次
func (db *ClickHouseDB) GetRankByScore(score int64) (*models.RankResponse, error) {
	var rank int64
	var year int
	
	// 查询2024年数据中对应分数的位次
	query := `
	SELECT lowest_rank, year FROM admission_data 
	WHERE year = 2024 AND lowest_points <= ? 
	ORDER BY lowest_points DESC 
	LIMIT 1
	`
	
	row := db.conn.QueryRow(context.Background(), query, score)
	err := row.Scan(&rank, &year)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.RankResponse{
				Code: 1,
				Msg:  "未找到对应分数的位次信息",
			}, nil
		}
		return nil, err
	}

	return &models.RankResponse{
		Code: 0,
		Msg:  "success",
		Rank: rank,
		Year: year,
	}, nil
}

// 查询报表数据
func (db *ClickHouseDB) GetReportData(rank int64, classComb string, page, pageSize int64) (*models.Response, error) {
	// 首先根据位次计算去年等位分
	var lastYearScore int64
	scoreQuery := `
	SELECT lowest_points FROM admission_data 
	WHERE year = 2023 AND lowest_rank >= ? 
	ORDER BY lowest_rank ASC 
	LIMIT 1
	`
	
	row := db.conn.QueryRow(context.Background(), scoreQuery, rank)
	err := row.Scan(&lastYearScore)
	if err != nil {
		if err == sql.ErrNoRows {
			lastYearScore = 0
		} else {
			return nil, err
		}
	}

	// 计算分数范围：去年等位分上+20分，下减30分
	upperScore := lastYearScore + 20
	lowerScore := lastYearScore - 30
	if lowerScore < 0 {
		lowerScore = 0
	}

	// 构建选科条件
	classCondition := buildClassCondition(classComb)
	
	// 查询总数
	countQuery := fmt.Sprintf(`
	SELECT COUNT(*) FROM admission_data 
	WHERE year = 2024 
	AND lowest_points BETWEEN ? AND ? 
	%s
	`, classCondition)
	
	var totalCount int64
	row = db.conn.QueryRow(context.Background(), countQuery, lowerScore, upperScore)
	row.Scan(&totalCount)

	// 计算分页
	offset := (page - 1) * pageSize
	totalPages := (totalCount + pageSize - 1) / pageSize

	// 查询数据
	dataQuery := fmt.Sprintf(`
	SELECT id, college_code, college_name, special_interest_group_code, 
		   professional_name, class_demand, lowest_points, lowest_rank, description
	FROM admission_data 
	WHERE year = 2024 
	AND lowest_points BETWEEN ? AND ? 
	%s
	ORDER BY lowest_points DESC
	LIMIT ? OFFSET ?
	`, classCondition)

	rows, err := db.conn.Query(context.Background(), dataQuery, lowerScore, upperScore, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.List
	for rows.Next() {
		var item models.List
		var id int64
		var collegeCode, collegeName, groupCode, profName, classDemand, desc string
		var lowestPoints, lowestRank int64

		err := rows.Scan(&id, &collegeCode, &collegeName, &groupCode, &profName, &classDemand, &lowestPoints, &lowestRank, &desc)
		if err != nil {
			log.Printf("扫描行数据错误: %v", err)
			continue
		}

		item = models.List{
			ID:                       &id,
			ColledgeCode:             &collegeCode,
			ColledgeName:             &collegeName,
			SpecialInterestGroupCode: &groupCode,
			ProfessionalName:         profName,
			ClassDemand:              &classDemand,
			LowestPoints:             &lowestPoints,
			LowestRank:               &lowestRank,
			Description:              &desc,
		}
		list = append(list, item)
	}

	conf := &models.Conf{
		Page:        page,
		PageSize:    pageSize,
		TotalNumber: totalCount,
		TotalPage:   totalPages,
	}

	return &models.Response{
		Code: 0,
		Msg:  "success",
		Data: models.Data{
			Conf: conf,
			List: list,
		},
	}, nil
}

// 构建选科条件
func buildClassCondition(classComb string) string {
	if classComb == "" {
		return ""
	}

	// 移除引号
	classComb = strings.Trim(classComb, "\"")
	
	// 物理、化学、生物、政治、历史、地理
	// 1     2     3     4     5     6
	subjectMap := map[string]string{
		"1": "物理",
		"2": "化学", 
		"3": "生物",
		"4": "政治",
		"5": "历史",
		"6": "地理",
	}

	var subjects []string
	for _, char := range classComb {
		if subject, exists := subjectMap[string(char)]; exists {
			subjects = append(subjects, subject)
		}
	}

	if len(subjects) == 0 {
		return ""
	}

	// 构建SQL条件，选科要求包含用户选的科目或者不限
	var conditions []string
	for _, subject := range subjects {
		conditions = append(conditions, fmt.Sprintf("class_demand LIKE '%%%s%%'", subject))
	}
	
	return fmt.Sprintf("AND (class_demand = '不限' OR %s)", strings.Join(conditions, " OR "))
} 