package tick

import (
	"context"
	"log"
	"time"

	"github.com/Wladim1r/statcounter/internal/api/repository"
)

func TruncateToTickerMonthlyWithContext(ctx context.Context, repo repository.DjnRepo) {
	// Функция для удаления старых данных
	deleteOldData := func() {
		cutoffDate := time.Now().AddDate(0, 0, -3)
		if err := repo.DeleteOlderThan(cutoffDate); err != nil {
			log.Printf("error deleting old data: %v", err)
		} else {
			log.Printf("successfully deleted data older than %v", cutoffDate.Format("2006-01-02"))
		}
	}

	// Выполняем первую очистку сразу при запуске
	deleteOldData()

	// Вычисляем время до следующей полночи
	now := time.Now()
	nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	delay := nextMidnight.Sub(now)

	log.Printf("waiting %v until next midnight for cleanup schedule", delay)

	// Ждем до полночи или пока контекст не отменится
	select {
	case <-time.After(delay):
		// Продолжаем
	case <-ctx.Done():
		log.Println("cleanup goroutine cancelled before first scheduled run")
		return
	}

	// Создаем тикер на каждые 24 часа
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Основной цикл
	for {
		select {
		case <-ticker.C:
			deleteOldData()
		case <-ctx.Done():
			log.Println("cleanup goroutine cancelled")
			return
		}
	}
}
