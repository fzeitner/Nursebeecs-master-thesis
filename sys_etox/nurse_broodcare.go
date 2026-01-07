package sys_etox

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/params_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/util"
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
	s.rng = ecs.GetResource[resource.Rand](w)
}

func (s *Nbroodcare) Update(w *ecs.World) {
	// implement rules for killing of larvae here as per old brood care proc

	if !s.NurseParams.NewBroodCare { // old broodcare
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
		if s.nglobals.AbortNursing { // stop nursing completely and kill everything that needs active care aside from thermoregulation
			if s.nglobals.KillDrones {
				s.killDrones()
			}
			nonPupaeBrood := s.pop.DroneEggs + s.pop.DroneLarvae + s.pop.WorkerEggs + s.pop.WorkerLarvae
			s.killBrood(nonPupaeBrood, true) // simply kill all larvae and eggs; capped pupae do not die, as they are not reliant on being fed, they might still die via reduction of excess brood though

			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for literal feeding of brood through nurses

			excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-nonPupaeBrood)-maxBrood)), 0)
			s.ReduceBroodCells(excessBrood)

		} else if s.NurseParams.ForesightedCannibalism { // normal case for new broodcare; gets called as long as there are some nurses left
			if s.nglobals.KillDrones {
				s.killDrones()
			}

			// calculate relative cannibalism intensity here; this is experimental for now
			cann_mean := 0.
			cann_rel := 0.

			for i := 0; i < 4; i++ { // calculate the mean realistic amount of cannibalizable brood (Schmickl&Crailsheim 2001)
				cann_mean += s.NurseParams.BroodCannibalismChance[i] * float64(s.larvae.Drones[i]+s.larvae.Workers[i]) // assume drone and worker larvae get treated indifferently
			}
			// calculate brood to be cannibalized based on pollen influx in the last few days
			if s.nglobals.LastPollenInflux > 0 { // if there has not been a pollen influx this day increase cannibalized brood fraction
				cann_rel = min(float64(s.nglobals.LastPollenInflux/5), 1.5) // adds relative cannibalism rate on top of this based on time of last pollen influx (assume relative strengt grows linearly over 3 days; Schmickl&Crailsheim 2001)
			}
			toStarve := cann_rel * cann_mean

			starved := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // THIS is a lack of protein or a lack of "feeding nurses" that provide this protein which results in cannibalism
			if float64(starved) > toStarve {
				cann_rel = util.Clamp(float64(starved)/cann_mean, 0, 1.5) // calculate how the cannibalism rates shall be reduced relative starved brood from ProteinFactorNurses
			} else {
				starved = int(toStarve)
			}

			killed := s.Cannibalize(starved, cann_rel)

			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for literal feeding of brood through nurses
			excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-killed)-maxBrood)), 0)

			s.ReduceBroodCells(excessBrood)
		} else { // no cannibalism based on days since last pollen influx
			if s.nglobals.KillDrones {
				s.killDrones()
			}

			starved := int(math.Ceil((float64(s.pop.WorkerLarvae+s.pop.DroneLarvae) * (1.0 - s.stores.ProteinFactorNurses)))) // and THIS is a lack of protein or a lack of "feeding nurses" that provide this protein which results in cannibalism

			// calculate relative cannibalism intensity here; this is experimental for now
			cann_mean := 0. // mean count of typically cannibalizable brood in times of no pollen income; based on Schmickl&Crailsheim 2001
			cann_rel := 0.  // relative cannibalism intensity that is necessary to reduce brood far enough based on mean cannibalism estimates and protein stores of nurses

			for i := 0; i < 4; i++ { // calculate the mean realistic amount of cannibalize brood (Schmickl&Crailsheim 2001)
				cann_mean += s.NurseParams.BroodCannibalismChance[i] * float64(s.larvae.Drones[i]+s.larvae.Workers[i]) // assumes drone and worker larvae get treated indifferently
			}
			if cann_mean != 0 {
				cann_rel = min(float64(starved)/cann_mean, 1.5) // calculate how the cannibalism rates shall be reduced relative starved brood from ProteinFactorNurses
			}

			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // I actually think Matthias Becher means a maxBrood for thermoregulation capacities here and NOT for literal feeding of brood through nurses

				/*
					// add some sort of nurse-based death mechanism here depending on scenario; this has not been finished and atm both competition types do not differ from each other
					if s.NurseParams.ScrambleComp { // scramble competition; assumes minimum of direct cannibalism based on foraging income and otherwise weakens larvae if there is too little protein

						//starved := int(cann_rel * cann_mean)  // define cannibalized brood here only dependent on foraging based cannibalism
						killed := s.Cannibalize(starved, cann_rel)

						excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-killed)-maxBrood)), 0)
						s.ReduceBroodCells(excessBrood)

						// weaken brood depending on protein stores / ProteinFactorNurses here
						// as of now I am sadly not sure if BEEHAVE can be seen as a good framework for a change of this caliber

					} else { */
			// "winner takes it all"; basically assumes the old BEEHAVE brood care dynamics stay correct with the addition of foraging-income based cannibalism
			killed := s.Cannibalize(starved, cann_rel)

			excessBrood := util.MaxInt(int(math.Ceil(float64(s.pop.TotalBrood-killed)-maxBrood)), 0)
			s.ReduceBroodCells(excessBrood)
			//}
		}
	}
}

func (s *Nbroodcare) Finalize(w *ecs.World) {}

func (s *Nbroodcare) killDrones() {
	for i := 0; i < 9; i++ { // if the age of maturity for drones changes this should be changed too; this is overall rather crude and should not be kept forever
		s.inHive.Drones[i] = 0 // technically all drones rely on getting fed by nurses and get yeeted out of the brood nest in times of resource scarcity, so this should be changed one day
	}
}

func (s *Nbroodcare) Cannibalize(excess int, cann_rel float64) (killed int) {
	killed = 0
	if excess <= 0 && cann_rel == 0. {
		return
	}

	if excess, killed = reduceByCannibalism(s.larvae.Drones, s.larvae.Workers, excess, cann_rel, s.NurseParams.BroodCannibalismChance, s.rng); excess == 0 {
		return
	}

	// now this only happens if "normal" killing of brood through cannibalizing is not enough to regulate brood levels
	oldexc := excess
	if excess = reduceCohorts(s.larvae.Drones, excess); excess == 0 {
		killed += oldexc
		return
	}
	killed += (oldexc - excess)
	oldexc = excess
	if excess = reduceCohorts(s.larvae.Workers, excess); excess == 0 {
		killed += oldexc
		return
	}

	panic("still brood to kill - code should not be reachable")
}

func reduceByCannibalism(cohD []int, cohW []int, exc int, cannibalismMod float64, m []float64, rng rand.Source) (excess int, killed int) {
	// reduces drone and worker cohorts below age 4 stochastically with an observed relative cannibalism chance per age
	// that was taken from Schmickl&Crailsheim (2001)
	excess = exc
	killed = 0
	for i := 0; i < 4; i++ {
		// drone larvae
		num := cohD[i]
		rng_poisson := distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie := int(rng_poisson.Rand())
		excess -= util.MinInt(toDie, num)
		killed += util.MinInt(toDie, num)
		cohD[i] = util.MaxInt(0, num-toDie)

		// and worker larvae
		num = cohW[i]
		rng_poisson = distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m[i] * cannibalismMod * float64(num),
		}
		toDie = int(rng_poisson.Rand())
		excess -= util.MinInt(toDie, num)
		killed += util.MinInt(toDie, num)
		cohW[i] = util.MaxInt(0, num-toDie)
	}
	return util.MaxInt(0, excess), killed // if more got killed due to stochasticity we still return an excess of 0
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
	// necessary here because drones take a bit longer to develope and thus have more cohorts
	if excess > 0 {
		for i := len(cohW); i < len(cohD); i++ {
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
