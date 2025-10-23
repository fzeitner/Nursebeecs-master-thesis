package obs

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE implementation.
//
// Primarily meant for validation of beecs against BEEHAVE.
type DebugNursing struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	cons     *globals.ConsumptionStats
	nglobals *globals_etox.Nursing_globals
	nstats   *globals_etox.Nursing_stats
	aff      *globals.AgeFirstForaging

	data []float64
}

func (o *DebugNursing) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
	o.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	o.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)
	o.aff = ecs.GetResource[globals.AgeFirstForaging](w)

	o.data = make([]float64, len(o.Header()))
}
func (o *DebugNursing) Update(w *ecs.World) {}
func (o *DebugNursing) Header() []string {
	return []string{"Pollendaily", "HoneyDaily", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHbees", "TotalForagers", "NurseAgeMax", "Aff", "NurseWorkLoad", "ProteinFactorNurses", "TotalNurses", "NurseLarvaRatio", "FractionNurses", "NonNurseIHbees", "NurseMaxPollenIntake", "NurseMeanPollenIntake"}
}
func (o *DebugNursing) Values(w *ecs.World) []float64 {
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
	o.data[16] = float64(o.nstats.NonNurseIHbees)

	o.data[17] = o.nstats.MaxPollenIntake
	o.data[18] = o.nstats.MeanPollenIntake

	return o.data
}
