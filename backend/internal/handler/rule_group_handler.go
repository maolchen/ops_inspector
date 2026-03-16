package handler

import (
	"net/http"
	"ops-inspection/internal/model"
	"ops-inspection/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleGroupHandler struct {
	service *service.RuleGroupService
}

func NewRuleGroupHandler(service *service.RuleGroupService) *RuleGroupHandler {
	return &RuleGroupHandler{service: service}
}

// List 获取规则组列表
func (h *RuleGroupHandler) List(c *gin.Context) {
	groups, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": groups})
}

// Get 获取单个规则组
func (h *RuleGroupHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则组ID"})
		return
	}

	group, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则组不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": group})
}

// Create 创建规则组
func (h *RuleGroupHandler) Create(c *gin.Context) {
	var group model.RuleGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if group.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "规则组名称不能为空"})
		return
	}

	if group.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "规则组标识不能为空"})
		return
	}

	if err := h.service.Create(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": group})
}

// Update 更新规则组
func (h *RuleGroupHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则组ID"})
		return
	}

	var req model.RuleGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则组不存在"})
		return
	}

	existing.Name = req.Name
	existing.Code = req.Code
	existing.Description = req.Description
	existing.SortOrder = req.SortOrder

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existing})
}

// Delete 删除规则组
func (h *RuleGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则组ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法删除，规则组下存在关联规则"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
