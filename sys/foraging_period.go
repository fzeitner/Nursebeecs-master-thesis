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

	waterPeriodParams *params.WaterForagingPeriod
	waterPeriodData   *globals.WaterForagingPeriodData
	waterNeeds        globals.WaterNeeds
	waterParams       *params.WaterParams

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

	s.waterParams = ecs.GetResource[params.WaterParams](w)
	s.waterPeriodParams = ecs.GetResource[params.WaterForagingPeriod](w)
	s.waterPeriodData = ecs.GetResource[globals.WaterForagingPeriodData](w)
	s.waterNeeds = globals.WaterNeeds{}
	ecs.AddResource(w, &s.waterNeeds)

}

func (s *CalcForagingPeriod) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		dayOfYear := int(s.time.Tick%365) - 1
		if dayOfYear < 0 {
			dayOfYear = 364
		}
		if dayOfYear == 0 {
			if s.periodParams.RandomYears {
				s.periodData.CurrentYear = s.rng.IntN(len(s.periodData.Years))
			} else {
				s.periodData.CurrentYear = int(s.time.Tick/365) % len(s.periodData.Years)
			}
		}
		year := s.periodData.Years[s.periodData.CurrentYear]
		s.period.SecondsToday = int(year[dayOfYear] * 3600)

		if s.waterParams.WaterForaging { // not bugfixed yet, WaterForaging should not be turned on anyways and thus this never gets accessed
			if dayOfYear == 0 {
				if s.waterPeriodParams.RandomYears {
					s.waterPeriodData.CurrentYear = s.rng.IntN(len(s.waterPeriodData.Years))
				} else {
					s.waterPeriodData.CurrentYear = int(s.time.Tick/365) % len(s.waterPeriodData.Years)
				}
			}
			wateryear := s.waterPeriodData.Years[s.waterPeriodData.CurrentYear]
			s.waterNeeds.ETOX_Waterneedforcooling = wateryear[dayOfYear]
		}
	}
}

func (s *CalcForagingPeriod) Finalize(w *ecs.World) {}
