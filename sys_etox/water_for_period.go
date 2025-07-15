package sys_etox

// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type CalcWaterForagingPeriod struct {
	time *resource.Tick
	rng  *rand.Rand

	waterPeriodParams *params_etox.WaterForagingPeriod
	waterPeriodData   *globals_etox.WaterForagingPeriodData
	waterNeeds        globals_etox.WaterNeeds
	waterParams       *params_etox.WaterParams
}

func (s *CalcWaterForagingPeriod) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	src := ecs.GetResource[resource.Rand](w)
	s.rng = rand.New(src)

	s.waterParams = ecs.GetResource[params_etox.WaterParams](w)
	s.waterPeriodParams = ecs.GetResource[params_etox.WaterForagingPeriod](w)
	s.waterPeriodData = ecs.GetResource[globals_etox.WaterForagingPeriodData](w)
	s.waterNeeds = globals_etox.WaterNeeds{}
	ecs.AddResource(w, &s.waterNeeds)
}

func (s *CalcWaterForagingPeriod) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		dayOfYear := int((s.time.Tick - 1) % 365)

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
}

func (s *CalcWaterForagingPeriod) Finalize(w *ecs.World) {}
