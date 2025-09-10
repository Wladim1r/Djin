package summa

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Wladim1r/statcounter/internal/models"
)

// Структура для хранения статистики по регионам
type RegionalStats struct {
	mu         sync.RWMutex
	stats      map[uint]models.StatDaily // regionID -> aggregated stats
	quantities map[uint]int              // regionID -> quantity of reports
}

var (
	regionalStats = &RegionalStats{
		stats:      make(map[uint]models.StatDaily),
		quantities: make(map[uint]int),
	}
)

func UpdateStatForRegion(regionID uint, oldStat, newStat models.StatDaily) {
	regionalStats.mu.Lock()
	defer regionalStats.mu.Unlock()

	fmt.Printf("Updating stat for region %d: old=%+v, new=%+v\n", regionID, oldStat, newStat)

	currentStat := regionalStats.stats[regionID]

	// Вычитаем старые значения
	currentStat.SeedPlan = roundFloat(currentStat.SeedPlan-oldStat.SeedPlan, 2)
	currentStat.SeedFact = roundFloat(currentStat.SeedFact-oldStat.SeedFact, 2)
	currentStat.SeedDif = roundFloat(currentStat.SeedDif-oldStat.SeedDif, 2)
	currentStat.PumpkinPlan = roundFloat(currentStat.PumpkinPlan-oldStat.PumpkinPlan, 2)
	currentStat.PumpkinFact = roundFloat(currentStat.PumpkinFact-oldStat.PumpkinFact, 2)
	currentStat.PumpkinDif = roundFloat(currentStat.PumpkinDif-oldStat.PumpkinDif, 2)
	currentStat.PeanutPlan = roundFloat(currentStat.PeanutPlan-oldStat.PeanutPlan, 2)
	currentStat.PeanutFact = roundFloat(currentStat.PeanutFact-oldStat.PeanutFact, 2)
	currentStat.PeanutDif = roundFloat(currentStat.PeanutDif-oldStat.PeanutDif, 2)
	currentStat.AKB1 -= oldStat.AKB1
	currentStat.AKB2 -= oldStat.AKB2
	currentStat.NewTT -= oldStat.NewTT
	currentStat.Mix -= oldStat.Mix
	currentStat.NpOne -= oldStat.NpOne
	currentStat.SetShel -= oldStat.SetShel
	currentStat.DMP -= oldStat.DMP
	currentStat.TopFive -= oldStat.TopFive
	currentStat.News -= oldStat.News

	// Добавляем новые значения
	currentStat.SeedPlan = roundFloat(currentStat.SeedPlan+newStat.SeedPlan, 2)
	currentStat.SeedFact = roundFloat(currentStat.SeedFact+newStat.SeedFact, 2)
	currentStat.SeedDif = roundFloat(currentStat.SeedDif+newStat.SeedDif, 2)
	currentStat.PumpkinPlan = roundFloat(currentStat.PumpkinPlan+newStat.PumpkinPlan, 2)
	currentStat.PumpkinFact = roundFloat(currentStat.PumpkinFact+newStat.PumpkinFact, 2)
	currentStat.PumpkinDif = roundFloat(currentStat.PumpkinDif+newStat.PumpkinDif, 2)
	currentStat.PeanutPlan = roundFloat(currentStat.PeanutPlan+newStat.PeanutPlan, 2)
	currentStat.PeanutFact = roundFloat(currentStat.PeanutFact+newStat.PeanutFact, 2)
	currentStat.PeanutDif = roundFloat(currentStat.PeanutDif+newStat.PeanutDif, 2)
	currentStat.AKB1 += newStat.AKB1
	currentStat.AKB2 += newStat.AKB2
	currentStat.NewTT += newStat.NewTT
	currentStat.Mix += newStat.Mix
	currentStat.NpOne += newStat.NpOne
	currentStat.SetShel += newStat.SetShel
	currentStat.DMP += newStat.DMP
	currentStat.TopFive += newStat.TopFive
	currentStat.News += newStat.News

	regionalStats.stats[regionID] = currentStat
}

// Добавить статистику для конкретного региона
func AddStatForRegion(regionID uint, stat models.StatDaily) {
	regionalStats.mu.Lock()
	defer regionalStats.mu.Unlock()

	fmt.Printf("Adding stat for region %d: %+v\n", regionID, stat)

	currentStat := regionalStats.stats[regionID]

	// Суммируем статистику
	currentStat.SeedPlan = roundFloat(currentStat.SeedPlan+stat.SeedPlan, 2)
	currentStat.SeedFact = roundFloat(currentStat.SeedFact+stat.SeedFact, 2)
	currentStat.SeedDif = roundFloat(currentStat.SeedDif+stat.SeedDif, 2)
	currentStat.PumpkinPlan = roundFloat(currentStat.PumpkinPlan+stat.PumpkinPlan, 2)
	currentStat.PumpkinFact = roundFloat(currentStat.PumpkinFact+stat.PumpkinFact, 2)
	currentStat.PumpkinDif = roundFloat(currentStat.PumpkinDif+stat.PumpkinDif, 2)
	currentStat.PeanutPlan = roundFloat(currentStat.PeanutPlan+stat.PeanutPlan, 2)
	currentStat.PeanutFact = roundFloat(currentStat.PeanutFact+stat.PeanutFact, 2)
	currentStat.PeanutDif = roundFloat(currentStat.PeanutDif+stat.PeanutDif, 2)
	currentStat.AKB1 += stat.AKB1
	currentStat.AKB2 += stat.AKB2
	currentStat.NewTT += stat.NewTT
	currentStat.Mix += stat.Mix
	currentStat.NpOne += stat.NpOne
	currentStat.SetShel += stat.SetShel
	currentStat.DMP += stat.DMP
	currentStat.TopFive += stat.TopFive
	currentStat.News += stat.News

	regionalStats.stats[regionID] = currentStat
	regionalStats.quantities[regionID]++
}

// Получить агрегированную статистику для региона
func GetStatsForRegion(regionID uint) (models.StatDaily, int) {
	regionalStats.mu.RLock()
	defer regionalStats.mu.RUnlock()

	stat := regionalStats.stats[regionID]
	quantity := regionalStats.quantities[regionID]
	return stat, quantity
}

// Получить общую статистику по всем регионам
func GetTotalStats() (models.StatDaily, int) {
	regionalStats.mu.RLock()
	defer regionalStats.mu.RUnlock()

	var totalStat models.StatDaily
	totalQuantity := 0

	for _, stat := range regionalStats.stats {
		totalStat.SeedPlan = roundFloat(totalStat.SeedPlan+stat.SeedPlan, 2)
		totalStat.SeedFact = roundFloat(totalStat.SeedFact+stat.SeedFact, 2)
		totalStat.SeedDif = roundFloat(totalStat.SeedDif+stat.SeedDif, 2)
		totalStat.PumpkinPlan = roundFloat(totalStat.PumpkinPlan+stat.PumpkinPlan, 2)
		totalStat.PumpkinFact = roundFloat(totalStat.PumpkinFact+stat.PumpkinFact, 2)
		totalStat.PumpkinDif = roundFloat(totalStat.PumpkinDif+stat.PumpkinDif, 2)
		totalStat.PeanutPlan = roundFloat(totalStat.PeanutPlan+stat.PeanutPlan, 2)
		totalStat.PeanutFact = roundFloat(totalStat.PeanutFact+stat.PeanutFact, 2)
		totalStat.PeanutDif = roundFloat(totalStat.PeanutDif+stat.PeanutDif, 2)
		totalStat.AKB1 += stat.AKB1
		totalStat.AKB2 += stat.AKB2
		totalStat.NewTT += stat.NewTT
		totalStat.Mix += stat.Mix
		totalStat.NpOne += stat.NpOne
		totalStat.SetShel += stat.SetShel
		totalStat.DMP += stat.DMP
		totalStat.TopFive += stat.TopFive
		totalStat.News += stat.News
	}

	for _, quantity := range regionalStats.quantities {
		totalQuantity += quantity
	}

	return totalStat, totalQuantity
}

// Очистить статистику для конкретного региона
func ClearStatsForRegion(regionID uint) {
	regionalStats.mu.Lock()
	defer regionalStats.mu.Unlock()

	delete(regionalStats.stats, regionID)
	delete(regionalStats.quantities, regionID)
}

// Очистить всю статистику
func ClearAllStats() {
	regionalStats.mu.Lock()
	defer regionalStats.mu.Unlock()

	regionalStats.stats = make(map[uint]models.StatDaily)
	regionalStats.quantities = make(map[uint]int)
}

// Инициализировать статистику при запуске приложения
func InitializeFromDB(db interface{}) error {
	// Здесь можно добавить логику загрузки существующих данных из БД
	// при старте приложения, если это необходимо
	return nil
}

// Получить список всех регионов со статистикой
func GetAllRegionalStats() map[uint]models.StatDaily {
	regionalStats.mu.RLock()
	defer regionalStats.mu.RUnlock()

	result := make(map[uint]models.StatDaily)
	for regionID, stat := range regionalStats.stats {
		result[regionID] = stat
	}
	return result
}

// Получить количество отчетов для всех регионов
func GetAllQuantities() map[uint]int {
	regionalStats.mu.RLock()
	defer regionalStats.mu.RUnlock()

	result := make(map[uint]int)
	for regionID, quantity := range regionalStats.quantities {
		result[regionID] = quantity
	}
	return result
}

func roundFloat(val float64, precision int) float64 {
	formatted := fmt.Sprintf("%.*f", precision, val)
	result, _ := strconv.ParseFloat(formatted, 64)
	return result
}
