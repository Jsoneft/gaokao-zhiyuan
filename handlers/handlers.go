package handlers

import (
	"net/http"
	"strconv"

	"gaokao-zhiyuan/database"
	_ "gaokao-zhiyuan/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *database.ClickHouseDB
}

func NewHandler(db *database.ClickHouseDB) *Handler {
	return &Handler{db: db}
}

// 查询位次接口
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

	score, err := strconv.ParseInt(scoreStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "score参数格式错误",
		})
		return
	}

	// 使用默认参数
	province := "重庆"
	year := 2024
	subjectType := "物理"
	classDemands := []string{"物", "化", "生"}

	rank, err := h.db.QueryRankByScore(province, year, float64(score), subjectType, classDemands)
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
		"year":  year,
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
		req.Province = "重庆"
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

// 报表查询接口
// GET /api/report/get?rank=12000&class_comb="123"&page=1&page_size=20
func (h *Handler) GetReport(c *gin.Context) {
	// 获取参数
	rankStr := c.Query("rank")
	classComb := c.Query("class_comb")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

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
		pageSize = 20
	}

	result, err := h.db.GetReportData(rank, classComb, page, pageSize)
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
