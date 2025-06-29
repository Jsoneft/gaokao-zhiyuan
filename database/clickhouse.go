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

// 创建新的湖北省数据表
func (db *ClickHouseDB) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS admission_hubei_wide_2024 (
		id                      UInt32,
		school_code             String,
		school_name             String,
		major_code              String,
		major_name              String,
		major_group_code        String,
		source_province         LowCardinality(String),
		school_province         LowCardinality(String),
		school_city             String,
		admission_batch         LowCardinality(String),
		subject_category        Enum8('物理'=1, '历史'=2),
		require_physics         Bool,
		require_chemistry       Bool,
		require_biology         Bool,
		require_politics        Bool,
		require_history         Bool,
		require_geography       Bool,
		subject_requirement_raw LowCardinality(String),
		school_type             LowCardinality(String),
		school_ownership        Enum8('公办'=1, '民办'=2),
		school_authority        LowCardinality(String),
		school_level            LowCardinality(String),
		school_tags             String,
		education_level         Enum8('本科'=1, '专科'=2),
		major_description       String,
		study_years             UInt8,
		tuition_fee             UInt32,
		is_new_major            Bool,
		min_score_2024          UInt16,
		min_rank_2024           UInt32,
		major_min_score_2024    Nullable(UInt16),
		enrollment_plan_2024    UInt16,
		is_science              Bool,
		is_engineering          Bool,
		is_medical              Bool,
		is_economics_mgmt_law   Bool,
		is_liberal_arts         Bool,
		is_design_arts          Bool,
		is_language             Bool
	) ENGINE = MergeTree()
	ORDER BY (id, school_code, major_code)
	SETTINGS index_granularity = 8192
	`
	return db.conn.Exec(context.Background(), query)
}

// 创建旧表（保持兼容性）
func (db *ClickHouseDB) CreateOldTable() error {
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

// 根据分数查询位次 - 使用新表
func (db *ClickHouseDB) QueryRankByScoreNew(score float64, subjectCategory string) (int64, error) {
	// 查询语句：根据分数查询位次
	query := `
		SELECT min_rank_2024
		FROM admission_hubei_wide_2024
		WHERE min_score_2024 >= $1
		AND min_rank_2024 > 0
		AND subject_category = $2
		ORDER BY min_score_2024 ASC
		LIMIT 1
	`

	var rank uint32
	err := db.conn.QueryRow(context.Background(), query, score, subjectCategory).Scan(&rank)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到记录，查询最高分对应的位次
			estimateQuery := `
				SELECT min_rank_2024
				FROM admission_hubei_wide_2024
				WHERE min_score_2024 > 0
				AND min_rank_2024 > 0
				AND subject_category = $1
				ORDER BY min_score_2024 DESC
				LIMIT 1
			`
			var estimateRank uint32
			err = db.conn.QueryRow(context.Background(), estimateQuery, subjectCategory).Scan(&estimateRank)
			if err != nil {
				return 0, errors.New("无法估算位次")
			}
			return int64(estimateRank), nil
		}
		return 0, err
	}

	return int64(rank), nil
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

// 新的报表查询接口 - 使用新表结构
func (db *ClickHouseDB) GetReportDataNew(rank int64, classFirstChoice string, classOptionalChoice []string, province string, page, pageSize int64, collegeLocation []string, interest []string, strategy int, fuzzySubjectCategory string) (*models.Response, error) {
	log.Printf("报表查询参数: rank=%d, classFirstChoice=%s, classOptionalChoice=%v, province=%s, page=%d, pageSize=%d, collegeLocation=%v, interest=%v, strategy=%d, fuzzySubjectCategory=%s",
		rank, classFirstChoice, classOptionalChoice, province, page, pageSize, collegeLocation, interest, strategy, fuzzySubjectCategory)

	// 根据位次查询对应分数
	var rankScoreUint16 uint16
	scoreQuery := `
		SELECT min_score_2024 
		FROM default.admission_hubei_wide_2024 
		WHERE min_rank_2024 <= ? AND min_rank_2024 > 0 AND subject_category = ?
		ORDER BY min_rank_2024 DESC 
		LIMIT 1
	`

	row := db.conn.QueryRow(context.Background(), scoreQuery, rank, classFirstChoice)
	err := row.Scan(&rankScoreUint16)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到精确位次，查询附近的位次
			nearbyQuery := `
				SELECT min_score_2024 
				FROM default.admission_hubei_wide_2024 
				WHERE min_rank_2024 > 0 AND subject_category = ?
				ORDER BY ABS(min_rank_2024 - ?)
				LIMIT 1
			`
			row = db.conn.QueryRow(context.Background(), nearbyQuery, classFirstChoice, rank)
			err = row.Scan(&rankScoreUint16)
			if err != nil {
				log.Printf("无法找到位次 %d 附近的数据，使用默认分数 500", rank)
				rankScoreUint16 = 500 // 默认分数
			} else {
				log.Printf("找到位次 %d 附近的数据，对应分数为 %d", rank, rankScoreUint16)
			}
		} else {
			log.Printf("查询位次 %d 对应分数时出错: %v", rank, err)
			return nil, err
		}
	} else {
		log.Printf("位次 %d 对应的分数为 %d", rank, rankScoreUint16)
	}

	// 注意：rankScore 变量在新的排名策略中不再需要，因为我们直接使用用户输入的排名

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 1. 一次筛选：选科分类
	subjectConditions := db.buildSubjectConditions(classFirstChoice, classOptionalChoice)
	if subjectConditions != "" {
		conditions = append(conditions, subjectConditions)
	}

	// 2. 二次筛选：院校所在省份
	if len(collegeLocation) > 0 {
		locationConditions := make([]string, len(collegeLocation))
		for i, location := range collegeLocation {
			locationConditions[i] = fmt.Sprintf("school_province = $%d", argIndex)
			args = append(args, location)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(locationConditions, " OR ")))
	}

	// 3. 三次筛选：意向专业方向
	if len(interest) > 0 {
		interestConditions := db.buildInterestConditions(interest)
		if interestConditions != "" {
			conditions = append(conditions, interestConditions)
		}
	}

	// 4. 模糊专业名称筛选
	if fuzzySubjectCategory != "" {
		conditions = append(conditions, fmt.Sprintf("major_name LIKE $%d", argIndex))
		args = append(args, "%"+fuzzySubjectCategory+"%")
		argIndex++
	}

	// 5. 分数（省生源地排位）筛选 - 冲稳保策略
	var upperScore, lowerScore int64
	var minScoreDiff, maxScoreDiff int64

	switch strategy {
	case 0: // 冲
		minScoreDiff = 3  // 分数比最低分高3分
		maxScoreDiff = 20 // 分数比最低分高20分
	case 1: // 稳
		minScoreDiff = -5 // 分数比最低分低5分
		maxScoreDiff = 3  // 分数比最低分高3分
	case 2: // 保
		minScoreDiff = -20 // 分数比最低分低20分
		maxScoreDiff = -5  // 分数比最低分低5分
	default: // 冲稳保混合
		minScoreDiff = -20 // 从保到冲的完整范围
		maxScoreDiff = 20
	}

	// 计算实际分数范围
	rankScore := int64(rankScoreUint16)
	lowerScore = rankScore
	upperScore = rankScore

	if minScoreDiff >= 0 {
		lowerScore = rankScore + minScoreDiff
	} else {
		if -minScoreDiff <= rankScore {
			lowerScore = rankScore - (-minScoreDiff)
		} else {
			lowerScore = 0 // 防止下溢
		}
	}

	if maxScoreDiff >= 0 {
		upperScore = rankScore + maxScoreDiff
	} else {
		if -maxScoreDiff <= rankScore {
			upperScore = rankScore - (-maxScoreDiff)
		} else {
			upperScore = 0 // 防止下溢
		}
	}

	conditions = append(conditions, fmt.Sprintf("min_score_2024 BETWEEN $%d AND $%d", argIndex, argIndex+1))
	args = append(args, lowerScore, upperScore)
	argIndex += 2

	// 构建WHERE子句
	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	} else {
		whereClause = ""
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) AS total_count 
		FROM default.admission_hubei_wide_2024 
		%s
	`, whereClause)

	log.Printf("执行计数查询: %s, args: %v", countQuery, args)
	var totalCountUint uint64
	err = db.conn.QueryRow(context.Background(), countQuery, args...).Scan(&totalCountUint)
	if err != nil {
		log.Printf("计数查询失败: %v", err)
		totalCountUint = 0
	}
	totalCount := int64(totalCountUint)
	log.Printf("查询到符合条件的记录总数: %d", totalCount)

	// 计算分页
	offset := (page - 1) * pageSize
	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + pageSize - 1) / pageSize
	}

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, school_name, school_code, major_group_code, 
			   subject_requirement_raw, school_province, school_city, 
			   school_ownership, school_type, school_authority, school_level, 
			   school_tags, education_level, major_description, tuition_fee, is_new_major,
			   min_score_2024, min_rank_2024, major_name, study_years, major_min_score_2024
		FROM default.admission_hubei_wide_2024 
		%s
		ORDER BY min_score_2024 DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	log.Printf("执行数据查询: %s", dataQuery)
	rows, err := db.conn.Query(context.Background(), dataQuery, args...)
	if err != nil {
		log.Printf("数据查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var list []models.List
	for rows.Next() {
		var item models.List
		var id uint32
		var schoolName, schoolCode, groupCode, subjectReq, schoolProvince, schoolCity string
		var schoolOwnership, schoolType, schoolAuthority, schoolLevel, schoolTags string
		var educationLevel, majorDesc, majorName string
		var tuitionFee uint32
		var isNewMajor bool
		var minScore uint16
		var minRank uint32
		var studyYears sql.NullString
		var majorMinScore *uint16

		err := rows.Scan(&id, &schoolName, &schoolCode, &groupCode, &subjectReq,
			&schoolProvince, &schoolCity, &schoolOwnership, &schoolType, &schoolAuthority,
			&schoolLevel, &schoolTags, &educationLevel, &majorDesc, &tuitionFee, &isNewMajor,
			&minScore, &minRank, &majorName, &studyYears, &majorMinScore)
		if err != nil {
			log.Printf("扫描行数据错误: %v", err)
			continue
		}

		// 处理学制字段
		var studyYearsPtr *string
		if studyYears.Valid {
			studyYearsPtr = &studyYears.String
		}

		// 处理专业最低分字段 - 直接使用扫描出的指针
		var majorMinScorePtr *uint16 = majorMinScore

		// 计算专业最低分对应的2024年排名
		var majorMinRank2024Ptr *int
		if majorMinScore != nil && *majorMinScore > 0 {
			// 直接使用用户选择的首选科目类型
			subjectType := classFirstChoice
			// 使用专业最低分计算2024年排名
			rank2024 := GetRankByScore2024(int(*majorMinScore), subjectType)
			majorMinRank2024Ptr = &rank2024
		}

		// 转换数据类型 - 确保类型匹配
		idUint64 := uint64(id)
		lowestPointsInt64 := int64(minScore)
		lowestRankInt64 := int64(minRank)

		item = models.List{
			ID:                       &idUint64,
			CollegeName:              &schoolName,
			CollegeCode:              &schoolCode,
			SpecialInterestGroupCode: &groupCode,
			ClassDemand:              &subjectReq,
			CollegeProvince:          &schoolProvince,
			CollegeCity:              &schoolCity,
			CollegeOwnership:         &schoolOwnership,
			CollegeType:              &schoolType,
			CollegeAuthority:         &schoolAuthority,
			CollegeLevel:             &schoolLevel,
			CollegeTags:              &schoolTags,
			EducationLevel:           &educationLevel,
			MajorDescription:         &majorDesc,
			TuitionFee:               &tuitionFee,
			IsNewMajor:               &isNewMajor,
			LowestPoints:             &lowestPointsInt64,
			LowestRank:               &lowestRankInt64,
			ProfessionalName:         majorName,
			StudyYears:               studyYearsPtr,
			MajorMinScore2024:        majorMinScorePtr,
			MajorMinRank2024:         majorMinRank2024Ptr,
		}
		list = append(list, item)
	}
	log.Printf("查询到 %d 条符合条件的记录", len(list))

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
func (db *ClickHouseDB) buildSubjectConditions(classFirstChoice string, classOptionalChoice []string) string {
	var conditions []string

	// 首选科目条件 - 使用subject_category字段匹配
	if classFirstChoice == "物理" {
		conditions = append(conditions, "subject_category = '物理'")
	} else if classFirstChoice == "历史" {
		conditions = append(conditions, "subject_category = '历史'")
	}

	// 可选科目条件 - 简化逻辑：用户选择的科目能够满足专业要求
	if len(classOptionalChoice) > 0 {
		// 解析可选科目
		subjectMap := map[string]string{
			"化学": "require_chemistry",
			"生物": "require_biology",
			"政治": "require_politics",
			"历史": "require_history",
			"地理": "require_geography",
		}

		// 用户没有选择的科目，专业不能要求
		userSelectedSubjects := make(map[string]bool)
		for _, subject := range classOptionalChoice {
			userSelectedSubjects[subject] = true
		}

		var subjectConditions []string
		allSubjects := []string{"化学", "生物", "政治", "历史", "地理"}
		for _, subject := range allSubjects {
			if field, exists := subjectMap[subject]; exists {
				if userSelectedSubjects[subject] {
					// 用户选择了这个科目，专业可以要求也可以不要求
					// 不添加限制条件
				} else {
					// 用户没有选择这个科目，专业不能要求
					subjectConditions = append(subjectConditions, fmt.Sprintf("%s = false", field))
				}
			}
		}

		if len(subjectConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(subjectConditions, " AND ")+")")
		}
	}

	if len(conditions) > 0 {
		return strings.Join(conditions, " AND ")
	}
	return ""
}

// 构建专业兴趣条件
func (db *ClickHouseDB) buildInterestConditions(interests []string) string {
	interestMap := map[string][]string{
		"理科":     {"数学", "物理", "化学", "生物", "天文", "地理", "统计"},
		"工科":     {"工程", "机械", "电子", "计算机", "软件", "土木", "建筑", "材料"},
		"文科":     {"文学", "历史", "哲学", "语言", "新闻", "传播", "艺术"},
		"经管法":    {"经济", "管理", "商务", "金融", "法学", "法律", "会计"},
		"医科":     {"医学", "临床", "护理", "药学", "中医", "口腔"},
		"设计与艺术类": {"设计", "艺术", "美术", "音乐", "舞蹈", "戏剧"},
		"语言类":    {"英语", "日语", "法语", "德语", "俄语", "西班牙语", "阿拉伯语"},
	}

	var conditions []string
	for _, interest := range interests {
		if keywords, exists := interestMap[interest]; exists {
			var keywordConditions []string
			for _, keyword := range keywords {
				keywordConditions = append(keywordConditions, fmt.Sprintf("major_name LIKE '%%%s%%'", keyword))
			}
			if len(keywordConditions) > 0 {
				conditions = append(conditions, "("+strings.Join(keywordConditions, " OR ")+")")
			}
		}
	}

	if len(conditions) > 0 {
		return "(" + strings.Join(conditions, " OR ") + ")"
	}
	return ""
}

// 查询报表数据
func (db *ClickHouseDB) GetReportData(rank int64, classComb string, province string, page, pageSize int64) (*models.Response, error) {
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
				log.Printf("无法找到位次 %d 附近的数据，使用默认分数 500", rank)
				rankScore = 500 // 默认分数
			} else {
				log.Printf("找到位次 %d 附近的数据，对应分数为 %d", rank, rankScore)
			}
		} else {
			log.Printf("查询位次 %d 对应分数时出错: %v", rank, err)
			return nil, err
		}
	} else {
		log.Printf("位次 %d 对应的分数为 %d", rank, rankScore)
	}

	// 计算分数范围：在位次对应分数基础上，上浮+20分，下浮-30分
	upperScore := rankScore + 20
	lowerScore := rankScore - 30
	if lowerScore < 0 {
		lowerScore = 0
	}
	log.Printf("分数范围设置为 %d-%d", lowerScore, upperScore)

	// 构建选科条件
	classCondition := buildClassCondition(classComb)

	// 构建省份条件
	provinceCondition := ""
	if province != "" {
		provinceCondition = fmt.Sprintf("AND province = '%s'", province)
		log.Printf("添加省份筛选条件: %s", province)
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
	SELECT COUNT(*) AS total_count FROM gaokao.admission_data 
	WHERE year = 2024 
	AND lowest_points BETWEEN ? AND ? 
	%s
	%s
	`, classCondition, provinceCondition)

	log.Printf("执行计数查询: %s", countQuery)
	var totalCountUint uint64 // 使用uint64接收COUNT()结果
	err = db.conn.QueryRow(context.Background(), countQuery, lowerScore, upperScore).Scan(&totalCountUint)
	if err != nil {
		log.Printf("计数查询失败: %v", err)
		totalCountUint = 0
	}
	totalCount := int64(totalCountUint) // 转换为int64
	log.Printf("查询到符合条件的记录总数: %d", totalCount)

	// 计算分页
	offset := (page - 1) * pageSize
	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + pageSize - 1) / pageSize
	}
	log.Printf("分页信息: 当前页=%d, 每页条数=%d, 总页数=%d, 偏移量=%d",
		page, pageSize, totalPages, offset)

	// 查询数据
	dataQuery := fmt.Sprintf(`
	SELECT id, college_code, college_name, special_interest_group_code, 
		   professional_name, class_demand, lowest_points, lowest_rank, description
	FROM gaokao.admission_data 
	WHERE year = 2024 
	AND lowest_points BETWEEN ? AND ? 
	%s
	%s
	ORDER BY lowest_points DESC
	LIMIT ? OFFSET ?
	`, classCondition, provinceCondition)

	log.Printf("执行数据查询: %s", dataQuery)
	rows, err := db.conn.Query(context.Background(), dataQuery, lowerScore, upperScore, pageSize, offset)
	if err != nil {
		log.Printf("数据查询失败: %v", err)
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
			CollegeCode:              &collegeCode,
			CollegeName:              &collegeName,
			SpecialInterestGroupCode: &groupCode,
			ProfessionalName:         profName,
			ClassDemand:              &classDemand,
			LowestPoints:             &lowestPoints,
			LowestRank:               &lowestRank,
			Description:              &desc,
		}
		list = append(list, item)
	}
	log.Printf("查询到 %d 条符合条件的记录", len(list))

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
		log.Printf("未提供选科组合，不添加选科筛选条件")
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
		log.Printf("选科组合 %s 无法识别任何有效科目，不添加选科筛选条件", classComb)
		return ""
	}

	log.Printf("选择科目: %v", subjects)

	// 构建SQL条件，选科要求包含用户选的科目或者不限
	var conditions []string

	// 添加包含所有用户选科的条件
	for _, subject := range subjects {
		conditions = append(conditions, fmt.Sprintf("class_demand LIKE '%%%s%%'", subject))
	}

	// 添加不限选项
	conditions = append(conditions, "class_demand = '不限'", "class_demand = ''")

	conditionStr := fmt.Sprintf("AND (%s)", strings.Join(conditions, " OR "))
	log.Printf("构建的选科SQL条件: %s", conditionStr)

	return conditionStr
}

// 获取数据记录数
func (db *ClickHouseDB) GetDataCount() (int64, error) {
	var count int64
	row := db.conn.QueryRow(context.Background(), "SELECT count() FROM admission_hubei_wide_2024")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 根据分数查询位次（简化版，不考虑科类和选科条件）
func (db *ClickHouseDB) QueryRankByScoreSimple(province string, year int, score float64) (int64, error) {
	// 使用新表查询，默认查询物理类
	return db.QueryRankByScoreNew(score, "物理")
}
