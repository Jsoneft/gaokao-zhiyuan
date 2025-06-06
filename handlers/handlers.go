package handlers

import (
	"net/http"
	"strconv"

	"gaokao-zhiyuan/database"
	"gaokao-zhiyuan/models"

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

	result, err := h.db.GetRankByScore(score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
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