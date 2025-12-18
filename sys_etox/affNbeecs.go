package sys_etox

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
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

	nparamsNew *params_etox.Nursing
	nstats     *globals_etox.Nursing_stats
	nglobals   *globals_etox.Nursing_globals
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

	s.nparamsNew = ecs.GetResource[params_etox.Nursing](w)
	s.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
}

func (s *CalcAffNbeecs) Update(w *ecs.World) {
	pollenTH := 0.5
	proteinTH := 1.0
	honeyTH := 35.0 * (s.consStats.HoneyDaily / 1000) * s.energyParams.Honey
	broodTH := 0.1
	foragerToWorkerTH := 0.3

	//maxNurseTH := 0.5
	//minNurseTH := 0.2
	//minIHbeeTH := 0.1 // non-Nurse IHbees are meant here

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
	/*
		if float64(s.nstats.WinterBees)/float64(s.pop.TotalAdults) <= 0.5 {
			if s.nglobals.SuffNurses && float64(s.nstats.TotalNurses)/float64(s.pop.WorkersInHive+s.pop.WorkersForagers) > minNurseTH { // regulate aff down if there are suficient nurses
				aff -= 1 // can adjust to -1 as well
			}
			if !s.nglobals.SuffNurses && float64(s.nstats.TotalNurses)/float64(s.pop.WorkersInHive+s.pop.WorkersForagers) < maxNurseTH { // regulate it up if there are insufficient
				aff += 1
			}
			if float64(s.nstats.NonNurseIHbees)/float64(s.pop.WorkersInHive+s.pop.WorkersForagers) < minIHbeeTH { // if there are completely inplausible amounts of non-nurse IHbees upregulate aff as well
				aff += 1
			}
		}*/

	if s.aff.Aff < s.affParams.Base-7 {
		aff++
	} else if s.aff.Aff > s.affParams.Base+7 {
		aff--
	}
	/*
		if s.nparamsNew.Nursebeecsv1 {
			if aff < s.aff.Aff && aff > s.nglobals.NurseAgeMax {
				aff = s.aff.Aff - 1
			} else if aff > s.aff.Aff {
				aff = s.aff.Aff + 1
			}
		} else { // old aff*/
	if aff < s.aff.Aff {
		aff = s.aff.Aff - 1
	} else if aff > s.aff.Aff {
		aff = s.aff.Aff + 1
	}
	//}
	s.aff.Aff = util.Clamp(aff, s.affParams.Min, s.affParams.Max)
}

func (s *CalcAffNbeecs) Finalize(w *ecs.World) {}
