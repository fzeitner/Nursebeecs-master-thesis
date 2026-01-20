package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// DebugForaging is a row observer for several foraging variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_ecotox against BEEHAVE_ecotox.
type DebugForaging struct {
	pop        *globals.PopulationStats
	popetox    *globals.PopulationStatsEtox
	storesEtox *globals.StoragesEtox
	stores     *globals.Stores
	foraging   *globals.ForagingPeriod
	data       []float64
	forstats   *globals.ForagingStatsEtox
	cons       *globals.ConsumptionStats
}

func (o *DebugForaging) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popetox = ecs.GetResource[globals.PopulationStatsEtox](w)
	o.storesEtox = ecs.GetResource[globals.StoragesEtox](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.forstats = ecs.GetResource[globals.ForagingStatsEtox](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
}

func (o *DebugForaging) Update(w *ecs.World) {}
func (o *DebugForaging) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "nIHbeeCohorts", "contactonce", "contactrepeat", "ForagingRounds", "ForagingSpontaneousProb", "summedTripDuration", "TotalSearches", "ForagerDied", "Collectionflightstotal", "Pollensuccess"}
}
func (o *DebugForaging) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.foraging.SecondsToday)
	o.data[1] = o.stores.Honey
	o.data[2] = o.stores.Pollen

	o.data[3] = float64(o.pop.WorkerEggs)
	o.data[4] = float64(o.pop.WorkerLarvae)
	o.data[5] = float64(o.pop.WorkerPupae)
	o.data[6] = float64(o.pop.WorkersInHive)
	o.data[7] = float64(o.pop.WorkersForagers)

	o.data[8] = float64(o.popetox.NumberIHbeeCohorts)
	o.data[9] = float64(o.forstats.ContactExp_once)
	o.data[10] = float64(o.forstats.ContactExp_repeat)

	o.data[11] = float64(len(o.forstats.Rounds))
	o.data[12] = float64(o.forstats.Prob)
	o.data[13] = float64(o.forstats.SumDur)
	o.data[14] = float64(o.forstats.TotalSearches)
	o.data[15] = float64(o.forstats.Foragerdied)
	o.data[16] = float64(o.forstats.Collectionflightstotal)
	o.data[17] = float64(o.forstats.Pollensuccess)

	return o.data
}
