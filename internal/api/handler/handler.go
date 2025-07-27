package handler

import (
	"errors"
	"net/http"

	"github.com/Wladim1r/statcounter/internal/api/service"
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

func (h *DjnHandler) PostStat(c *gin.Context) {
	var obj models.Stat
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid Body Request",
		})
		return
	}

	if err := h.serv.PostStat(obj); err != nil {
		switch {
		case errors.Is(err, errs.ErrUniqueName):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "Duplicate name",
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

func (h *DjnHandler) GetStats(c *gin.Context) {
	stats, err := h.serv.GetStat()
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
	response := gin.H{
		// Основная информация
		"total_reports": summa.CurrentlyQuantity,

		// Семечка
		"total_seed_plan": summa.CurrentlyStat.SeedPlan,
		"total_seed_fact": summa.CurrentlyStat.SeedFact,
		"total_seed_dif":  summa.CurrentlyStat.SeedDif,

		// Тыква
		"total_pumpkin_plan": summa.CurrentlyStat.PumpkinPlan,
		"total_pumpkin_fact": summa.CurrentlyStat.PumpkinFact,
		"total_pumpkin_dif":  summa.CurrentlyStat.PumpkinDif,

		// Арахис
		"total_peanut_plan": summa.CurrentlyStat.PeanutPlan,
		"total_peanut_fact": summa.CurrentlyStat.PeanutFact,
		"total_peanut_dif":  summa.CurrentlyStat.PeanutDif,

		// Дополнительные метрики
		"total_akb1":         summa.CurrentlyStat.AKB1,
		"total_akb2":         summa.CurrentlyStat.AKB2,
		"total_newtt":        summa.CurrentlyStat.NewTT,
		"total_mix":          summa.CurrentlyStat.Mix,
		"total_npone":        summa.CurrentlyStat.NpOne,
		"total_set_shelving": summa.CurrentlyStat.SetShel,
		"total_dmp":          summa.CurrentlyStat.DMP,
		"total_top_five":     summa.CurrentlyStat.TopFive,
		"total_news":         summa.CurrentlyStat.News,
	}

	c.JSON(http.StatusOK, response)
}
