package handler

import (
	"net/http"
	"ops-inspection/internal/model"
	"ops-inspection/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	inspectionService *service.InspectionService
}

func NewSystemHandler(inspectionService *service.InspectionService) *SystemHandler {
	return &SystemHandler{inspectionService: inspectionService}
}

// GetConfigs 获取所有系统配置
func (h *SystemHandler) GetConfigs(c *gin.Context) {
	var configs []model.SystemConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为 map 格式方便前端使用
	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateConfig 更新系统配置
func (h *SystemHandler) UpdateConfig(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := model.SetConfigValue(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// CleanupReports 手动清理过期报告
func (h *SystemHandler) CleanupReports(c *gin.Context) {
	daysStr := c.Query("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	count, err := h.inspectionService.CleanupOldReports(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "清理完成",
		"count":   count,
	})
}
