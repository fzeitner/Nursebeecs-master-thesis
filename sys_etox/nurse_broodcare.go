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

	stores   *globals.Stores
	eggs     *globals.Eggs
	larvae   *globals.Larvae
	pupae    *globals.Pupae
	inHive   *globals.InHive
	nglobals *globals_etox.Nursing_globals
	pop      *globals.PopulationStats
	time     *resource.Tick
	rng      *resource.Rand
}

func (s *Nbroodcare) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params_etox.Nursing](w)

	s.stores = ecs.GetResource[globals.Stores](w)
	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.rng = ecs.GetResource[resource.Rand](w)
}

func (s *Nbroodcare) Update(w *ecs.World) {
	if s.time.Tick > 0 {

		//  implement rules for killing of larvae here as per old brood care proc
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

		} else {
			// new broodcare
			if s.nglobals.AbortNursing {
				nonPupaeBrood := s.pop.DroneEggs + s.pop.DroneLarvae + s.pop.WorkerEggs + s.pop.WorkerLarvae
				s.killBrood(nonPupaeBrood, true) // simply kill all larvae and eggs; capped pupae do not die, as they are not reliant on being fed, they might still die via reduction of excess brood though

				maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
					s.oldNurseParams.MaxBroodNurseRatio

				excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-nonPupaeBrood)-maxBrood)), 0)
				s.ReduceBroodCells(excessBrood)
			} else {
				// calculate relative cannibalism intensity here; this is experimental for now
				cann_mean := 0. // mean count of typically cannibalizable brood in times of no pollen income; based on Schmickl&Crailsheim 2001
				cann_rel := 0.  // relative cannibalism intensity that is necessary to reduce brood far enough based on mean cannibalism estimates and protein stores of nurses

				for i := 0; i < 4; i++ { // calculate the mean realistic amount of cannibalize brood (Schmickl&Crailsheim 2001)
					cann_mean += s.NurseParams.BroodCannibalismChance[i] * float64(s.larvae.Drones[i]+s.larvae.Workers[i]) // assume drone and worker larvae get treated indifferently
				}

				if s.nglobals.LastPollenInflux > 0 { // if there has not been a pollen influx this day increase cannibalized brood fraction
					cann_rel += min(float64(s.nglobals.LastPollenInflux/5), 1.0) // adds relative cannibalism rate on top of this based on time of last pollen influx (assume relative strengt grows linearly over 5 days; Schmickl&Crailsheim 2001)
				}

				maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
					s.oldNurseParams.MaxBroodNurseRatio // I actually think Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for literal feeding of brood through nurses

				// maybe add some sort of nurse-based death mechanism here depending on scenario
				if s.NurseParams.ScrambleComp { // scramble competition; assumes minimum of direct cannibalism based on foraging income and otherwise weakens larvae if there is too little protein
					starved := int(cann_rel * cann_mean)

					_, starved = s.Cannibalize(starved, cann_rel)

					excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-starved)-maxBrood)), 0)

					s.ReduceBroodCells(excessBrood)
					// weaken brood depending on protein stores / ProteinFactorNurses here

				} else { // "winner takes it all"; basically assumes the old BEEHAVE brood care dynamics stay correct with the addition of foraging-income based cannibalism
					toStarve := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // and THIS is a lack of protein or a lack of "feeding nurses" that provide this protein which results in cannibalism

					if cann_mean != 0 {
						cann_rel += min(float64(toStarve)/cann_mean, 1.0) // calculate how the cannibalism rates shall be reduced relative starved brood from ProteinFactorNurses
						cann_rel = util.Clamp(cann_rel, 0., 1.5)          // assume at max 1.5 times "normally observed" cannibalism rate possible
					}

					toCann := util.MaxInt(toStarve, int(cann_rel*cann_mean)) // this is now just to calculate excessBrood based on an estimate of total brood starved
					_, starved := s.Cannibalize(toCann, cann_rel)

					excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-starved)-maxBrood)), 0)
					s.ReduceBroodCells(excessBrood)
				}
			}
		}
	}
}

func (s *Nbroodcare) Finalize(w *ecs.World) {}

func (s *Nbroodcare) Cannibalize(excess int, cann_rel float64) (int, killed int) {
	if excess, killed = reduceByCannibalism(s.larvae.Drones, s.larvae.Workers, excess, cann_rel, s.NurseParams.BroodCannibalismChance, s.rng); excess <= 0 {
		return excess, killed
	}

	// now this only happens if "normal" killing of brood through cannibalizing is not enough to regulate brood levels
	exc := excess
	if excess = reduceCohorts(s.larvae.Drones, excess); excess <= 0 {
		killed += exc
		return excess, killed
	}

	killed += exc
	exc = excess
	if excess = reduceCohorts(s.larvae.Workers, excess); excess <= 0 {
		killed += exc
		return excess, killed
	}

	panic("still brood to kill - code should not be reachable")
}

func reduceByCannibalism(cohD []int, cohW []int, ex int, cannibalismMod float64, m []float64, rng rand.Source) (excess int, killed int) {
	// reduces drone and worker cohorts below age 4 stochastically with an observed relative cannibalism chance per age
	// that was taken from Schmickl&Crailsheim (2001)
	excess = ex
	for i := 0; i < 4; i++ {
		// drone larvae
		num := cohD[i]
		rng_poisson := distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie := int(rng_poisson.Rand())
		killed += util.MinInt(toDie, num)
		excess -= util.MinInt(toDie, num)
		cohD[i] = util.MaxInt(0, num-toDie)

		// and worker larvae
		num = cohW[i]
		rng_poisson = distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie = int(rng_poisson.Rand())
		killed += util.MinInt(toDie, num)
		excess -= util.MinInt(toDie, num)
		cohW[i] = util.MaxInt(0, num-toDie)
	}
	return excess, killed
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
	// if there still is an excess keep on reducing the rest of the drone cohorts
	if excess > 0 {
		for i := len(cohD) - 2; i < len(cohD); i++ {
			if cohD[i] >= excess {
				cohD[i] -= excess
				return 0
			}
			excess -= cohD[i]
			cohD[i] = 0
		}
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
