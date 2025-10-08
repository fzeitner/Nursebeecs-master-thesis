package sys_etox

import (
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

	newCons  *params_etox.ConsumptionRework
	larvae   *globals.Larvae
	inHive   *globals.InHive
	nglobals *globals_etox.Nursing_globals

	pop  *globals.PopulationStats
	cons *globals.ConsumptionStats
	time *resource.Tick
}

func (s *Newbroodcare) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params_etox.Nursing](w)

	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *Newbroodcare) Update(w *ecs.World) {
	if s.time.Tick > 0 {

	}
}
