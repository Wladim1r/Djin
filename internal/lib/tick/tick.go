package tick

import (
	"log"
	"time"

	"github.com/Wladim1r/statcounter/internal/api/repository"
)

func TruncateToTicker(repo repository.DjnRepo) {
	go func() {
		for {

			now := time.Now()
			nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
			delay := nextMidnight.Sub(now)

			select {
			case <-time.After(delay): // Таймер до 00:00
				if err := repo.Truncate(); err != nil {
					log.Printf("error when TRUNCATE: %v", err)
				}
			}

			// После первого запуска — бьём каждые 24 часа
			ticker := time.NewTicker(24 * time.Hour)

			for {
				select {
				case <-ticker.C:
					if err := repo.Truncate(); err != nil {
						log.Printf("error in TRUNCATE: %v", err)
					}
				}
			}
		}
	}()
}
