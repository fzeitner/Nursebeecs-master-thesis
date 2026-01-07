package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals_etox"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs_ecotox against BEEHAVE_ecotox.
type DebugEcotox struct {
	pop         *globals.PopulationStats
	popetox     *globals_etox.PopulationStats_etox
	stores_etox *globals_etox.Storages_etox
	stores      *globals.Stores
	foraging    *globals.ForagingPeriod
	data        []float64
	forstats    *globals_etox.ForagingStats_etox
}

func (o *DebugEcotox) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popetox = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	o.stores_etox = ecs.GetResource[globals_etox.Storages_etox](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
	o.forstats = ecs.GetResource[globals_etox.ForagingStats_etox](w)
}
func (o *DebugEcotox) Update(w *ecs.World) {}
func (o *DebugEcotox) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "ETOX_Mean_Dose_Larvae_mug", "ETOX_Mean_Dose_IHbee_mug", "ETOX_Mean_Dose_Forager_mug", "ETOX_Mean_Dose_Nurses_mug", "ETOX_Cum_Dose_Larvae_mug", "ETOX_Cum_Dose_IHbee_mug", "ETOX_Cum_Dose_Forager_mug", "ETOX_Cum_Dose_Nurses_mug", "pollenconcbeforeeating_mug_g", "nectarconcbeforeeating_mug_kJ", "TotalPop"}
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

	o.data[8] = float64(o.popetox.MeanDoseLarvae)
	o.data[9] = float64(o.popetox.MeanDoseIHBees)
	o.data[10] = float64(o.popetox.MeanDoseForager)
	o.data[11] = float64(o.popetox.MeanDoseNurses)

	o.data[12] = float64(o.popetox.CumDoseLarvae)
	o.data[13] = float64(o.popetox.CumDoseIHBees)
	o.data[14] = float64(o.popetox.CumDoseForagers)
	o.data[15] = float64(o.popetox.CumDoseNurses)

	o.data[16] = float64(o.stores_etox.Pollenconcbeforeeating)
	o.data[17] = float64(o.stores_etox.Nectarconcbeforeeating)
	o.data[18] = float64(o.pop.WorkerEggs + o.pop.WorkerLarvae + o.pop.WorkerPupae + o.pop.WorkersInHive + o.pop.WorkersForagers + o.pop.DroneEggs + o.pop.DroneLarvae + o.pop.DronePupae + o.pop.DronesInHive)

	return o.data
}
