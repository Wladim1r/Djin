package routes

import (
	"net/http"

	"github.com/Wladim1r/statcounter/internal/api/handler"
	"github.com/Wladim1r/statcounter/internal/auth"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authController *auth.AuthController,
	djnHandler *handler.DjnHandler,
) {
	r.Static("/images", "./web/images")
	r.Static("/static", "./web/static")

	// Публичные маршруты (без авторизации)
	r.GET("/login", authController.ShowLoginForm)
	r.POST("/auth/login", authController.Login)
	r.POST("/auth/logout", authController.Logout)

	// Защищенные маршруты (требуют авторизации)
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware(), auth.RegionContextMiddleware())
	{
		// Информация о текущем пользователе
		protected.GET("/auth/me", authController.GetMe)

		// Маршруты для работы со статистикой
		djinRoutes := protected.Group("/djin")
		{
			djinRoutes.POST("/stat", djnHandler.PostStat)
			djinRoutes.PATCH("/stat", djnHandler.PatchStat)
			djinRoutes.GET("/stat", djnHandler.GetStatByRegion)
			djinRoutes.GET("/total", djnHandler.GetInfo)
			djinRoutes.GET("/month", djnHandler.GetStatByMonth)
		}

		// Административные маршруты (только для администраторов)
		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(auth.AdminMiddleware())
		{
			// Управление пользователями
			adminRoutes.POST("/users", authController.CreateUser)
			adminRoutes.GET("/users", authController.GetAllUsers)
			adminRoutes.PUT("/users/:id", authController.UpdateUser)
			adminRoutes.DELETE("/users/:id", authController.DeleteUser)

			// Получение списка регионов
			adminRoutes.GET("/regions", authController.GetRegions)

			// Административная панель
			adminRoutes.GET("/panel", func(c *gin.Context) {
				c.HTML(http.StatusOK, "admin_panel.html", gin.H{
					"title":       "Административная панель",
					"username":    c.GetString("username"),
					"region_name": c.GetString("region_name"),
				})
			})
		}

		protected.GET("/inputStat", func(c *gin.Context) {
			c.File("./web/static/inputStat.html")
		})
		protected.GET("/viewStats", func(c *gin.Context) {
			c.File("./web/static/viewStats.html")
		})
		protected.GET("/monthStat", func(c *gin.Context) {
			c.File("./web/static/monthStat.html")
		})

		// Статические страницы
		protected.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":       "Панель управления",
				"username":    c.GetString("username"),
				"role":        c.GetString("role"),
				"region_name": c.GetString("region_name"),
			})
		})
	}

	// Редирект с корня на дашборд
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/dashboard")
	})
}
