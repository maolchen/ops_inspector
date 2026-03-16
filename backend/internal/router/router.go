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
	userRepo := repository.NewUserRepository(model.DB)

	// 初始化 Service
	projectService := service.NewProjectService(projectRepo)
	ruleGroupService := service.NewRuleGroupService(ruleGroupRepo)
	prometheusService := service.NewPrometheusService()
	ruleService := service.NewRuleService(ruleRepo, ruleGroupRepo)
	inspectionService := service.NewInspectionService(reportRepo, ruleRepo, projectRepo, prometheusService)
	authService := service.NewAuthService(userRepo)

	// 初始化 Handler
	projectHandler := handler.NewProjectHandler(projectService)
	ruleGroupHandler := handler.NewRuleGroupHandler(ruleGroupService)
	ruleHandler := handler.NewRuleHandler(ruleService, prometheusService)
	inspectionHandler := handler.NewInspectionHandler(inspectionService)
	authHandler := handler.NewAuthHandler(authService)

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由（无需登录）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// 认证相关
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/user", authHandler.GetCurrentUser)
			protected.PUT("/auth/password", authHandler.ChangePassword)

			// 项目管理
			protected.GET("/projects", projectHandler.List)
			protected.GET("/projects/:id", projectHandler.Get)
			protected.POST("/projects", projectHandler.Create)
			protected.PUT("/projects/:id", projectHandler.Update)
			protected.DELETE("/projects/:id", projectHandler.Delete)
			protected.POST("/projects/:id/test", projectHandler.TestConnection)

			// 规则组管理
			protected.GET("/rule-groups", ruleGroupHandler.List)
			protected.GET("/rule-groups/:id", ruleGroupHandler.Get)
			protected.POST("/rule-groups", ruleGroupHandler.Create)
			protected.PUT("/rule-groups/:id", ruleGroupHandler.Update)
			protected.DELETE("/rule-groups/:id", ruleGroupHandler.Delete)

			// 规则管理
			protected.GET("/rules", ruleHandler.List)
			protected.GET("/rules/:id", ruleHandler.Get)
			protected.POST("/rules", ruleHandler.Create)
			protected.PUT("/rules/:id", ruleHandler.Update)
			protected.DELETE("/rules/:id", ruleHandler.Delete)
			protected.POST("/rules/:id/toggle", ruleHandler.Toggle)
			protected.POST("/rules/:id/test", ruleHandler.Test)

			// 巡检管理
			protected.GET("/inspections", inspectionHandler.List)
			protected.GET("/inspections/:id", inspectionHandler.Get)
			protected.POST("/inspections/start", inspectionHandler.Start)
			protected.PUT("/inspections/:id/summary", inspectionHandler.UpdateSummary)
		}
	}

	return r
}
