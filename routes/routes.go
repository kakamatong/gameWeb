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
		mail.Use(middleware.AuthMiddlewareByJWT()) // 更新中间件引用
		{
			mail.POST("/list", controller.GetClientMailList)
			mail.POST("/detail/:id", controller.GetClientMailDetail)
			mail.POST("/read/:id", controller.MarkMailAsRead)
			mail.POST("/getaward/:id", controller.GetMailAward)
		}

		// 管理后台路由组
		admin := api.Group("/admin")
		{
			// 管理员认证相关路由（无需JWT认证）
			admin.POST("/login", controller.AdminLogin)
			
			// 需要JWT认证的管理员路由
			authorized := admin.Group("/")
			authorized.Use(middleware.AdminJWTMiddleware())
			{
				// 管理员认证管理
				authorized.POST("/logout", controller.AdminLogout)
				authorized.GET("/info", controller.GetAdminInfo)
				
				// 管理员管理（仅超级管理员）
				superAdmin := authorized.Group("/")
				superAdmin.Use(middleware.RequireSuperAdmin())
				{
					superAdmin.POST("/create-admin", controller.CreateAdmin)
				}
				
				// 用户管理相关路由
				users := authorized.Group("/users")
				{
					users.GET("/", controller.GetUserList)
					users.GET("/:userid", controller.GetUserDetail)
					users.PUT("/:userid", controller.UpdateUser)
				}
				
				// 日志查询相关路由
				logs := authorized.Group("/logs")
				{
					logs.GET("/auth", controller.GetUserAuthLogs)
					logs.GET("/game", controller.GetUserGameLogs)
					logs.GET("/login-stats", controller.GetUserLoginStats)
					logs.GET("/game-stats", controller.GetUserGameStats)
				}
				
				// 系统邮件相关路由
				mails := authorized.Group("/mails")
				{
					mails.POST("/send", controller.SendSystemMail)
					mails.GET("/", controller.GetAdminMailList)
					mails.GET("/:id", controller.GetAdminMailDetail)
					mails.PUT("/:id/status", controller.UpdateMailStatus)
					mails.GET("/stats", controller.GetMailStats)
				}
			}
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
