package routes

import (
	"gameWeb/app/controller"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine) {
	// API分组
	api := router.Group("/api")
	{
		// 游戏相关路由
		game := api.Group("/game")
		{
			game.GET("/authlist", controller.GetAuthGameList)
		}

		// 用户相关路由
		// user := api.Group("/user")
		// {
		// 	user.GET("/info", controller.GetUserInfo)
		// 	user.POST("/login", controller.UserLogin)
		// 	user.POST("/register", controller.UserRegister)
		// }

		// 其他路由可以根据需要添加
	}
}
