package summa

import (
	"fmt"
	"strconv"

	"github.com/Wladim1r/statcounter/internal/models"
)

var (
	CurrentlyStat     models.Stat
	CurrentlyQuantity int
)

func Sum(stat models.Stat) {
	fmt.Printf("Updating stat: %+v\n", stat)

	CurrentlyStat.SeedPlan = roundFloat(CurrentlyStat.SeedPlan+stat.SeedPlan, 2)
	CurrentlyStat.SeedFact = roundFloat(CurrentlyStat.SeedFact+stat.SeedFact, 2)
	CurrentlyStat.SeedDif = roundFloat(CurrentlyStat.SeedDif+stat.SeedDif, 2)
	CurrentlyStat.PumpkinPlan = roundFloat(CurrentlyStat.PumpkinPlan+stat.PumpkinPlan, 2)
	CurrentlyStat.PumpkinFact = roundFloat(CurrentlyStat.PumpkinFact+stat.PumpkinFact, 2)
	CurrentlyStat.PumpkinDif = roundFloat(CurrentlyStat.PumpkinDif+stat.PumpkinDif, 2)
	CurrentlyStat.PeanutPlan = roundFloat(CurrentlyStat.PeanutPlan+stat.PeanutPlan, 2)
	CurrentlyStat.PeanutFact = roundFloat(CurrentlyStat.PeanutFact+stat.PeanutFact, 2)
	CurrentlyStat.PeanutDif = roundFloat(CurrentlyStat.PeanutDif+stat.PeanutDif, 2)
	CurrentlyStat.AKB1 += stat.AKB1
	CurrentlyStat.AKB2 += stat.AKB2
	CurrentlyStat.NewTT += stat.NewTT
	CurrentlyStat.Mix += stat.Mix
	CurrentlyStat.NpOne += stat.NpOne
	CurrentlyStat.SetShel += stat.SetShel
	CurrentlyStat.DMP += stat.DMP
	CurrentlyStat.TopFive += stat.TopFive
	CurrentlyStat.News += stat.News
}

func roundFloat(val float64, precision int) float64 {
	formatted := fmt.Sprintf("%.*f", precision, val)
	result, _ := strconv.ParseFloat(formatted, 64)
	return result
}
