package sys_etox

import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// this regulates brood care processes when the explicit modelling of nurse bees is activated
// the need for nurses based on current amount of brood gets calculated and brood gets killed off depending on the
// nurse:larva ratio. This is experimental and bound to be adjusted along the line

type Newbroodcare struct {
	oldNurseParams *params.Nursing
	NurseParams    *params_etox.Nursing
	newCons        *params_etox.ConsumptionRework

	stores   *globals.Stores
	eggs     *globals.Eggs
	larvae   *globals.Larvae
	pupae    *globals.Pupae
	inHive   *globals.InHive
	nglobals *globals_etox.Nursing_globals
	nstats   *globals_etox.Nursing_stats
	pop      *globals.PopulationStats
	cons     *globals.ConsumptionStats
	time     *resource.Tick
}

func (s *Newbroodcare) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params_etox.Nursing](w)
	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)

	s.stores = ecs.GetResource[globals.Stores](w)
	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	s.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *Newbroodcare) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		// implement rules for a) nurse recruitment
		if !s.nglobals.SuffNurses {
			s.nglobals.NurseAgeMax += 1 // if there were insufficient nurses the current age threshold gets increased
		}

		if s.nglobals.SuffNurses && s.nstats.NL_ratio >= 2. && float64(s.nstats.TotalNurses)/float64(s.pop.WorkersInHive) >= 0.3 {
			s.nglobals.NurseAgeMax -= 1 // nurse age can be reduced if nurses are sufficient and certain ratios are kept in line, THIS WILL NEED TO BE ADJUSTED
		}
		// maybe it is a better idea to calulate every step if nursing demands could still be fulfilled if nurse max age was lowered by one cohort and then lower it by one

		// and implement rules for b) killing of larvae here as per old brood care proc
		// maybe also go into scenario analysis here for competition types --> the winner takes it all (old BEEHAVE) vs. scramble (every larva weakened)

		// old broodcare
		if !s.newCons.NewBroodCare {
			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // I actually think Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for nursing (literal feeding of brood through bees)

			excessBrood := int(math.Ceil(float64(s.pop.TotalBrood) - maxBrood))
			lacksNurses := excessBrood > 0 // so this actually a lack of total bees thermoregulating the hive and not a lack of "feeding nurses"

			starved := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // and THIS is a lack of protein or a lack of "feeding nurses" that provide this protein

			if starved > excessBrood {
				excessBrood = starved
			}
			s.killBrood(excessBrood, lacksNurses)

		} else { // new broodcare
			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // I actually think Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for nursing (literal feeding of brood through bees)

			excessBrood := int(math.Ceil(float64(s.pop.TotalBrood) - maxBrood))
			lacksThermo := excessBrood > 0 // so this actually a lack of total bees thermoregulating the hive and not a lack of "feeding nurses"

			starved := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // and THIS is a lack of protein or a lack of "feeding nurses" that provide this protein
			// maybe add some sort of nurse-based death mechanism here

			if starved > excessBrood {
				excessBrood = starved
			}

			s.killBrood(excessBrood, lacksThermo)

		}
	}
}

func (s *Newbroodcare) Finalize(w *ecs.World) {}

func (s *Newbroodcare) killBrood(excess int, lacksNurses bool) {
	if excess <= 0 {
		return
	}

	if lacksNurses {
		if excess = reduceCohorts(s.eggs.Drones, excess); excess == 0 {
			return
		}
	}
	if excess = reduceCohorts(s.larvae.Drones, excess); excess == 0 {
		return
	}

	if lacksNurses {
		if excess = reduceCohorts(s.eggs.Workers, excess); excess == 0 {
			return
		}
	}
	if excess = reduceCohorts(s.larvae.Workers, excess); excess == 0 {
		return
	}

	if lacksNurses {
		if excess = reduceCohorts(s.pupae.Drones, excess); excess == 0 {
			return
		}
		if excess = reduceCohorts(s.pupae.Workers, excess); excess == 0 {
			return
		}
	}

	panic("still brood to kill - code should not be reachable")
}

func reduceCohorts(coh []int, excess int) int {
	for i, v := range coh {
		if v >= excess {
			coh[i] -= excess
			return 0
		}
		coh[i] = 0
		excess -= v
	}
	return excess
}
