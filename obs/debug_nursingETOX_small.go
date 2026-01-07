package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals_etox"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE implementation.
//
// Primarily meant for validation of beecs against BEEHAVE.
type DebugNursingEtox struct {
	pop         *globals.PopulationStats
	popetox     *globals_etox.PopulationStats_etox
	stores      *globals.Stores
	stores_etox *globals_etox.Storages_etox
	foraging    *globals.ForagingPeriod
	cons        *globals.ConsumptionStats
	nglobals    *globals_etox.Nursing_globals
	nstats      *globals_etox.Nursing_stats
	aff         *globals.AgeFirstForaging

	data []float64
}

func (o *DebugNursingEtox) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popetox = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.stores_etox = ecs.GetResource[globals_etox.Storages_etox](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
	o.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	o.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)
	o.aff = ecs.GetResource[globals.AgeFirstForaging](w)

	o.data = make([]float64, len(o.Header()))
}
func (o *DebugNursingEtox) Update(w *ecs.World) {}
func (o *DebugNursingEtox) Header() []string {
	return []string{"HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "NurseAgeMax", "Aff", "NurseWorkLoad", "ProteinFactorNurses", "TotalNurses", "NonNurseIHbees", "ETOX_Mean_Dose_Larvae_mug", "ETOX_Mean_Dose_IHbee_mug", "ETOX_Mean_Dose_Forager_mug", "ETOX_Mean_Dose_Nurses_mug", "pollenconcbeforeeating_mug_g", "nectarconcbeforeeating_mug_kJ", "NurseMeanHoneyIntake", "NurseMeanPollenIntake", "Winterbees", "RevertedForagers", "TotalPop"}
}
func (o *DebugNursingEtox) Values(w *ecs.World) []float64 {
	o.data[0] = o.stores.Honey
	o.data[1] = o.stores.Pollen

	o.data[2] = float64(o.pop.WorkerEggs)
	o.data[3] = float64(o.pop.WorkerLarvae)
	o.data[4] = float64(o.pop.WorkerPupae)
	o.data[5] = float64(o.pop.WorkersInHive)
	o.data[6] = float64(o.pop.WorkersForagers)

	o.data[7] = float64(o.nglobals.NurseAgeMax)
	o.data[8] = float64(o.aff.Aff)
	o.data[9] = float64(o.nglobals.NurseWorkLoad)
	o.data[10] = float64(o.stores.ProteinFactorNurses)

	o.data[11] = float64(o.nstats.TotalNurses)
	o.data[12] = float64(o.nstats.NonNurseIHbees)

	o.data[13] = float64(o.popetox.MeanDoseLarvae)
	o.data[14] = float64(o.popetox.MeanDoseIHBees)
	o.data[15] = float64(o.popetox.MeanDoseForager)
	o.data[16] = float64(o.popetox.MeanDoseNurses)

	o.data[17] = float64(o.stores_etox.Pollenconcbeforeeating)
	o.data[18] = float64(o.stores_etox.Nectarconcbeforeeating)

	o.data[19] = o.nstats.MeanHoneyIntake
	o.data[20] = o.nstats.MeanPollenIntake

	o.data[21] = float64(o.nstats.WinterBees)
	o.data[22] = float64(o.nstats.RevertedForagers)
	o.data[23] = float64(o.pop.WorkerEggs + o.pop.WorkerLarvae + o.pop.WorkerPupae + o.pop.WorkersInHive + o.pop.WorkersForagers + o.pop.DroneEggs + o.pop.DroneLarvae + o.pop.DronePupae + o.pop.DronesInHive)

	return o.data
}
