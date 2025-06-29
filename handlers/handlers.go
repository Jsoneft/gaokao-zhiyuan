package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"gaokao-zhiyuan/database"
	_ "gaokao-zhiyuan/models"

	"log"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *database.ClickHouseDB
}

func NewHandler(db *database.ClickHouseDB) *Handler {
	return &Handler{db: db}
}

// 查询位次接口 - 使用新的数据源
// GET /api/rank/get?score=555
func (h *Handler) GetRank(c *gin.Context) {
	scoreStr := c.Query("score")
	if scoreStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "缺少score参数",
		})
		return
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "score参数格式错误",
		})
		return
	}

	// 获取科目类别参数，默认为物理
	subjectCategory := c.DefaultQuery("subject_category", "物理")

	// 使用新的查询方法
	rank, err := h.db.QueryRankByScoreNew(score, subjectCategory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "success",
		"rank":  rank,
		"year":  2024,
		"score": score,
	})
}

// 高级查询位次接口
// POST /api/v1/query_rank
func (h *Handler) QueryRank(c *gin.Context) {
	var req struct {
		Province    string   `json:"province"`
		Year        int      `json:"year"`
		Score       int64    `json:"score"`
		SubjectType string   `json:"subject_type"`
		ClassDemand []string `json:"class_demand"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 参数验证
	if req.Province == "" {
		req.Province = "湖北"
	}
	if req.Year == 0 {
		req.Year = 2024
	}
	if req.SubjectType == "" {
		req.SubjectType = "物理"
	}
	if len(req.ClassDemand) == 0 {
		req.ClassDemand = []string{"物", "化", "生"}
	}

	rank, err := h.db.QueryRankByScore(req.Province, req.Year, float64(req.Score), req.SubjectType, req.ClassDemand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         0,
		"msg":          "success",
		"rank":         rank,
		"year":         req.Year,
		"province":     req.Province,
		"subject_type": req.SubjectType,
		"score":        req.Score,
	})
}

// 报表查询接口 - 新版本
// GET /api/report/get?rank=333&class_first_choise=物理&class_optional_choise=["化学","生物"]&province=湖北&page=1&page_size=10&college_location=["湖北"]&interest=["理科","工科"]&strategy=0&fuzzy_subject_category=物理
func (h *Handler) GetReport(c *gin.Context) {
	// 获取参数
	rankStr := c.Query("rank")
	classFirstChoice := c.Query("class_first_choise")
	classOptionalChoiceStr := c.Query("class_optional_choise")
	province := c.Query("province")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	collegeLocationStr := c.Query("college_location")
	interestStr := c.Query("interest")
	strategyStr := c.DefaultQuery("strategy", "0")
	fuzzySubjectCategory := c.Query("fuzzy_subject_category")

	// 参数验证
	if rankStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "缺少rank参数",
		})
		return
	}

	rank, err := strconv.ParseInt(rankStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "rank参数格式错误",
		})
		return
	}

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	strategy, err := strconv.Atoi(strategyStr)
	if err != nil {
		strategy = 0
	}

	// SQL注入防护：fuzzy_subject_category参数校验
	if fuzzySubjectCategory != "" {
		// 只允许字母、数字、中文和基本标点符号，防止SQL注入
		validPattern := regexp.MustCompile(`^[a-zA-Z0-9\p{Han}\s\-_()（）]+$`)
		if !validPattern.MatchString(fuzzySubjectCategory) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 1,
				"msg":  "fuzzy_subject_category参数包含非法字符",
			})
			return
		}
		// 限制参数长度，防止过长的输入
		if len(fuzzySubjectCategory) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 1,
				"msg":  "fuzzy_subject_category参数长度不能超过50个字符",
			})
			return
		}
	}

	// 解析JSON数组参数
	var classOptionalChoice []string
	if classOptionalChoiceStr != "" {
		if err := json.Unmarshal([]byte(classOptionalChoiceStr), &classOptionalChoice); err != nil {
			log.Printf("解析class_optional_choise参数失败: %v", err)
			classOptionalChoice = []string{}
		}
	}

	var collegeLocation []string
	if collegeLocationStr != "" {
		if err := json.Unmarshal([]byte(collegeLocationStr), &collegeLocation); err != nil {
			log.Printf("解析college_location参数失败: %v", err)
			collegeLocation = []string{}
		}
	}

	var interest []string
	if interestStr != "" {
		if err := json.Unmarshal([]byte(interestStr), &interest); err != nil {
			log.Printf("解析interest参数失败: %v", err)
			interest = []string{}
		}
	}

	log.Printf("报表查询请求: rank=%d, classFirstChoice=%s, classOptionalChoice=%v, province=%s, page=%d, pageSize=%d, collegeLocation=%v, interest=%v, strategy=%d, fuzzySubjectCategory=%s",
		rank, classFirstChoice, classOptionalChoice, province, page, pageSize, collegeLocation, interest, strategy, fuzzySubjectCategory)

	// 使用新的查询方法，传递fuzzy_subject_category参数
	result, err := h.db.GetReportDataNew(rank, classFirstChoice, classOptionalChoice, province, page, pageSize, collegeLocation, interest, strategy, fuzzySubjectCategory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"msg":    "高考志愿填报辅助系统后端服务运行正常",
	})
}
