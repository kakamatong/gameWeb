package routes

import (
	"gameWeb/app/controller"
	"gameWeb/middleware" // 更新导入路径

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine) {
	// API分组
	api := router.Group("/api")
	{
		// 游戏相关路由 - 需要验签
		game := api.Group("/game")
		{
			game.POST("/authlist", controller.GetAuthGameList)
		}

		// 邮件相关路由 - 需要验签
		mail := api.Group("/mail")
		mail.Use(middleware.AuthMiddleware()) // 更新中间件引用
		{
			mail.POST("/list", controller.GetMailList)
			mail.POST("/detail/:id", controller.GetMailDetail)
			mail.POST("/read/:id", controller.MarkMailAsRead)
			mail.POST("/getaward/:id", controller.GetMailAward)
		}

		// 用户相关路由 - 未注释部分可以添加验签
		// user := api.Group("/user")
		// user.Use(log.AuthMiddleware("your-secret-key", 300))
		// {
		//     user.GET("/info", controller.GetUserInfo)
		//     user.POST("/login", controller.UserLogin)
		//     user.POST("/register", controller.UserRegister)
		// }
	}
}
