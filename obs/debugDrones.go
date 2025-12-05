package obs

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE implementation.
//
// Primarily meant for validation of beecs against BEEHAVE.
type DebugDrones struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	aff      *globals.AgeFirstForaging
	data     []float64
}

func (o *DebugDrones) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *DebugDrones) Update(w *ecs.World) {}
func (o *DebugDrones) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "TotalPop", "Aff", "TotalDrones", "TotalDroneEggs", "TotalDroneLarvae"}
}
func (o *DebugDrones) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.foraging.SecondsToday)
	o.data[1] = o.stores.Honey
	o.data[2] = o.stores.Pollen

	o.data[3] = float64(o.pop.WorkerEggs)
	o.data[4] = float64(o.pop.WorkerLarvae)
	o.data[5] = float64(o.pop.WorkerPupae)
	o.data[6] = float64(o.pop.WorkersInHive)
	o.data[7] = float64(o.pop.WorkersForagers)
	o.data[8] = float64(o.pop.TotalPopulation)

	o.data[9] = float64(o.aff.Aff)

	o.data[10] = float64(o.pop.DronesInHive)
	o.data[11] = float64(o.pop.DroneEggs)
	o.data[12] = float64(o.pop.DroneLarvae)

	return o.data
}
