package sys

import (
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
}

func (s *FixedTermination) Initialize(w *ecs.World) {
	s.termRes = ecs.GetResource[resource.Termination](w)
	s.termParams = ecs.GetResource[params.Termination](w)
	s.popStats = ecs.GetResource[globals.PopulationStats](w)
	s.step = 0
	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *FixedTermination) Update(w *ecs.World) {

	if s.termParams.OnExtinction && s.termParams.WinterCritExtinction { // extinction at the end of year depending on colony size; exactly the same as in BEEHAVE now
		if (s.time.Tick)%365 == 0 && (s.popStats.WorkersForagers+s.popStats.WorkersInHive) < s.termParams.CritColonySizeWinter {
			s.termRes.Terminate = true
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
