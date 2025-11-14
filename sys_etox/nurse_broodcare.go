package sys_etox

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"gonum.org/v1/gonum/stat/distuv"
)

// this regulates brood care processes when the explicit modelling of nurse bees is activated
// the need for nurses gets calculated based on the current food needs of the brood and brood gets killed off
// depending on the pollen stores and time since last pollen influx. This is experimental and bound to be adjusted along the line

type Nbroodcare struct {
	oldNurseParams *params.Nursing
	NurseParams    *params_etox.Nursing
	newCons        *params_etox.ConsumptionRework

	aff      *globals.AgeFirstForaging
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
	rng      *resource.Rand
}

func (s *Nbroodcare) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params_etox.Nursing](w)
	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)

	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
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
	s.rng = ecs.GetResource[resource.Rand](w)
}

func (s *Nbroodcare) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		// implement rules for a) nurse recruitment
		NonNurseIHbees := 0
		for i := s.nglobals.NurseAgeMax + 1; i < s.aff.Aff; i++ {
			NonNurseIHbees += s.inHive.Workers[i]
		}
		if !s.nglobals.SuffNurses && len(s.nglobals.WinterBees) < 10 && float64(s.nstats.TotalNurses/s.pop.TotalAdults) < 0.5 && s.nglobals.NurseAgeMax < s.aff.Aff { // && float64(NonNurseIHbees/s.pop.TotalAdults) >= 0.1 {
			s.nglobals.NurseAgeMax = util.Clamp(s.nglobals.NurseAgeMax+1, 5, 50)
		}
		if s.time.Tick > 204 { // debugging hook; will be taken out
			a := 1
			a++
		}
		if (s.stores.ProteinFactorNurses < 0.9 && len(s.nglobals.WinterBees) == 0) || s.nglobals.Reductionpossible {
			s.nglobals.NurseAgeMax = util.Clamp(s.nglobals.NurseAgeMax-1, 5, 50)
		}
		// maybe it is a better idea to calulate every step if nursing demands could still be fulfilled if nurse max age was lowered by the next nonzero-cohort and then lower NurseAgeMax to that age

		// and implement rules for b) killing of larvae here as per old brood care proc
		// maybe also go into scenario analysis here for competition types --> the winner takes it all (old BEEHAVE) vs. scramble (every larva weakened)

		// old broodcare
		if !s.NurseParams.NewBroodCare {
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
			starved := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // and THIS is a lack of protein or a lack of "feeding nurses" that provide this protein which results in cannibalism

			// calculate relative cannibalism intensity here; this is experimental for now
			cann_mean := 0.
			for i := 0; i < 4; i++ { // calculate the mean realistic amount of cannibalize brood (Schmickl&Crailsheim 2001)
				cann_mean += s.NurseParams.BroodCannibalismChance[i] * float64(s.larvae.Drones[i]+s.larvae.Workers[i]) // assume drone and worker larvae get treated indifferently
			}
			cann_rel := 0.
			cann_rel = min(float64(starved)/cann_mean, 1.5) // calculate how the cannibalism rates shall be reduced relative starved brood from ProteinFactorNurses
			//if s.nglobals.LastPollenInflux > 0 {            // if there has not been a pollen influx this day increase cannibalized brood fraction
			//	cann_rel += min(float64(s.nglobals.LastPollenInflux/5), 1.0) // adds relative cannibalism rate on top of this based on time of last pollen influx (assume relative strengt grows linearly over 5 days; Schmickl&Crailsheim 2001)
			//	cann_rel = util.Clamp(cann_rel, 0., 1.5)
			//}

			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // I actually think Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for literal feeding of brood through nurses

			excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-starved)-maxBrood)), 0)
			// maybe add some sort of nurse-based death mechanism here depending on scenario
			if s.NurseParams.ScrambleComp { // this does not do anything atm and is a placeholder for eventual scenario creation
				s.Cannibalize(starved, cann_rel)

				s.ReduceBroodCells(excessBrood)
			} else {
				s.Cannibalize(starved, cann_rel)

				s.ReduceBroodCells(excessBrood)
			}

		}
	}
}

func (s *Nbroodcare) Finalize(w *ecs.World) {}

func (s *Nbroodcare) Cannibalize(excess int, cann_rel float64) {
	if excess <= 0 {
		return
	}

	if excess = reduceByCannibalism(s.larvae.Drones, s.larvae.Workers, excess, cann_rel, s.NurseParams.BroodCannibalismChance, s.rng); excess == 0 {
		return
	}

	// now this only happens if "normal" killing of brood through cannibalizing is not enough to regulate brood levels
	if excess = reduceCohorts(s.larvae.Drones, excess); excess == 0 {
		return
	}

	if excess = reduceCohorts(s.larvae.Workers, excess); excess == 0 {
		return
	}

	panic("still brood to kill - code should not be reachable")
}

func reduceByCannibalism(cohD []int, cohW []int, excess int, cannibalismMod float64, m []float64, rng rand.Source) int {
	// reduces drone and worker cohorts below age 4 stochastically with an observed relative cannibalism chance per age
	// that was taken from Schmickl&Crailsheim (2001)
	for i := 0; i < 4; i++ {
		// drone larvae
		num := cohD[i]
		rng_poisson := distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie := int(rng_poisson.Rand())
		excess -= util.MinInt(toDie, num)
		cohD[i] = util.MaxInt(0, num-toDie)

		// and worker larvae
		num = cohW[i]
		rng_poisson = distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie = int(rng_poisson.Rand())
		excess -= util.MinInt(toDie, num)
		cohW[i] = util.MaxInt(0, num-toDie)
	}
	return util.MaxInt(0, excess) // if more got killed due to stochasticity we still return an excess of 0
}

func (s *Nbroodcare) ReduceBroodCells(excess int) {
	// now starts with reducing eggs of drones and workers if there is a need to reduce
	// total brood capacity based on colony size
	if excess <= 0 {
		return
	}

	if excess = reduceCohortsByAge(s.eggs.Drones, s.eggs.Workers, excess); excess == 0 {
		return
	}

	if excess = reduceCohortsByAge(s.larvae.Drones, s.larvae.Workers, excess); excess == 0 {
		return
	}

	if excess = reduceCohortsByAge(s.pupae.Drones, s.pupae.Workers, excess); excess == 0 {
		return
	}

	panic("still brood to kill - code should not be reachable") // reactivate the other reductions if this gets triggered under normal circumstances, just to test for now
}

func reduceCohortsByAge(cohD []int, cohW []int, excess int) int {
	// first reduces drone cohorts with lowest age
	for i := range cohW {
		if cohD[i] >= excess {
			cohD[i] -= excess
			return 0
		}
		excess -= cohD[i]
		cohD[i] = 0
		// then reduces worker cohorts with the lowest age
		if cohW[i] >= excess {
			cohW[i] -= excess
			return 0
		}
		excess -= cohW[i]
		cohW[i] = 0
		// and then iterates over the next age
	}
	return excess
}

func (s *Nbroodcare) killBrood(excess int, lacksNurses bool) {
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
