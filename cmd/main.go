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
	"github.com/Wladim1r/statcounter/internal/db"
	"github.com/Wladim1r/statcounter/internal/lib/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	db, err := db.InitDB()
	if err != nil {
		panic(err)
	}

	repo := repository.NewDjnRepo(db)
	serv := service.NewDjnService(repo)
	hand := handler.NewDjnHandler(serv)

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.Use(gin.LoggerWithFormatter(logger.Log))

	// API routes
	router.POST("/djin/stat", hand.PostStat)
	router.GET("/djin/stat", hand.GetStats)
	router.GET("/djin/total", hand.GetInfo)

	// Static files
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/index.html", "./web/index.html")
	router.StaticFile("/inputStat.html", "./web/inputStat.html")
	router.StaticFile("/viewStats.html", "./web/viewStats.html")
	router.Static("/images", "./web/images/")

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
	}

	// Проверяем и парсим интервал для truncate
	intervalStr := os.Getenv("TRUNCATE_INTERVAL")
	if intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			log.Printf("Ошибка парсинга TRUNCATE_INTERVAL: %v", err)
		} else {
			timer := time.NewTicker(interval)
			defer timer.Stop()

			go func() {
				for {
					select {
					case <-timer.C:
						if err := repo.Truncate(); err != nil {
							log.Print(err)
						}
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

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
