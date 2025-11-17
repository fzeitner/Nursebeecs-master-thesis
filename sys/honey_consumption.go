package sys

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// HoneyConsumption calculates the daily honey consumption
// and removes it from the honey store in [globals.Stores].

type HoneyConsumption struct {
	needs        *params.HoneyNeeds
	workerDev    *params.WorkerDevelopment
	nurseParams  *params.Nursing
	energyParams *params.EnergyContent
	storesParams *params.Stores

	stores *globals.Stores
	pop    *globals.PopulationStats
	cons   *globals.ConsumptionStats
	time   *resource.Tick
}

func (s *HoneyConsumption) Initialize(w *ecs.World) {
	s.needs = ecs.GetResource[params.HoneyNeeds](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.nurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storesParams = ecs.GetResource[params.Stores](w)

	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *HoneyConsumption) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		thermoRegBrood := (s.needs.WorkerNurse - s.needs.WorkerResting) / s.nurseParams.MaxBroodNurseRatio
		if s.pop.WorkersInHive+s.pop.WorkersForagers == 0 { // to prevent bugs; if there are no adults there cannot be honey used to warm brood; hive is dead anyways
			thermoRegBrood = 0
		}

		needLarva := s.needs.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)

		needAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.needs.WorkerResting + float64(s.pop.DronesInHive)*s.needs.Drone
		needLarvae := float64(s.pop.WorkerLarvae)*needLarva + float64(s.pop.DroneLarvae)*s.needs.DroneLarva

		if s.pop.WorkersInHive+s.pop.WorkersForagers == 0 { // to prevent bugs; if there are no adults there cannot be honey used to warm brood; hive is dead anyways
			thermoRegBrood = 0
		}

		consumption := needAdult + needLarvae + float64(s.pop.TotalBrood)*thermoRegBrood
		consumptionEnergy := 0.001 * consumption * s.energyParams.Honey

		s.stores.Honey -= consumptionEnergy
		s.cons.HoneyDaily = consumption
	}
}

func (s *HoneyConsumption) Finalize(w *ecs.World) {}
