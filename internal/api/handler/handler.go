package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Wladim1r/statcounter/internal/api/service"
	"github.com/Wladim1r/statcounter/internal/auth"
	"github.com/Wladim1r/statcounter/internal/lib/errs"
	"github.com/Wladim1r/statcounter/internal/lib/summa"
	"github.com/Wladim1r/statcounter/internal/models"
	"github.com/gin-gonic/gin"
)

type DjnHandler struct {
	serv service.DjnService
}

func NewDjnHandler(serv service.DjnService) DjnHandler {
	return DjnHandler{serv: serv}
}

func (h *DjnHandler) GetStatByMonth(c *gin.Context) {
	regionID := auth.GetRegionIDFromContext(c)
	username := c.GetString("username")

	date := c.Query("date")
	userOnly := c.Query("user") == "true"

	var (
		stats []models.StatDaily
		err   error
	)

	if userOnly {
		stats, err = h.serv.GetStatsByMonthAndUser(regionID, username, date)
	} else {
		stats, err = h.serv.GetStatsByMonth(regionID, date)
	}

	if err != nil {
		switch {
		case errors.Is(err, errs.ErrNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Record not found",
			})
		case errors.Is(err, errs.ErrDBOperation):
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database operation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *DjnHandler) PatchStat(c *gin.Context) {
	regionID := auth.GetRegionIDFromContext(c)
	if regionID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Region is not defined",
		})
		return
	}

	// Получаем username из контекста
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var obj models.StatDaily
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid Body Request",
		})
		return
	}

	// Проверяем, что пользователь пытается изменить только свои данные
	if obj.Name != username {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You can only edit your own statistics",
		})
		return
	}

	log.Printf("Received data for region %d, name %s: %+v", regionID, obj.Name, obj)

	if err := h.serv.PatchStat(regionID, obj); err != nil {
		switch {
		case errors.Is(err, errs.ErrNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Record not found",
			})
		case errors.Is(err, errs.ErrDBOperation):
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database operation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
		"status":  "success",
	})
}

func (h *DjnHandler) PostStat(c *gin.Context) {
	regionID := auth.GetRegionIDFromContext(c)
	if regionID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Region is not defined",
		})
		return
	}

	// Получаем username из контекста
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var obj models.StatDaily
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid Body Request",
		})
		return
	}

	// Автоматически устанавливаем имя пользователя из сессии
	obj.Name = username
	obj.RegionID = regionID
	obj.Date = time.Now().Format("2006-01-02")

	if err := h.serv.PostStat(obj); err != nil {
		switch {
		case errors.Is(err, errs.ErrUniqueName):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "You have already submitted data for today",
			})
		case errors.Is(err, errs.ErrDBOperation):
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database operation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data saved successfully",
		"status":  "success",
	})
}

func (h *DjnHandler) GetStatByRegion(c *gin.Context) {
	regionID := auth.GetRegionIDFromContext(c)
	if regionID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Region is not defined",
		})
		return
	}

	// Получаем username из контекста для фильтрации
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	// Используем новый метод для получения статистики конкретного пользователя
	stats, err := h.serv.GetStatByRegionAndUser(regionID, username)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Could not found records",
			})
		case errors.Is(err, errs.ErrDBOperation):
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database operation failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Internal server error",
			})
		}

		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *DjnHandler) GetInfo(c *gin.Context) {
	regionID := auth.GetRegionIDFromContext(c)
	if regionID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Region is not defined",
		})
		return
	}

	regionStat, regionQuantity := summa.GetStatsForRegion(regionID)

	response := gin.H{
		// Основная информация
		"total_reports": regionQuantity,
		"region_id":     regionID,

		// Семечка
		"total_seed_plan": regionStat.SeedPlan,
		"total_seed_fact": regionStat.SeedFact,
		"total_seed_dif":  regionStat.SeedDif,

		// Тыква
		"total_pumpkin_plan": regionStat.PumpkinPlan,
		"total_pumpkin_fact": regionStat.PumpkinFact,
		"total_pumpkin_dif":  regionStat.PumpkinDif,

		// Арахис
		"total_peanut_plan": regionStat.PeanutPlan,
		"total_peanut_fact": regionStat.PeanutFact,
		"total_peanut_dif":  regionStat.PeanutDif,

		// Дополнительные метрики
		"total_akb1":         regionStat.AKB1,
		"total_akb2":         regionStat.AKB2,
		"total_newtt":        regionStat.NewTT,
		"total_mix":          regionStat.Mix,
		"total_npone":        regionStat.NpOne,
		"total_set_shelving": regionStat.SetShel,
		"total_dmp":          regionStat.DMP,
		"total_top_five":     regionStat.TopFive,
		"total_news":         regionStat.News,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DjnHandler) GetAllRegionalStats(c *gin.Context) {
	allStats := summa.GetAllRegionalStats()
	allQuantities := summa.GetAllQuantities()

	response := make(map[string]interface{})

	for regionID, stat := range allStats {
		quantity := allQuantities[regionID]
		response[fmt.Sprintf("region_%d", regionID)] = gin.H{
			"region_id":     regionID,
			"total_reports": quantity,
			"seed_plan":     stat.SeedPlan,
			"seed_fact":     stat.SeedFact,
			"seed_dif":      stat.SeedDif,
			"pumpkin_plan":  stat.PumpkinPlan,
			"pumpkin_fact":  stat.PumpkinFact,
			"pumpkin_dif":   stat.PumpkinDif,
			"peanut_plan":   stat.PeanutPlan,
			"peanut_fact":   stat.PeanutFact,
			"peanut_dif":    stat.PeanutDif,
			"akb1":          stat.AKB1,
			"akb2":          stat.AKB2,
			"newtt":         stat.NewTT,
			"mix":           stat.Mix,
			"npone":         stat.NpOne,
			"set_shelving":  stat.SetShel,
			"dmp":           stat.DMP,
			"top_five":      stat.TopFive,
			"news":          stat.News,
		}
	}

	c.JSON(http.StatusOK, response)
}
