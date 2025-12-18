package sys

import (
	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// FixedTermination terminates the simulation after the number of ticks
// given in [params.Termination.MaxTicks].
type FixedTermination struct {
	termRes    *resource.Termination
	termParams *params.Termination
	popStats   *globals.PopulationStats
	step       int64
	time       *resource.Tick

	foragersFilter *ecs.Filter0
	toKill         []ecs.Entity
	inHive         *globals.InHive
}

func (s *FixedTermination) Initialize(w *ecs.World) {
	s.termRes = ecs.GetResource[resource.Termination](w)
	s.termParams = ecs.GetResource[params.Termination](w)
	s.popStats = ecs.GetResource[globals.PopulationStats](w)
	s.step = 0
	s.time = ecs.GetResource[resource.Tick](w)

	s.foragersFilter = ecs.NewFilter0(w).With(ecs.C[comp.Age]())
	s.inHive = ecs.GetResource[globals.InHive](w)
}

func (s *FixedTermination) Update(w *ecs.World) {

	if s.termParams.WinterCritExtinction { // extinction at the end of year depending on colony size; exactly the same as in BEEHAVE now
		if (s.time.Tick)%365 == 0 && (s.popStats.WorkersForagers+s.popStats.WorkersInHive) < s.termParams.CritColonySizeWinter {
			s.KillColony(w)
			//s.termRes.Terminate = true
		}
	}
	if s.termParams.OnExtinction && s.popStats.TotalPopulation == 0 {
		s.termRes.Terminate = true
	}
	if s.termParams.MaxTicks > 0 && s.step+1 >= int64(s.termParams.MaxTicks) {
		s.termRes.Terminate = true
	}
	s.step++
}

// Finalize the system
func (s *FixedTermination) Finalize(w *ecs.World) {}

func (s *FixedTermination) KillColony(w *ecs.World) {
	for i := range s.inHive.Workers {
		s.inHive.Workers[i] = 0
	}
	for i := range s.inHive.Drones {
		s.inHive.Drones[i] = 0
	}

	q := s.foragersFilter.Query()
	for q.Next() {
		s.toKill = append(s.toKill, q.Entity())
	}
	for _, e := range s.toKill {
		w.RemoveEntity(e)
	}
	s.toKill = s.toKill[:0]
}
