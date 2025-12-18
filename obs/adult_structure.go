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
type AdultStructure struct {
	pop      *globals.PopulationStats
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	cons     *globals.ConsumptionStats
	nglobals *globals_etox.Nursing_globals
	nstats   *globals_etox.Nursing_stats
	aff      *globals.AgeFirstForaging

	data []float64
}

func (o *AdultStructure) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.cons = ecs.GetResource[globals.ConsumptionStats](w)
	o.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	o.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)
	o.aff = ecs.GetResource[globals.AgeFirstForaging](w)

	o.data = make([]float64, len(o.Header()))
}
func (o *AdultStructure) Update(w *ecs.World) {}
func (o *AdultStructure) Header() []string {
	return []string{"TotalIHbees", "TotalForagers", "TotalNurses", "IHbeeNurses", "NonNurseIHbees", "Winterbees", "NormalForagers", "RevertedForagers", "SquadstoReduce"}
}
func (o *AdultStructure) Values(w *ecs.World) []float64 {
	CurrentAdultPop := float64(o.pop.WorkersForagers + o.pop.WorkersInHive)
	if CurrentAdultPop == 0 {
		CurrentAdultPop = 1.
	}
	o.data[0] = float64(o.pop.WorkersInHive) / CurrentAdultPop * 100
	o.data[1] = float64(o.pop.WorkersForagers) / CurrentAdultPop * 100

	o.data[2] = float64(o.nstats.TotalNurses) / CurrentAdultPop * 100
	o.data[3] = float64(o.nstats.IHbeeNurses) / CurrentAdultPop * 100
	o.data[4] = float64(o.nstats.NonNurseIHbees) / CurrentAdultPop * 100
	o.data[5] = float64(o.nstats.WinterBees) / CurrentAdultPop * 100
	o.data[6] = float64(o.pop.WorkersForagers-o.nstats.WinterBees-o.nstats.RevertedForagers) / CurrentAdultPop * 100
	o.data[7] = float64(o.nstats.RevertedForagers) / CurrentAdultPop * 100

	o.data[8] = float64(o.nglobals.SquadstoReduce)

	return o.data
}
