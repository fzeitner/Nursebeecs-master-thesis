package sys_etox

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/GUTS"
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

	larvae      *globals.Larvae
	larvae_etox *globals_etox.Larvae_etox
	inHive      *globals.InHive
	inHive_etox *globals_etox.InHive_etox
	popstats    *globals_etox.PopulationStats_etox

	etox  *params_etox.ETOXparams
	toxic *params_etox.Toxicityparams
	guts  *params_etox.GUTSParams

	time *resource.Tick
	rng  *resource.Rand
}

func (s *MortalityCohorts_etox) Initialize(w *ecs.World) {
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.droneMort = ecs.GetResource[params.DroneMortality](w)

	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.larvae_etox = ecs.GetResource[globals_etox.Larvae_etox](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)
	s.popstats = ecs.GetResource[globals_etox.PopulationStats_etox](w)

	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)
	s.guts = ecs.GetResource[params_etox.GUTSParams](w)

	s.time = ecs.GetResource[resource.Tick](w)
	s.rng = ecs.GetResource[resource.Rand](w)

}

func (s *MortalityCohorts_etox) Update(w *ecs.World) {
	if s.time.Tick > 0 {

		s.applyMortalityEtox(s.larvae.Workers, s.larvae_etox.WorkerCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)
		s.applyMortalityEtox(s.larvae.Drones, s.larvae_etox.DroneCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)

		if s.etox.GUTS {
			s.applyMortalityGUTS(s.inHive.Workers, s.inHive_etox.WorkerCohortDose, s.inHive_etox.WorkerCohortC_i, w)
		} else {
			s.applyMortalityEtox(s.inHive.Workers, s.inHive_etox.WorkerCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)
		}
		s.applyMortalityEtox(s.inHive.Drones, s.inHive_etox.DroneCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)

		s.popstats.Reset() // resets cumulative and mean doses for the timestep
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

			if num > 100 { // introduced this to make survival for lower numbers of individuals in cohorts more realisitcally stochastic
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
		}
		coh[i] = util.MaxInt(0, num-toDie)
		dose[i] = 0. // doses get reset to 0 after the mortality check in every timestep, only dose from previous day is ever relevant
	}
}

func (s *MortalityCohorts_etox) applyMortalityGUTS(coh []int, dose []float64, C_i []float64, w *ecs.World) {
	if s.guts.Type == "SD" {
		r := rand.New(s.rng)
		for i := range coh {
			if coh[i] != 0 && dose[i]+C_i[i] > 0 {
				coh[i], dose[i], C_i[i] = GUTS.SD_IHbee(coh[i], dose[i], C_i[i], r, w) // this might work now
			}
			if coh[i] == 0 {
				C_i[i] = 0.
				dose[i] = 0.
			}
		}
	} else {
		for i := range coh {
			lethaldose := false
			if coh[i] != 0 && dose[i]+C_i[i] > 0 {
				lethaldose, dose[i], _, C_i[i] = GUTS.IT(s.inHive_etox.WorkerCohortITthreshold[i], dose[i], 0, C_i[i], w)
			}
			if lethaldose {
				coh[i] = 0
				C_i[i] = 0.
				dose[i] = 0.
			}
		}
	}
}
