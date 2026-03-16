package handler

import (
	"net/http"
	"ops-inspection/internal/model"
	"ops-inspection/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// List 获取项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 隐藏 Token
	for i := range projects {
		if projects[i].Token != "" {
			projects[i].Token = "**********"
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// Get 获取单个项目
func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 隐藏 Token
	if project.Token != "" {
		project.Token = "**********"
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var project model.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if project.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目名称不能为空"})
		return
	}

	if project.PrometheusURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prometheus地址不能为空"})
		return
	}

	if err := h.service.Create(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 隐藏返回的 Token
	if project.Token != "" {
		project.Token = "**********"
	}

	c.JSON(http.StatusCreated, gin.H{"data": project})
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req model.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有项目
	existing, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 更新字段
	existing.Name = req.Name
	existing.Description = req.Description
	existing.PrometheusURL = req.PrometheusURL

	// 只有当 Token 不是掩码时才更新
	if req.Token != "" && req.Token != "**********" {
		existing.Token = req.Token
	}

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 隐藏返回的 Token
	if existing.Token != "" {
		existing.Token = "**********"
	}

	c.JSON(http.StatusOK, gin.H{"data": existing})
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// TestConnection 测试项目连接
func (h *ProjectHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 这里需要 Prometheus 服务来测试连接
	c.JSON(http.StatusOK, gin.H{"message": "连接测试功能需要 Prometheus 服务支持"})
}
