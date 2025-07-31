package sys_etox

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// TransitionForagers transitions all in-hive bees of age equal or above [globals.AgeFirstForaging.Aff]
// to forager squadrons.
type TransitionForagers_GUTS struct {
	time        *resource.Tick
	params      *params.Foragers
	factory     *globals.ForagerFactory
	inHive      *globals.InHive
	aff         *globals.AgeFirstForaging
	newCohorts  *globals.NewCohorts
	inHive_etox *globals_etox.InHive_etox
	etox        *params_etox.ETOXparams
	GUTS        *params_etox.GUTSParams
}

func (s *TransitionForagers_GUTS) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	s.params = ecs.GetResource[params.Foragers](w)
	s.factory = ecs.GetResource[globals.ForagerFactory](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.GUTS = ecs.GetResource[params_etox.GUTSParams](w)
}

func (s *TransitionForagers_GUTS) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		aff := s.aff.Aff
		newForagers := 0
		squadrons := 0
		remainder := 0
		if s.etox.Application && s.etox.GUTS && s.GUTS.Type == "IT" { // ToDo: clean this up
			c := 0
			newITval := 0.
			newCival := 0.
			newOraldose := 0.
			for i := aff; i < len(s.inHive.Workers); i++ {
				newForagers = s.inHive.Workers[i]
				s.inHive.Workers[i] = 0
				squadrons += newForagers / s.params.SquadronSize // need to test this
				remainder += newForagers % s.params.SquadronSize // need to test this

				newITval += s.inHive_etox.WorkerCohortITthreshold[i] * float64(newForagers)
				s.inHive_etox.WorkerCohortITthreshold[i] = 0.
				newCival += s.inHive_etox.WorkerCohortC_i[i] * float64(newForagers)
				s.inHive_etox.WorkerCohortC_i[i] = 0.
				newOraldose += s.inHive_etox.WorkerCohortDose[i] * float64(newForagers)
				s.inHive_etox.WorkerCohortDose[i] = 0.
				c += newForagers
			}
			if c == 0 {
				c = 1
			}
			newITval /= float64(c)
			s.newCohorts.NewForITthreshold = newITval
			newCival /= float64(c)
			s.newCohorts.NewForC_i = newCival
			newOraldose /= float64(c)
			s.newCohorts.NewForOralDose = newOraldose

			if s.inHive.Workers[aff-1] > 0 {
				s.inHive_etox.WorkerCohortITthreshold[aff-1] = CalcRemainderDose(remainder, s.inHive.Workers[aff-1], s.inHive_etox.WorkerCohortITthreshold[aff-1], newITval)
				s.inHive_etox.WorkerCohortC_i[aff-1] = CalcRemainderDose(remainder, s.inHive.Workers[aff-1], s.inHive_etox.WorkerCohortC_i[aff-1], newCival)
				s.inHive_etox.WorkerCohortDose[aff-1] = CalcRemainderDose(remainder, s.inHive.Workers[aff-1], s.inHive_etox.WorkerCohortDose[aff-1], newOraldose)

				s.inHive.Workers[aff-1] += remainder

			} else {
				s.inHive.Workers[aff-1] = remainder
				s.inHive_etox.WorkerCohortITthreshold[aff-1] = newITval
				s.inHive_etox.WorkerCohortC_i[aff-1] = newCival
				s.inHive_etox.WorkerCohortDose[aff-1] = newOraldose
			}
			s.newCohorts.Foragers = squadrons

		} else if s.etox.Application && s.etox.GUTS && s.GUTS.Type == "SD" { // clean this up
			c := 0
			newCival := 0.
			newOraldose := 0.

			for i := aff; i < len(s.inHive.Workers); i++ {
				newForagers = s.inHive.Workers[i]
				s.inHive.Workers[i] = 0
				squadrons += newForagers / s.params.SquadronSize // need to test this
				remainder += newForagers % s.params.SquadronSize // need to test this

				newCival += s.inHive_etox.WorkerCohortC_i[i] * float64(newForagers)
				s.inHive_etox.WorkerCohortC_i[i] = 0.
				newOraldose += s.inHive_etox.WorkerCohortDose[i] * float64(newForagers)
				s.inHive_etox.WorkerCohortDose[i] = 0.
				c += newForagers
			}
			if c == 0 {
				c = 1
			}
			newCival /= float64(c)
			s.newCohorts.NewForC_i = newCival
			newOraldose /= float64(c)
			s.newCohorts.NewForOralDose = newOraldose

			if s.inHive.Workers[aff-1] > 0 {
				s.inHive_etox.WorkerCohortC_i[aff-1] = CalcRemainderDose(remainder, s.inHive.Workers[aff-1], s.inHive_etox.WorkerCohortC_i[aff-1], newCival)
				s.inHive_etox.WorkerCohortDose[aff-1] = CalcRemainderDose(remainder, s.inHive.Workers[aff-1], s.inHive_etox.WorkerCohortDose[aff-1], newOraldose)

				s.inHive.Workers[aff-1] += remainder
			} else {
				s.inHive.Workers[aff-1] = remainder
				s.inHive_etox.WorkerCohortC_i[aff-1] = newCival
				s.inHive_etox.WorkerCohortDose[aff-1] = newOraldose
			}
			s.newCohorts.Foragers = squadrons

		} else {
			for i := aff; i < len(s.inHive.Workers); i++ {
				newForagers = s.inHive.Workers[i]
				s.inHive.Workers[i] = 0
				squadrons += newForagers / s.params.SquadronSize // need to test this
				remainder += newForagers % s.params.SquadronSize // need to test this
			}

			s.inHive.Workers[aff-1] += remainder
			s.newCohorts.Foragers = squadrons
		}
	}
}

func (s *TransitionForagers_GUTS) Finalize(w *ecs.World) {}

func CalcRemainderDose(remainder int, cohortnumber int, oldval float64, remainderval float64) (newval float64) {
	newnum := remainder + cohortnumber
	oldconc := oldval * float64(cohortnumber)
	remconc := remainderval * float64(remainder)
	newval = (oldconc + remconc) / float64(newnum)
	return
}
