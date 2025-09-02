package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wladim1r/statcounter/internal/lib/errs"
	"github.com/Wladim1r/statcounter/internal/lib/summa"
	"github.com/Wladim1r/statcounter/internal/models"
	"gorm.io/gorm"
)

type DjnRepo interface {
	PostStat(stat *models.StatDaily) error
	PatchStat(regionID uint, stat *models.StatDaily) error
	GetStatsByRegion(regionID uint) ([]models.StatDaily, error)
	GetStatsByRegionAndUser(regionID uint, username string) ([]models.StatDaily, error)
	DeleteOlderThan(cutoffDate time.Time) error
	GetStatsByMonth(regionID uint, date string) ([]models.StatDaily, error)
	GetStatsByMonthAndUser(regionID uint, username string, date string) ([]models.StatDaily, error)
}

type djnRepo struct {
	db *gorm.DB
}

func NewDjnRepo(db *gorm.DB) DjnRepo {
	return &djnRepo{db: db}
}

func (r *djnRepo) DeleteOlderThan(cutoffDate time.Time) error {
	if err := r.db.Exec("DELETE FROM stat_dailies WHERE date < $1", cutoffDate).Error; err != nil {
		return fmt.Errorf("%w: %v", errs.ErrDBOperation, err)
	}

	summa.ClearAllStats()

	return nil
}

func (r *djnRepo) GetStatsByMonth(regionID uint, date string) ([]models.StatDaily, error) {
	var stats []models.StatDaily
	result := r.db.Where("region_id = ? AND date = ?", regionID, date).Find(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w", errs.ErrNotFound)
	}

	return stats, nil
}

func (r *djnRepo) GetStatsByMonthAndUser(
	regionID uint,
	username string,
	date string,
) ([]models.StatDaily, error) {
	var stats []models.StatDaily
	result := r.db.Where("region_id = ? AND name = ? AND date = ?", regionID, username, date).
		Find(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w", errs.ErrNotFound)
	}

	return stats, nil
}

func (r *djnRepo) PatchStat(regionID uint, stat *models.StatDaily) error {
	// Сначала получаем старые данные
	var oldStat models.StatDaily
	if err := r.db.Where("region_id = ? AND name = ?", regionID, stat.Name).First(&oldStat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("%w: failed to get old record %v", errs.ErrDBOperation, err)
	}

	// Обновляем запись
	result := r.db.Where("region_id = ? AND name = ?", regionID, stat.Name).Updates(stat)
	if result.Error != nil {
		return fmt.Errorf("%w: failed update %v", errs.ErrDBOperation, result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}

	// Обновляем summa
	summa.UpdateStatForRegion(regionID, oldStat, *stat)

	return nil
}

func (r *djnRepo) PostStat(stat *models.StatDaily) error {
	result := r.db.Create(stat)
	if result.Error != nil {
		// Проверяем различные типы ошибок уникальности
		errorStr := strings.ToLower(result.Error.Error())
		if strings.Contains(errorStr, "duplicate key") ||
			strings.Contains(errorStr, "unique constraint") ||
			strings.Contains(errorStr, "idx_unique_daily_stat") {
			return fmt.Errorf("%w: %s", errs.ErrUniqueName, "You have already submitted data for today")
		}
		return fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}

	summa.AddStatForRegion(stat.RegionID, *stat)
	return nil
}

func (r *djnRepo) GetStatsByRegion(regionID uint) ([]models.StatDaily, error) {
	var stats []models.StatDaily
	today := time.Now().Format("2006-01-02") // YYYY-MM-DD

	result := r.db.Where("region_id = ? AND date = ?", regionID, today).Find(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w", errs.ErrNotFound)
	}

	return stats, nil
}

func (r *djnRepo) GetStatsByRegionAndUser(
	regionID uint,
	username string,
) ([]models.StatDaily, error) {
	var stats []models.StatDaily
	today := time.Now().Format("2006-01-02") // YYYY-MM-DD

	result := r.db.Where("region_id = ? AND name = ? AND date = ?", regionID, username, today).
		Find(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w", errs.ErrNotFound)
	}

	return stats, nil
}
