package sys

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// MortalityForagers applies worker mortality, including
//   - background mortality from [params.WorkerMortality.InHive]
//   - removal of squadrons reaching [params.WorkerDevelopment.MaxLifespan]
//   - removal of squadrons exceeding [params.WorkerMortality.MaxMilage]
type MortalityForagers struct {
	rng           *resource.Rand
	time          *resource.Tick
	workerMort    *params.WorkerMortality
	workerDev     *params.WorkerDevelopment
	toRemove      []ecs.Entity
	foragerFilter *ecs.Filter2[comp.Age, comp.Milage]
	forstats      *globals_etox.ForagingStats_etox
}

func (s *MortalityForagers) Initialize(w *ecs.World) {
	s.rng = ecs.GetResource[resource.Rand](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.forstats = ecs.GetResource[globals_etox.ForagingStats_etox](w)
}

func (s *MortalityForagers) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		r := rand.New(s.rng)
		query := s.foragerFilter.Query()
		for query.Next() {
			a, m := query.Get()
			if int(s.time.Tick-1)-a.DayOfBirth < 0 {
				panic("Fatal error in age calculation, there are foragers with negative age!")
			}
			if int(s.time.Tick-1)-a.DayOfBirth >= s.workerDev.MaxLifespan ||
				m.Total >= s.workerMort.MaxMilage ||
				r.Float64() < s.workerMort.InHive {
				s.toRemove = append(s.toRemove, query.Entity())
			}
		}
		s.forstats.ForagerDiedLifespan += len(s.toRemove)
		for _, e := range s.toRemove {
			w.RemoveEntity(e)
		}
		s.toRemove = s.toRemove[:0]
	}
}

func (s *MortalityForagers) Finalize(w *ecs.World) {}
