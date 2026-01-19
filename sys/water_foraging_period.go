package sys

// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day
// this is unfinished and not validated in the Netlogo model version. Can be ignored for now.

import (
	"math/rand/v2"

	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type CalcWaterForagingPeriod struct {
	time *resource.Tick
	rng  *rand.Rand

	waterPeriodParams *params.WaterForagingPeriod
	waterPeriodData   *globals.WaterForagingPeriodData
	waterNeeds        globals.WaterNeeds
	waterParams       *params.WaterForaging
}

func (s *CalcWaterForagingPeriod) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	src := ecs.GetResource[resource.Rand](w)
	s.rng = rand.New(src)

	s.waterParams = ecs.GetResource[params.WaterForaging](w)
	s.waterPeriodParams = ecs.GetResource[params.WaterForagingPeriod](w)
	s.waterPeriodData = ecs.GetResource[globals.WaterForagingPeriodData](w)
	s.waterNeeds = globals.WaterNeeds{}
	ecs.AddResource(w, &s.waterNeeds)
}

func (s *CalcWaterForagingPeriod) Update(w *ecs.World) {
	dayOfYear := int((s.time.Tick) % 365)

	if s.waterParams.WaterForaging { // not bugfixed yet, WaterForaging should not be turned on anyways and thus this never gets accessed
		if dayOfYear == 0 {
			if s.waterPeriodParams.RandomYears {
				s.waterPeriodData.CurrentYear = s.rng.IntN(len(s.waterPeriodData.Years))
			} else {
				s.waterPeriodData.CurrentYear = int((s.time.Tick-1)/365) % len(s.waterPeriodData.Years)
			}
		}
		wateryear := s.waterPeriodData.Years[s.waterPeriodData.CurrentYear]
		s.waterNeeds.ETOX_Waterneedforcooling = wateryear[dayOfYear]
	}
}

func (s *CalcWaterForagingPeriod) Finalize(w *ecs.World) {}
