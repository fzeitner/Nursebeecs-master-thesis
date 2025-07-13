package sys

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"gonum.org/v1/gonum/stat/distuv"
)

// MortalityCohorts applies background mortality to all cohort-based development stages
// (i.e. all except foragers).
type MortalityCohorts struct {
	workerMort *params.WorkerMortality
	droneMort  *params.DroneMortality
	rng        *resource.Rand

	eggs   *globals.Eggs
	larvae *globals.Larvae
	pupae  *globals.Pupae
	inHive *globals.InHive

	popStats *globals.PopulationStats
	etox     *params.ETOXparams
	time     *resource.Tick
}

func (s *MortalityCohorts) Initialize(w *ecs.World) {
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.droneMort = ecs.GetResource[params.DroneMortality](w)
	s.rng = ecs.GetResource[resource.Rand](w)

	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)

	s.popStats = ecs.GetResource[globals.PopulationStats](w)
	s.etox = ecs.GetResource[params.ETOXparams](w)
	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *MortalityCohorts) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		applyMortality(s.eggs.Workers, s.workerMort.Eggs, s.rng)
		applyMortality(s.eggs.Drones, s.droneMort.Eggs, s.rng)

		applyMortalityWithEtox(s.larvae.Workers, s.workerMort.Larvae, s.rng, s.larvae.WorkerCohortDose, s.etox.LarvaeOralSlope, s.etox.LarvaeOralLD50)
		applyMortalityWithEtox(s.larvae.Drones, s.droneMort.Larvae, s.rng, s.larvae.DroneCohortDose, s.etox.LarvaeOralSlope, s.etox.LarvaeOralLD50)

		applyMortality(s.pupae.Workers, s.workerMort.Pupae, s.rng)
		applyMortality(s.pupae.Drones, s.droneMort.Pupae, s.rng)

		applyMortalityWithEtox(s.inHive.Workers, s.workerMort.InHive, s.rng, s.inHive.WorkerCohortDose, s.etox.ForagerOralSlope, s.etox.ForagerOralLD50)
		applyMortalityWithEtox(s.inHive.Drones, s.droneMort.InHive, s.rng, s.inHive.DroneCohortDose, s.etox.ForagerOralSlope, s.etox.ForagerOralLD50)
	}
}

func (s *MortalityCohorts) Finalize(w *ecs.World) {}

// adapted the mortaliy function for cohorts that can be exposed to PPP, after the normal background mortality there is now
// etox-based mortality depending on the dose of the cohort and a simple dose-response relationship; dose reset every tick
func applyMortalityWithEtox(coh []int, m float64, rng rand.Source, dose []float64, slope float64, LD50 float64) {
	for i := range coh {
		num := coh[i]
		rng := distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m * float64(num),
		}
		toDie := int(rng.Rand())
		coh[i] = util.MaxInt(0, num-toDie)

		if dose[i] > 1e-20 { // simple dose response relationship for all larvae/IHBees/drones
			num = coh[i]
			ldx := (1 - (1 / (1 + math.Pow((dose[i]/LD50), slope))))
			if ldx > 0.99 { // introduced this because netlogo-version behaves the same way. This makes it much less likely to have single digit cohorts left over after lethal PPP events
				ldx = 1
			}
			toDie = int((float64(num) * ldx))
		}
		coh[i] = util.MaxInt(0, num-toDie)
		dose[i] = 0. // doses get reset to 0 after the mortality check in every timestep, only dose from previous day is ever relevant
	}
}

func applyMortality(coh []int, m float64, rng rand.Source) {
	for i := range coh {
		num := coh[i]
		rng := distuv.Poisson{
			Src:    &util.RandWrapper{Src: rng},
			Lambda: m * float64(num),
		}
		toDie := int(rng.Rand())
		coh[i] = util.MaxInt(0, num-toDie)
	}
}
