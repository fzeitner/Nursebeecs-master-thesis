package sys

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/util"
	"github.com/mlange-42/ark/ecs"
)

// CalcAff calculates the age of first foraging,
// based on population structure and stores of the hive.
type CalcAffNbeecs struct {
	affParams    *params.AgeFirstForaging
	energyParams *params.EnergyContent
	nurseParams  *params.Nursing
	consStats    *globals.ConsumptionStats
	stores       *globals.Stores
	pop          *globals.PopulationStats

	aff globals.AgeFirstForaging

	nparamsNew *params.NursingRework
	nstats     *globals.NursingStats
	nglobals   *globals.NursingGlobals
}

func (s *CalcAffNbeecs) Initialize(w *ecs.World) {
	s.affParams = ecs.GetResource[params.AgeFirstForaging](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.nurseParams = ecs.GetResource[params.Nursing](w)
	s.consStats = ecs.GetResource[globals.ConsumptionStats](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)

	s.aff = globals.AgeFirstForaging{
		Aff: s.affParams.Base,
	}
	ecs.AddResource(w, &s.aff)

	s.nparamsNew = ecs.GetResource[params.NursingRework](w)
	s.nstats = ecs.GetResource[globals.NursingStats](w)
	s.nglobals = ecs.GetResource[globals.NursingGlobals](w)
}

func (s *CalcAffNbeecs) Update(w *ecs.World) {
	pollenTH := 0.5
	proteinTH := 1.0
	honeyTH := 35.0 * (s.consStats.HoneyDaily / 1000) * s.energyParams.Honey
	broodTH := 0.1
	foragerToWorkerTH := 0.3

	aff := s.aff.Aff

	if s.stores.Pollen/s.stores.IdealPollen < pollenTH { // included a second pollen criterion, because Protein criterion based on ProteinFactorNurses got removed
		aff--
	}
	if s.nparamsNew.Nursebeecsv1 {
		if s.stores.Pollen <= 0. { // introduced this second pollen criterion, because Protein criterion based on ProteinFactorNurses got removed
			aff--
		}
	} else {
		if s.stores.ProteinFactorNurses < proteinTH {
			aff--
		}
	}
	if s.stores.Honey < honeyTH {
		aff -= 2
	}
	if s.pop.WorkersInHive > 0 &&
		float64(s.pop.WorkersForagers)/float64(s.pop.WorkersInHive) < foragerToWorkerTH {
		aff--
	}
	maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.nurseParams.ForagerNursingContribution) *
		s.nurseParams.MaxBroodNurseRatio
	if maxBrood > 0 && float64(s.pop.TotalBrood)/maxBrood > broodTH {
		aff += 2
	}

	if s.aff.Aff < s.affParams.Base-7 {
		aff++
	} else if s.aff.Aff > s.affParams.Base+7 {
		aff--
	}

	if aff < s.aff.Aff {
		aff = s.aff.Aff - 1
	} else if aff > s.aff.Aff {
		aff = s.aff.Aff + 1
	}
	//}
	s.aff.Aff = util.Clamp(aff, s.affParams.Min, s.affParams.Max)
}

func (s *CalcAffNbeecs) Finalize(w *ecs.World) {}
