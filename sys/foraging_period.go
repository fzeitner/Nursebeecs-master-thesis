package sys

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type CalcForagingPeriod struct {
	time         *resource.Tick
	periodParams *params.ForagingPeriod
	periodData   *globals.ForagingPeriodData
	period       globals.ForagingPeriod

	rng *rand.Rand
}

func (s *CalcForagingPeriod) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	s.periodParams = ecs.GetResource[params.ForagingPeriod](w)
	s.periodData = ecs.GetResource[globals.ForagingPeriodData](w)

	s.period = globals.ForagingPeriod{}
	ecs.AddResource(w, &s.period)

	src := ecs.GetResource[resource.Rand](w)
	s.rng = rand.New(src)
}

func (s *CalcForagingPeriod) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		dayOfYear := int((s.time.Tick - 1) % 365)

		if dayOfYear == 0 {
			if s.periodParams.RandomYears {
				s.periodData.CurrentYear = s.rng.IntN(len(s.periodData.Years))
			} else {
				s.periodData.CurrentYear = int((s.time.Tick)/365) % len(s.periodData.Years)
			}
		}
		year := s.periodData.Years[s.periodData.CurrentYear]
		s.period.SecondsToday = int(year[dayOfYear] * 3600)
	}
}

func (s *CalcForagingPeriod) Finalize(w *ecs.World) {}
