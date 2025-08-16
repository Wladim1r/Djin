package service

import (
	"fmt"
	"time"

	"github.com/Wladim1r/statcounter/internal/api/repository"
	"github.com/Wladim1r/statcounter/internal/lib/errs"
	"github.com/Wladim1r/statcounter/internal/models"
)

type DjnService interface {
	PostStat(stat models.StatDaily) error
	PatchStat(regionID uint, stat models.StatDaily) error
	GetStatByRegion(regionID uint) ([]models.StatDaily, error)
	GetStatByRegionAndUser(regionID uint, username string) ([]models.StatDaily, error)
	GetStatsByMonth(regionID uint, date string) ([]models.StatDaily, error)
	GetStatsByMonthAndUser(regionID uint, username string, date string) ([]models.StatDaily, error)
}

type djnService struct {
	repo repository.DjnRepo
}

func NewDjnService(repo repository.DjnRepo) DjnService {
	return &djnService{repo: repo}
}

func (s *djnService) GetStatsByMonth(regionID uint, date string) ([]models.StatDaily, error) {
	if date == "" {
		return nil, fmt.Errorf(
			"%w: %v",
			errs.ErrBadRequest,
			"Date parameter is required (format: YYYY-MM-DD)",
		)
	}

	// Валидируем формат даты
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Проверяем, что дата не старше 30 дней
	requestedDate, _ := time.Parse("2006-01-02", date)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if requestedDate.Before(thirtyDaysAgo) {
		return nil, fmt.Errorf("%w: %v", errs.ErrBadRequest, "Requested date is older than 30 days")
	}

	return s.repo.GetStatsByMonth(regionID, date)
}

func (s *djnService) PatchStat(regionID uint, stat models.StatDaily) error {

	seedDif := stat.SeedFact - stat.SeedPlan
	pumpkinDif := stat.PumpkinFact - stat.PumpkinPlan
	peanutDif := stat.PeanutFact - stat.PeanutPlan

	stat.SeedDif = seedDif
	stat.PumpkinDif = pumpkinDif
	stat.PeanutDif = peanutDif

	return s.repo.PatchStat(regionID, &stat)
}

func (s *djnService) PostStat(stat models.StatDaily) error {

	seedDif := stat.SeedFact - stat.SeedPlan
	pumpkinDif := stat.PumpkinFact - stat.PumpkinPlan
	peanutDif := stat.PeanutFact - stat.PeanutPlan

	stat.SeedDif = seedDif
	stat.PumpkinDif = pumpkinDif
	stat.PeanutDif = peanutDif

	return s.repo.PostStat(&stat)
}

func (s *djnService) GetStatByRegion(regionID uint) ([]models.StatDaily, error) {
	return s.repo.GetStatsByRegion(regionID)
}

func (s *djnService) GetStatByRegionAndUser(
	regionID uint,
	username string,
) ([]models.StatDaily, error) {
	return s.repo.GetStatsByRegionAndUser(regionID, username)
}

func (s *djnService) GetStatsByMonthAndUser(
	regionID uint,
	username string,
	date string,
) ([]models.StatDaily, error) {
	return s.repo.GetStatsByMonthAndUser(regionID, username, date)
}
