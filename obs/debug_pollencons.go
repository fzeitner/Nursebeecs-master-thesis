package obs

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_ecotox against BEEHAVE_ecotox.
type DebugPollenCons struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	data     []float64
	cons     *globals.ConsumptionStats
}

func (o *DebugPollenCons) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
}

func (o *DebugPollenCons) Update(w *ecs.World) {}
func (o *DebugPollenCons) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "ProteinFactorNurses", "DailyHoneyConsumption_mg", "DailyPollenConsumption_g", "TotalDroneEggs", "TotalDroneLarvae", "TotalDronePupae", "TotalDrones", "TotalPop"}
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

	o.data[8] = float64(o.stores.ProteinFactorNurses)

	o.data[9] = float64(o.cons.HoneyDaily)
	o.data[10] = float64(o.cons.PollenDaily)
	o.data[11] = float64(o.pop.DroneEggs)
	o.data[12] = float64(o.pop.DroneLarvae)
	o.data[13] = float64(o.pop.DronePupae)
	o.data[14] = float64(o.pop.DronesInHive)
	o.data[15] = float64(o.pop.WorkerEggs + o.pop.WorkerLarvae + o.pop.WorkerPupae + o.pop.WorkersInHive + o.pop.WorkersForagers + o.pop.DroneEggs + o.pop.DroneLarvae + o.pop.DronePupae + o.pop.DronesInHive)

	return o.data
}
