package repository

import (
	"fmt"
	"strings"

	"github.com/Wladim1r/statcounter/internal/lib/errs"
	"github.com/Wladim1r/statcounter/internal/lib/summa"
	"github.com/Wladim1r/statcounter/internal/models"
	"gorm.io/gorm"
)

type DjnRepo interface {
	Truncate() error
	PostStat(stat *models.Stat) error
	GetStats() ([]models.Stat, error)
}

type djnRepo struct {
	db *gorm.DB
}

func NewDjnRepo(db *gorm.DB) DjnRepo {
	return &djnRepo{db: db}
}

func (r *djnRepo) Truncate() error {
	if err := r.db.Exec("TRUNCATE TABLE stats RESTART IDENTITY CASCADE;").Error; err != nil {
		return fmt.Errorf("%w: %v", errs.ErrDBOperation, err)
	}

	summa.CurrentlyStat = models.Stat{}
	summa.CurrentlyQuantity = 0

	return nil
}

func (r *djnRepo) PostStat(stat *models.Stat) error {
	result := r.db.Create(stat)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			return fmt.Errorf("%w: %s", errs.ErrUniqueName, result.Error)
		}
		return fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}

	summa.Sum(*stat)
	summa.CurrentlyQuantity++

	return nil
}

func (r *djnRepo) GetStats() ([]models.Stat, error) {
	var stats []models.Stat

	result := r.db.Find(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDBOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w", errs.ErrNotFound)
	}

	return stats, nil
}
