package obs

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_ecotox against BEEHAVE_ecotox.
type DebugPollenCons struct {
	pop         *globals.PopulationStats
	popetox     *globals_etox.PopulationStats_etox
	stores_etox *globals_etox.Storages_etox
	stores      *globals.Stores
	foraging    *globals.ForagingPeriod
	data        []float64
	forstats    *globals_etox.ForagingStats_etox
	cons        *globals.ConsumptionStats
}

func (o *DebugPollenCons) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popetox = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	o.stores_etox = ecs.GetResource[globals_etox.Storages_etox](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.forstats = ecs.GetResource[globals_etox.ForagingStats_etox](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
}

func (o *DebugPollenCons) Update(w *ecs.World) {}
func (o *DebugPollenCons) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "nIHbeeCohorts", "contactonce", "contactrepeat", "DailyHoneyConsumption", "DailyPollenConsumption_g", "TotalDroneEggs", "TotalDroneLarvae", "TotalDronePupae", "TotalDrones"}
}
func (o *DebugPollenCons) Values(w *ecs.World) []float64 {
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

	o.data[11] = float64(o.cons.HoneyDaily)
	o.data[12] = float64(o.cons.PollenDaily)
	o.data[13] = float64(o.pop.DroneEggs)
	o.data[14] = float64(o.pop.DroneLarvae)
	o.data[15] = float64(o.pop.DronePupae)
	o.data[16] = float64(o.pop.DronesInHive)

	return o.data
}
