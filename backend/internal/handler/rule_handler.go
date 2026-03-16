package handler

import (
	"net/http"
	"ops-inspection/internal/model"
	"ops-inspection/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	service     *service.RuleService
	promService *service.PrometheusService
}

func NewRuleHandler(service *service.RuleService, promService *service.PrometheusService) *RuleHandler {
	return &RuleHandler{service: service, promService: promService}
}

// List 获取规则列表
func (h *RuleHandler) List(c *gin.Context) {
	// 支持按分组过滤
	groupIDStr := c.Query("group_id")

	if groupIDStr != "" {
		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组ID"})
			return
		}

		rules, err := h.service.GetByGroupID(uint(groupID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": rules})
		return
	}

	// 返回所有规则
	rules, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// Get 获取单个规则
func (h *RuleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// Create 创建规则
func (h *RuleHandler) Create(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rule.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "规则名称不能为空"})
		return
	}

	if rule.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PromQL查询不能为空"})
		return
	}

	if err := h.service.Create(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

// Update 更新规则
func (h *RuleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	var req model.Rule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	// 更新字段
	existing.GroupID = req.GroupID
	existing.Name = req.Name
	existing.Type = req.Type
	existing.ShowInTable = req.ShowInTable
	existing.Description = req.Description
	existing.Query = req.Query
	existing.TrendQuery = req.TrendQuery
	existing.Threshold = req.Threshold
	existing.Unit = req.Unit
	existing.Labels = req.Labels
	existing.ThresholdType = req.ThresholdType
	existing.ProjectScope = req.ProjectScope
	existing.Enabled = req.Enabled
	existing.SortOrder = req.SortOrder

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existing})
}

// Delete 删除规则
func (h *RuleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// Toggle 切换规则启用状态
func (h *RuleHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	if err := h.service.ToggleEnabled(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

// Test 测试规则（验证 PromQL）
func (h *RuleHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	// 获取项目 ID（用于测试）
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少项目ID参数"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "测试功能需要完整的项目服务支持",
		"rule":    rule,
	})
}
