package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	// 先尝试连接到指定数据库
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
		// 如果连接失败，可能是数据库不存在，尝试连接默认数据库并创建
		conn.Close()

		// 连接到默认数据库
		defaultConn, err := clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouseHost, cfg.ClickHousePort)},
			Auth: clickhouse.Auth{
				Username: cfg.ClickHouseUser,
				Password: cfg.ClickHousePassword,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("连接默认数据库失败: %v", err)
		}

		// 创建目标数据库
		if err := defaultConn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.ClickHouseDatabase)); err != nil {
			defaultConn.Close()
			return nil, fmt.Errorf("创建数据库失败: %v", err)
		}
		defaultConn.Close()

		// 重新连接到目标数据库
		conn, err = clickhouse.Open(&clickhouse.Options{
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
		batch String,
		subject_type String,
		class_demand String,
		college_code String,
		special_interest_group_code String,
		college_name String,
		professional_code String,
		professional_name String,
		lowest_points Int64,
		lowest_rank Int64,
		description String
	) ENGINE = MergeTree()
	ORDER BY (lowest_rank, lowest_points, year, province)
	`
	return db.conn.Exec(context.Background(), query)
}

// 批量插入数据
func (db *ClickHouseDB) BatchInsert(data []models.AdmissionData) error {
	batch, err := db.conn.PrepareBatch(context.Background(),
		"INSERT INTO admission_data (id, year, province, batch, subject_type, class_demand, college_code, special_interest_group_code, college_name, professional_code, professional_name, lowest_points, lowest_rank, description)")
	if err != nil {
		return err
	}

	for _, item := range data {
		err := batch.Append(
			item.ID,
			item.Year,
			item.Province,
			item.Batch,
			item.SubjectType,
			item.ClassDemand,
			item.CollegeCode,
			item.SpecialInterestGroupCode,
			item.CollegeName,
			item.ProfessionalCode,
			item.ProfessionalName,
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
func (db *ClickHouseDB) QueryRankByScore(province string, year int, score float64, subjectType string, classDemands []string) (int64, error) {
	// 构建科目类型和选科要求的条件
	classDemandCondition := ""
	if len(classDemands) > 0 {
		conditions := make([]string, 0, len(classDemands))
		for _, demand := range classDemands {
			conditions = append(conditions, fmt.Sprintf("class_demand LIKE '%%%s%%'", demand))
		}
		classDemandCondition = "AND (" + strings.Join(conditions, " OR ") + ")"
	}

	// 查询语句：根据分数查询位次
	// 使用lowest_rank排序，找到分数大于等于给定分数的最大位次
	query := fmt.Sprintf(`
		SELECT lowest_rank
		FROM gaokao.admission_data
		WHERE province = $1
		AND year = $2
		AND subject_type = $3
		%s
		AND lowest_points >= $4
		ORDER BY lowest_points ASC
		LIMIT 1
	`, classDemandCondition)

	var rank int64
	err := db.conn.QueryRow(context.Background(), query, province, year, subjectType, score).Scan(&rank)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到记录，查询该省份该年份最低分最高的记录的位次
			estimateQuery := `
				SELECT lowest_rank
				FROM gaokao.admission_data
				WHERE province = $1
				AND year = $2
				AND subject_type = $3
				AND lowest_points > 0
				ORDER BY lowest_points DESC
				LIMIT 1
			`
			var estimateRank int64
			err = db.conn.QueryRow(context.Background(), estimateQuery, province, year, subjectType).Scan(&estimateRank)
			if err != nil {
				return 0, errors.New("无法估算位次")
			}
			return estimateRank, nil
		}
		return 0, err
	}

	return rank, nil
}

// 根据位次查询分数
func (db *ClickHouseDB) QueryScoreByRank(province string, year int, rank int64, subjectType string, classDemands []string) (int64, error) {
	// 构建科目类型和选科要求的条件
	classDemandCondition := ""
	if len(classDemands) > 0 {
		conditions := make([]string, 0, len(classDemands))
		for _, demand := range classDemands {
			conditions = append(conditions, fmt.Sprintf("class_demand LIKE '%%%s%%'", demand))
		}
		classDemandCondition = "AND (" + strings.Join(conditions, " OR ") + ")"
	}

	// 查询语句：根据位次查询分数
	// 使用lowest_rank排序，找到位次小于等于给定位次的最低分
	query := fmt.Sprintf(`
		SELECT lowest_points
		FROM gaokao.admission_data
		WHERE province = $1
		AND year = $2
		AND subject_type = $3
		%s
		AND lowest_rank <= $4
		AND lowest_rank > 0
		ORDER BY lowest_rank DESC
		LIMIT 1
	`, classDemandCondition)

	var score int64
	err := db.conn.QueryRow(context.Background(), query, province, year, subjectType, rank).Scan(&score)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到记录，查询该省份该年份最高位次最低的记录的分数
			estimateQuery := `
				SELECT lowest_points
				FROM gaokao.admission_data
				WHERE province = $1
				AND year = $2
				AND subject_type = $3
				AND lowest_rank > 0
				ORDER BY lowest_rank ASC
				LIMIT 1
			`
			var estimateScore int64
			err = db.conn.QueryRow(context.Background(), estimateQuery, province, year, subjectType).Scan(&estimateScore)
			if err != nil {
				return 0, errors.New("无法估算分数")
			}
			return estimateScore, nil
		}
		return 0, err
	}

	return score, nil
}

// 查询报表数据
func (db *ClickHouseDB) GetReportData(rank int64, classComb string, page, pageSize int64) (*models.Response, error) {
	// 获取2024年对应位次的分数
	var rankScore int64
	scoreQuery := `
	SELECT lowest_points FROM gaokao.admission_data 
	WHERE year = 2024 AND lowest_rank <= ? AND lowest_rank > 0
	ORDER BY lowest_rank DESC 
	LIMIT 1
	`

	row := db.conn.QueryRow(context.Background(), scoreQuery, rank)
	err := row.Scan(&rankScore)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到精确位次，查询附近的位次
			nearbyQuery := `
			SELECT lowest_points FROM gaokao.admission_data 
			WHERE year = 2024 AND lowest_rank > 0
			ORDER BY ABS(lowest_rank - ?)
			LIMIT 1
			`
			row = db.conn.QueryRow(context.Background(), nearbyQuery, rank)
			err = row.Scan(&rankScore)
			if err != nil {
				rankScore = 500 // 默认分数
			}
		} else {
			return nil, err
		}
	}

	// 计算分数范围：在位次对应分数基础上，上浮+20分，下浮-30分
	upperScore := rankScore + 20
	lowerScore := rankScore - 30
	if lowerScore < 0 {
		lowerScore = 0
	}

	// 构建选科条件
	classCondition := buildClassCondition(classComb)

	// 查询总数
	countQuery := fmt.Sprintf(`
	SELECT COUNT(*) FROM gaokao.admission_data 
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
	FROM gaokao.admission_data 
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
		var id uint64
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

	// 调试输出
	log.Printf("处理选科组合: %s", classComb)

	// 物理、化学、生物、政治、历史、地理
	// 1     2     3     4     5     6
	subjectMap := map[string]string{
		"1": "物",
		"2": "化",
		"3": "生",
		"4": "政",
		"5": "历",
		"6": "地",
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

	log.Printf("选择科目: %v", subjects)

	// 构建SQL条件，选科要求包含用户选的科目或者不限
	var conditions []string
	for _, subject := range subjects {
		conditions = append(conditions, fmt.Sprintf("class_demand LIKE '%%%s%%'", subject))
	}

	// 添加不限选项
	conditions = append(conditions, "class_demand = '不限'", "class_demand = ''")

	return fmt.Sprintf("AND (%s)", strings.Join(conditions, " OR "))
}

// 获取数据记录数
func (db *ClickHouseDB) GetDataCount() (int64, error) {
	var count int64
	row := db.conn.QueryRow(context.Background(), "SELECT count() FROM gaokao.admission_data")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
