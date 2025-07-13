package obs

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_masterthesis against BEEHAVE_ecotox.
type DebugNursebees struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	data     []float64
	forstats *globals.ForagingStats
}

func (o *DebugNursebees) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.forstats = ecs.GetResource[globals.ForagingStats](w)
}
func (o *DebugNursebees) Update(w *ecs.World) {}
func (o *DebugNursebees) Header() []string {
	return []string{"HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "ETOX_Cum_Dose_Larvae", "ETOX_Cum_Dose_IHbee", "ETOX_Cum_Dose_Forager", "nIHbeeCohorts", "pollenconcbeforeeating", "nectarconcbeforeeating", "Cum_PPP_Nursebees"}
}
func (o *DebugNursebees) Values(w *ecs.World) []float64 {
	o.data[0] = o.stores.Honey
	o.data[1] = o.stores.Pollen

	o.data[2] = float64(o.pop.WorkerEggs)
	o.data[3] = float64(o.pop.WorkerLarvae)
	o.data[4] = float64(o.pop.WorkerPupae)
	o.data[5] = float64(o.pop.WorkersInHive)
	o.data[6] = float64(o.pop.WorkersForagers)

	o.data[7] = float64(o.pop.CumDoseLarvae)
	o.data[8] = float64(o.pop.CumDoseIHBees)
	o.data[9] = float64(o.pop.CumDoseForagers)

	o.data[10] = float64(o.pop.NumberIHbeeCohorts)
	o.data[11] = float64(o.stores.Pollenconcbeforeeating)
	o.data[12] = float64(o.stores.Nectarconcbeforeeating)
	o.data[13] = float64(o.pop.PPPNursebees)

	return o.data
}
