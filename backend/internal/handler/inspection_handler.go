package handler

import (
	"net/http"
	"ops-inspection/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InspectionHandler struct {
	service *service.InspectionService
}

func NewInspectionHandler(service *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{service: service}
}

// Start 启动巡检
func (h *InspectionHandler) Start(c *gin.Context) {
	var req struct {
		ProjectID uint   `json:"project_id" binding:"required"`
		Inspector string `json:"inspector"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.StartInspection(req.ProjectID, req.Inspector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

// Get 获取巡检报告详情
func (h *InspectionHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	report, items, err := h.service.GetReportByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"report": report,
			"items":  items,
		},
	})
}

// List 获取巡检报告列表
func (h *InspectionHandler) List(c *gin.Context) {
	reports, err := h.service.GetAllReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reports})
}

// UpdateSummary 更新巡检总结
func (h *InspectionHandler) UpdateSummary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	var req struct {
		Summary string `json:"summary"`
		Remark  string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSummary(uint(id), req.Summary, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
