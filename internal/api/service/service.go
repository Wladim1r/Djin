package service

import (
	"github.com/Wladim1r/statcounter/internal/api/repository"
	"github.com/Wladim1r/statcounter/internal/models"
)

type DjnService interface {
	PostStat(stat models.Stat) error
	GetStat() ([]models.Stat, error)
}

type djnService struct {
	repo repository.DjnRepo
}

func NewDjnService(repo repository.DjnRepo) DjnService {
	return &djnService{repo: repo}
}

func (s *djnService) PostStat(stat models.Stat) error {

	seedDif := stat.SeedFact - stat.SeedPlan
	pumpkinDif := stat.PumpkinFact - stat.PumpkinPlan
	peanutDif := stat.PeanutFact - stat.PeanutPlan

	stat.SeedDif = seedDif
	stat.PumpkinDif = pumpkinDif
	stat.PeanutDif = peanutDif

	return s.repo.PostStat(&stat)
}

func (s *djnService) GetStat() ([]models.Stat, error) {
	return s.repo.GetStats()
}
