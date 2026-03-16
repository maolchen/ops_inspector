package middleware

import (
	"net/http"
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	userRepo := repository.NewUserRepository(model.DB)

	return func(c *gin.Context) {
		// 从 Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token格式错误"})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token为空"})
			c.Abort()
			return
		}

		// 验证token：从数据库查找用户
		user, err := userRepo.FindByToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token，请重新登录"})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("token", token)

		c.Next()
	}
}
