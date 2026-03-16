package router

import (
	"ops-inspection/internal/handler"
	"ops-inspection/internal/middleware"
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
	"ops-inspection/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 使用 CORS 中间件
	r.Use(middleware.CORS())

	// 初始化 Repository
	projectRepo := repository.NewProjectRepository(model.DB)
	ruleGroupRepo := repository.NewRuleGroupRepository(model.DB)
	ruleRepo := repository.NewRuleRepository(model.DB)
	reportRepo := repository.NewReportRepository(model.DB)

	// 初始化 Service
	projectService := service.NewProjectService(projectRepo)
	ruleGroupService := service.NewRuleGroupService(ruleGroupRepo)
	prometheusService := service.NewPrometheusService()
	ruleService := service.NewRuleService(ruleRepo, ruleGroupRepo)
	inspectionService := service.NewInspectionService(reportRepo, ruleRepo, projectRepo, prometheusService)

	// 初始化 Handler
	projectHandler := handler.NewProjectHandler(projectService)
	ruleGroupHandler := handler.NewRuleGroupHandler(ruleGroupService)
	ruleHandler := handler.NewRuleHandler(ruleService, prometheusService)
	inspectionHandler := handler.NewInspectionHandler(inspectionService)

	// API 路由组
	api := r.Group("/api")
	{
		// 项目管理
		projects := api.Group("/projects")
		{
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.Get)
			projects.POST("", projectHandler.Create)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
			projects.POST("/:id/test", projectHandler.TestConnection)
		}

		// 规则组管理
		ruleGroups := api.Group("/rule-groups")
		{
			ruleGroups.GET("", ruleGroupHandler.List)
			ruleGroups.GET("/:id", ruleGroupHandler.Get)
			ruleGroups.POST("", ruleGroupHandler.Create)
			ruleGroups.PUT("/:id", ruleGroupHandler.Update)
			ruleGroups.DELETE("/:id", ruleGroupHandler.Delete)
		}

		// 规则管理
		rules := api.Group("/rules")
		{
			rules.GET("", ruleHandler.List)
			rules.GET("/:id", ruleHandler.Get)
			rules.POST("", ruleHandler.Create)
			rules.PUT("/:id", ruleHandler.Update)
			rules.DELETE("/:id", ruleHandler.Delete)
			rules.POST("/:id/toggle", ruleHandler.Toggle)
			rules.POST("/:id/test", ruleHandler.Test)
		}

		// 巡检管理
		inspections := api.Group("/inspections")
		{
			inspections.GET("", inspectionHandler.List)
			inspections.GET("/:id", inspectionHandler.Get)
			inspections.POST("/start", inspectionHandler.Start)
			inspections.PUT("/:id/summary", inspectionHandler.UpdateSummary)
		}
	}

	return r
}
