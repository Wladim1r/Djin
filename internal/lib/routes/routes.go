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
	// Публичные маршруты (без авторизации)
	r.GET("/login", authController.ShowLoginForm)
	r.POST("/auth/login", authController.Login)
	r.POST("/auth/logout", authController.Logout)

	// Защищенные маршруты (требуют авторизации)
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware(), auth.RegionContextMiddleware())
	{
		// Информация о текущем пользователе
		protected.GET("/auth/me", authController.GetMe) // Новый эндпоинт

		// Маршруты для работы со статистикой
		djinRoutes := protected.Group("/djin")
		{
			djinRoutes.POST("/stat", djnHandler.PostStat)
			djinRoutes.PATCH("/stat", djnHandler.PatchStat)
			djinRoutes.GET("/stat", djnHandler.GetStatByRegion)
			djinRoutes.GET("/total", djnHandler.GetInfo)
			djinRoutes.GET("/month", djnHandler.GetStatByMonth)
		}

		// Статические страницы
		protected.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title":       "Панель управления",
				"username":    c.GetString("username"),
				"region_name": c.GetString("region_name"),
			})
		})

		protected.Static("/images", "./web/images")
		protected.StaticFile("/inputStat.html", "./web/static/inputStat.html")
		protected.StaticFile("/viewStats.html", "./web/static/viewStats.html")
		protected.StaticFile("/monthStat.html", "./web/static/monthStat.html")
	}

	// Редирект с корня на дашборд
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/dashboard")
	})
}
