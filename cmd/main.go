package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wladim1r/statcounter/internal/api/handler"
	"github.com/Wladim1r/statcounter/internal/api/repository"
	"github.com/Wladim1r/statcounter/internal/api/service"
	"github.com/Wladim1r/statcounter/internal/auth"
	"github.com/Wladim1r/statcounter/internal/db"
	"github.com/Wladim1r/statcounter/internal/lib/logger"
	"github.com/Wladim1r/statcounter/internal/lib/routes"
	"github.com/Wladim1r/statcounter/internal/lib/tick"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	db, err := db.InitDB()
	if err != nil {
		panic(err)
	}

	// Инициализируем регионы и пользователей
	if err := auth.InitializeRegions(db); err != nil {
		log.Printf("Error initialize regions and users: %v", err)
	}

	// initialize services
	repo := repository.NewDjnRepo(db)
	serv := service.NewDjnService(repo)
	hand := handler.NewDjnHandler(serv)

	// initialize authorization
	authService := auth.NewAuthService(db)
	authController := auth.NewAuthController(authService)

	router := gin.Default()

	router.LoadHTMLGlob("web/templates/*.html")

	// session settings
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	})
	router.Use(sessions.Sessions("session", store))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "https://dzhinnkrk.ru"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Добавляем заголовки безопасности
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.Use(gin.LoggerWithFormatter(logger.Log))

	routes.SetupRoutes(router, authController, &hand)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	svr := &http.Server{
		Addr:    os.Getenv("SVR_PORT"),
		Handler: router,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go tick.TruncateToTickerMonthlyWithContext(ctx, repo)

	log.Printf("server start on potr %s\n", os.Getenv("SVR_PORT"))
	log.Printf("database info configuration\n")
	log.Printf("host %s\nuser %s\npassword %s\nname %s\nport %s\n",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	go func() {
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка сервера: %v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.Shutdown(shutdownCtx); err != nil {
		fmt.Println("failed to shutdown", err)
	} else {
		fmt.Println("server stopped gracefully")
	}
}
