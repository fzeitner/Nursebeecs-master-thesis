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
)

// MortalityCohorts applies background mortality to all cohort-based development stages
// (i.e. all except foragers).
type MortalityCohorts_etox struct {
	workerMort *params.WorkerMortality
	droneMort  *params.DroneMortality
	rng        *resource.Rand

	larvae      *globals.Larvae
	larvae_etox *globals_etox.Larvae_etox
	inHive      *globals.InHive
	inHive_etox *globals_etox.InHive_etox

	toxic *params_etox.Toxicityparams
	etox  *params_etox.ETOXparams

	time *resource.Tick
}

func (s *MortalityCohorts_etox) Initialize(w *ecs.World) {
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.droneMort = ecs.GetResource[params.DroneMortality](w)
	s.rng = ecs.GetResource[resource.Rand](w)

	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.larvae_etox = ecs.GetResource[globals_etox.Larvae_etox](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)

	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)

	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *MortalityCohorts_etox) Update(w *ecs.World) {
	if s.time.Tick > 0 {

		s.applyMortalityEtox(s.larvae.Workers, s.larvae_etox.WorkerCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)
		s.applyMortalityEtox(s.larvae.Drones, s.larvae_etox.DroneCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)

		s.applyMortalityEtox(s.inHive.Workers, s.inHive_etox.WorkerCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)
		s.applyMortalityEtox(s.inHive.Drones, s.inHive_etox.DroneCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)
	}

}

func (s *MortalityCohorts_etox) Finalize(w *ecs.World) {}

// adapted the mortaliy function for cohorts that can be exposed to PPP, after the normal background mortality there is now
// etox-based mortality depending on the dose of the cohort and a simple dose-response relationship; dose reset every tick

func (s *MortalityCohorts_etox) applyMortalityEtox(coh []int, dose []float64, slope float64, LD50 float64) {
	r := rand.New(s.rng)
	for i := range coh {
		num := coh[i]
		toDie := 0
		if dose[i] > 1e-20 { // simple dose response relationship for all larvae/IHBees/drones
			num = coh[i]
			ldx := (1 - (1 / (1 + math.Pow((dose[i]/LD50), slope))))
			if ldx > 0.99 { // introduced this because netlogo-version behaves the same way. This makes it much less likely to have single digit cohorts left over after lethal PPP events
				ldx = 1
			}
			if s.etox.RealisticStoch {
				if num > 100 { // introduced this to make survival for lower numbers of cohorts more realisitcally stochastic
					toDie = int((float64(num) * ldx))
				} else {
					i := 0
					for i < num {
						if r.Float64() < ldx {
							toDie++
						}
						i++
					}
				}
			} else {
				toDie = int((float64(num) * ldx))
			}
		}
		coh[i] = util.MaxInt(0, num-toDie)
		dose[i] = 0. // doses get reset to 0 after the mortality check in every timestep, only dose from previous day is ever relevant
	}
}
