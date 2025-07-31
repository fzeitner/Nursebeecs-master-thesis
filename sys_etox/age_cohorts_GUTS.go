package sys_etox

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// AgeCohorts moves all cohort-based age classes to the next day's cohort.
// It also handles transition from eggs to larvae, larvae to pupae and pupae to in-hive bees.
// It does not handle transition from in-hive bees to foragers.
type AgeCohorts_GUTS struct {
	eggs        *globals.Eggs
	larvae      *globals.Larvae
	pupae       *globals.Pupae
	inHive      *globals.InHive
	newCohorts  *globals.NewCohorts
	inHive_etox *globals_etox.InHive_etox
	etox        *params_etox.ETOXparams
	GUTS        *params_etox.GUTSParams

	time *resource.Tick
}

func (s *AgeCohorts_GUTS) Initialize(w *ecs.World) {
	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.GUTS = ecs.GetResource[params_etox.GUTSParams](w)

}

func (s *AgeCohorts_GUTS) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		s.newCohorts.IHbees = s.pupae.Workers[len(s.pupae.Workers)-1]
		s.newCohorts.Drones = s.pupae.Drones[len(s.pupae.Drones)-1]

		shiftCohorts(s.inHive.Workers, 0)
		if s.etox.Application && s.etox.GUTS {
			shiftCohortsFloat(s.inHive_etox.WorkerCohortC_i, 0)
			shiftCohortsFloat(s.inHive_etox.WorkerCohortDose, 0)
			if s.GUTS.Type == "IT" {
				shiftCohortsFloat(s.inHive_etox.WorkerCohortITthreshold, 0)
			}
		}

		shiftCohorts(s.inHive.Drones, 0)

		shiftCohorts(s.pupae.Workers, s.larvae.Workers[len(s.larvae.Workers)-1])
		shiftCohorts(s.pupae.Drones, s.larvae.Drones[len(s.larvae.Drones)-1])

		shiftCohorts(s.larvae.Workers, s.eggs.Workers[len(s.eggs.Workers)-1])
		shiftCohorts(s.larvae.Drones, s.eggs.Drones[len(s.eggs.Drones)-1])

		shiftCohorts(s.eggs.Workers, 0)
		shiftCohorts(s.eggs.Drones, 0)
	}
}

func (s *AgeCohorts_GUTS) Finalize(w *ecs.World) {}

func shiftCohorts(coh []int, add int) {
	for i := len(coh) - 1; i > 0; i-- {
		coh[i] = coh[i-1]
	}
	coh[0] = add
}

func shiftCohortsFloat(coh []float64, add float64) {
	for i := len(coh) - 1; i > 0; i-- {
		coh[i] = coh[i-1]
	}
	coh[0] = add
}
