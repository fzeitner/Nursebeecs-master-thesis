package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// DebugNursingEtox_All is a row observer for several colony structure variables.
//
// Primarily used for debugging Nursebeecs.
type DebugNursingEtox_All struct {
	pop        *globals.PopulationStats
	popetox    *globals.PopulationStatsEtox
	stores     *globals.Stores
	storesEtox *globals.StoragesEtox
	foraging   *globals.ForagingPeriod
	cons       *globals.ConsumptionStats
	nglobals   *globals.NursingGlobals
	nstats     *globals.NursingStats
	aff        *globals.AgeFirstForaging

	data []float64
}

func (o *DebugNursingEtox_All) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popetox = ecs.GetResource[globals.PopulationStatsEtox](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.storesEtox = ecs.GetResource[globals.StoragesEtox](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
	o.nglobals = ecs.GetResource[globals.NursingGlobals](w)
	o.nstats = ecs.GetResource[globals.NursingStats](w)
	o.aff = ecs.GetResource[globals.AgeFirstForaging](w)

	o.data = make([]float64, len(o.Header()))
}
func (o *DebugNursingEtox_All) Update(w *ecs.World) {}
func (o *DebugNursingEtox_All) Header() []string {
	return []string{"Pollendaily", "HoneyDaily", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "NurseAgeMax", "Aff", "NurseWorkLoad", "ProteinFactorNurses", "TotalNurses", "NurseLarvaRatio", "FractionNurses", "ETOX_Mean_Dose_Larvae_mug", "ETOX_Mean_Dose_IHbee_mug", "ETOX_Mean_Dose_Forager_mug", "ETOX_Mean_Dose_Nurses_mug", "ETOX_Cum_Dose_Larvae_mug", "ETOX_Cum_Dose_IHbee_mug", "ETOX_Cum_Dose_Forager_mug", "ETOX_Cum_Dose_Nurses_mug", "pollenconcbeforeeating_mug_g", "nectarconcbeforeeating_mug_kJ", "NonNurseIHbees", "NurseMeanHoneyIntake", "NurseMeanPollenIntake", "Winterbees", "RevertedForagers", "TotalPop"}
}
func (o *DebugNursingEtox_All) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.cons.PollenDaily)
	o.data[1] = float64(o.cons.HoneyDaily)
	o.data[2] = o.stores.Honey
	o.data[3] = o.stores.Pollen

	o.data[4] = float64(o.pop.WorkerEggs)
	o.data[5] = float64(o.pop.WorkerLarvae)
	o.data[6] = float64(o.pop.WorkerPupae)
	o.data[7] = float64(o.pop.WorkersInHive)
	o.data[8] = float64(o.pop.WorkersForagers)

	o.data[9] = float64(o.nglobals.NurseAgeMax)
	o.data[10] = float64(o.aff.Aff)
	o.data[11] = float64(o.nglobals.NurseWorkLoad)
	o.data[12] = float64(o.stores.ProteinFactorNurses)

	o.data[13] = float64(o.nstats.TotalNurses)
	o.data[14] = o.nstats.NL_ratio
	o.data[15] = o.nstats.NurseFraction

	o.data[16] = float64(o.popetox.MeanDoseLarvae)
	o.data[17] = float64(o.popetox.MeanDoseIHBees)
	o.data[18] = float64(o.popetox.MeanDoseForager)
	o.data[19] = float64(o.popetox.MeanDoseNurses)

	o.data[20] = float64(o.popetox.CumDoseLarvae)
	o.data[21] = float64(o.popetox.CumDoseIHBees)
	o.data[22] = float64(o.popetox.CumDoseForagers)
	o.data[23] = float64(o.popetox.CumDoseNurses)

	o.data[24] = float64(o.storesEtox.Pollenconcbeforeeating)
	o.data[25] = float64(o.storesEtox.Nectarconcbeforeeating)

	o.data[26] = float64(o.nstats.NonNurseIHbees)
	o.data[27] = o.nstats.MeanHoneyIntake
	o.data[28] = o.nstats.MeanPollenIntake

	o.data[29] = float64(o.nstats.WinterBees)
	o.data[30] = float64(o.nstats.RevertedForagers)
	o.data[31] = float64(o.pop.WorkerEggs + o.pop.WorkerLarvae + o.pop.WorkerPupae + o.pop.WorkersInHive + o.pop.WorkersForagers + o.pop.DroneEggs + o.pop.DroneLarvae + o.pop.DronePupae + o.pop.DronesInHive)

	return o.data
}
