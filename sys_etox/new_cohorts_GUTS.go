package sys_etox

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/GUTS"
	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
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
	factory     *globals.ForagerFactory
	aff         *globals.AgeFirstForaging
	etox        *params_etox.ETOXparams
	GUTS        *params_etox.GUTSParams
	toAdd       []ecs.Entity

	ageFilter     *ecs.Filter1[comp.Age]
	PPPexpoMapper *ecs.Map2[comp_etox.PPPExpo, comp_etox.PPPLoad]
	etoxExchanger *ecs.Map2[comp_etox.KnownPatch_etox, comp_etox.Activity_etox]

	rng  *rand.Rand
	time *resource.Tick
}

func (s *NewCohorts_GUTS) Initialize(w *ecs.World) {
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.factory = ecs.GetResource[globals.ForagerFactory](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.GUTS = ecs.GetResource[params_etox.GUTSParams](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)

	s.ageFilter = s.ageFilter.New(w)
	s.etoxExchanger = s.etoxExchanger.New(w)
	s.PPPexpoMapper = s.PPPexpoMapper.New(w)

	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *NewCohorts_GUTS) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		s.inHive.Drones[0] = s.newCohorts.Drones
		s.inHive.Workers[0] = s.newCohorts.IHbees

		if s.newCohorts.Foragers > 0 {
			s.factory.CreateSquadrons(s.newCohorts.Foragers, int(s.time.Tick-1)-s.aff.Aff)
		}

		if s.etox.Application {
			// adding etox components to the newly initialized forager entities from the base model
			agequery := s.ageFilter.Without(ecs.C[comp_etox.PPPExpo]()).Query()
			for agequery.Next() {
				s.toAdd = append(s.toAdd, agequery.Entity())
			}
			//exchanger := s.etoxExchanger.Removes(ecs.C[comp.KnownPatch](), ecs.C[comp.Activity]())      // maybe later check this why exchanger results in bugs, works the way it is right now though
			for _, entity := range s.toAdd {
				if s.etox.GUTS && s.GUTS.Type == "IT" {
					s.PPPexpoMapper.Add(entity, &comp_etox.PPPExpo{OralDose: s.newCohorts.NewForOralDose, ContactDose: 0., C_i: s.newCohorts.NewForC_i, RmdSurvivalIT: s.newCohorts.NewForITthreshold}, &comp_etox.PPPLoad{PPPLoad: 0.})
				} else if s.etox.GUTS && s.GUTS.Type == "SD" {
					s.PPPexpoMapper.Add(entity, &comp_etox.PPPExpo{OralDose: s.newCohorts.NewForOralDose, ContactDose: 0., C_i: s.newCohorts.NewForC_i}, &comp_etox.PPPLoad{PPPLoad: 0.})
				} else {
					s.PPPexpoMapper.Add(entity, &comp_etox.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: s.rng.Float64(), RdmSurvivalOral: s.rng.Float64()}, &comp_etox.PPPLoad{})
				}
				s.etoxExchanger.Add(entity, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting})

			}
			s.toAdd = s.toAdd[:0]
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
		s.newCohorts.Foragers = 0
	}
}

func (s *NewCohorts_GUTS) Finalize(w *ecs.World) {}
