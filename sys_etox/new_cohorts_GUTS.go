package sys_etox

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/GUTS"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// here new forager, IHbees and drone cohorts get initialized each tick
// had to decouple this from the other systems to get closer to original BEEHAVE,
// because the devs decided to have Broodcare happen after ageing of cohorts and calculating new Ihbees/Fors/Drones,
// but before initializing them as acutal new cohorts and having them counted to the total.
// overall only small difference but it is a difference in quanitative model behaviour

type NewCohorts_GUTS struct {
	inHive      *globals.InHive
	inHive_etox *globals_etox.InHive_etox
	newCohorts  *globals.NewCohorts
	etox        *params_etox.ETOXparams
	GUTS        *params_etox.GUTSParams

	rng  *rand.Rand
	time *resource.Tick
}

func (s *NewCohorts_GUTS) Initialize(w *ecs.World) {
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.GUTS = ecs.GetResource[params_etox.GUTSParams](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)

	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *NewCohorts_GUTS) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		s.inHive.Drones[0] = s.newCohorts.Drones
		s.inHive.Workers[0] = s.newCohorts.IHbees

		if s.etox.Application {
			if s.etox.GUTS {
				if s.GUTS.Type == "IT" {
					s.inHive_etox.WorkerCohortITthreshold[0] = GUTS.Calc_F(s.rng.Float64(), w)
					s.newCohorts.NewForITthreshold = 0.
				}
				s.inHive_etox.WorkerCohortC_i[0] = 0.
				s.inHive_etox.WorkerCohortDose[0] = 0.
				s.newCohorts.NewForC_i = 0.
				s.newCohorts.NewForOralDose = 0.
			}
		}
		s.newCohorts.Drones = 0
		s.newCohorts.IHbees = 0
	}
}

func (s *NewCohorts_GUTS) Finalize(w *ecs.World) {}
