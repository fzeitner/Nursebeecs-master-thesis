package obs

import (
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_ecotox against BEEHAVE_ecotox.
type DebugEcotox struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	data     []float64
	forstats *globals.ForagingStats
}

func (o *DebugEcotox) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.forstats = ecs.GetResource[globals.ForagingStats](w)
}
func (o *DebugEcotox) Update(w *ecs.World) {}
func (o *DebugEcotox) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "ETOX_Mean_Dose_Larvae", "ETOX_Mean_Dose_IHbee", "ETOX_Mean_Dose_Forager", "ETOX_Cum_Dose_Larvae", "ETOX_Cum_Dose_IHbee", "ETOX_Cum_Dose_Forager", "nIHbeeCohorts", "pollenconcbeforeeating", "nectarconcbeforeeating", "contactonce", "contactrepeat"}
}
func (o *DebugEcotox) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.foraging.SecondsToday)
	o.data[1] = o.stores.Honey
	o.data[2] = o.stores.Pollen

	o.data[3] = float64(o.pop.WorkerEggs)
	o.data[4] = float64(o.pop.WorkerLarvae)
	o.data[5] = float64(o.pop.WorkerPupae)
	o.data[6] = float64(o.pop.WorkersInHive)
	o.data[7] = float64(o.pop.WorkersForagers)

	o.data[8] = float64(o.pop.MeanDoseLarvae)
	o.data[9] = float64(o.pop.MeanDoseIHBees)
	o.data[10] = float64(o.pop.MeanDoseForager)

	o.data[11] = float64(o.pop.CumDoseLarvae)
	o.data[12] = float64(o.pop.CumDoseIHBees)
	o.data[13] = float64(o.pop.CumDoseForagers)

	o.data[14] = float64(o.pop.NumberIHbeeCohorts)
	o.data[15] = float64(o.stores.Pollenconcbeforeeating)
	o.data[16] = float64(o.stores.Nectarconcbeforeeating)
	o.data[17] = float64(o.forstats.ContactExp_once)
	o.data[18] = float64(o.forstats.ContactExp_repeat)

	return o.data
}
