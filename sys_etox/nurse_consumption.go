package sys_etox

import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// this regulates all consumption of honey and pollen in hive (non-foraging related) if nurse bee module is turned on
// therefore this is equivalent to honey_consumption.go and pollen_consumption.go within the normal beecs model.

type NurseConsumption struct {
	hneeds         *params.HoneyNeeds
	oldNurseParams *params.Nursing
	energyParams   *params.EnergyContent
	storeParams    *params.Stores

	newCons *params_etox.ConsumptionRework
	larvae  *globals.Larvae
	inHive  *globals.InHive

	nglobals *globals_etox.Nursing_globals
	stores   *globals.Stores
	pop      *globals.PopulationStats
	cons     *globals.ConsumptionStats
	time     *resource.Tick
}

func (s *NurseConsumption) Initialize(w *ecs.World) {
	s.hneeds = ecs.GetResource[params.HoneyNeeds](w)
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)

	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)

	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *NurseConsumption) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		// new honey consumption
		// thermoRegBrood stays as an old process for now, might need to change this though
		thermoRegBrood := (s.hneeds.WorkerNurse - s.hneeds.WorkerResting) / s.oldNurseParams.MaxBroodNurseRatio
		if s.pop.WorkersInHive+s.pop.WorkersForagers == 0 { // to prevent bugs; if there are no adults there cannot be honey used to warm brood; hive is dead anyways
			thermoRegBrood = 0
		}
		hneedLarvae := 0.
		for i := 0; i < len(s.larvae.Workers); i++ {
			hneedLarvae += s.newCons.HoneyWorkerLarva[i] * float64(s.larvae.Workers[i])
		}
		s.nglobals.Total_honey = hneedLarvae * 0.95 // assume 95% of pollen predigested by nurses
		s.nglobals.WLHoney = hneedLarvae * 0.05     // 5% of pollen eaten directly
		hneedLarvae = 0.
		for i := 0; i < len(s.larvae.Drones); i++ {
			hneedLarvae += s.newCons.HoneyDroneLarva[i] * float64(s.larvae.Drones[i])
		}
		s.nglobals.DLHoney = hneedLarvae * 0.05      // 5% of pollen eaten directly
		s.nglobals.Total_honey += hneedLarvae * 0.95 // assume 95% of honey need gets predigested by nurses, technically only 4+ day old larvae any get pollen directly though, so maybe adjust later

		hneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.HoneyAdultWorker + float64(s.pop.DronesInHive)*s.newCons.HoneyAdultDrone

		hconsumption := hneedAdult + hneedLarvae + float64(s.pop.TotalBrood)*thermoRegBrood
		consumptionEnergy := 0.001 * hconsumption * s.energyParams.Honey

		s.stores.Honey -= consumptionEnergy
		s.cons.HoneyDaily = hconsumption

		// new pollen consumption
		pneedLarvae := 0.
		for i := 0; i < len(s.larvae.Workers); i++ {
			pneedLarvae += s.newCons.PollenWorkerLarva[i] * float64(s.larvae.Workers[i])
		}
		s.nglobals.Total_pollen = pneedLarvae * 0.95 // assume 95% of pollen predigested by nurses
		s.nglobals.WLPollen = pneedLarvae * 0.05     // 5% of pollen eaten directly
		pneedLarvae = 0.
		for i := 0; i < len(s.larvae.Drones); i++ {
			pneedLarvae += s.newCons.PollenDroneLarva[i] * float64(s.larvae.Drones[i])
		}
		s.nglobals.DLPollen = pneedLarvae * 0.05      // 5% of pollen eaten directly
		s.nglobals.Total_pollen += pneedLarvae * 0.95 // assume 95% of pollen need gets predigested by nurses, technically only 4+ day old larvae any get pollen directly though, so maybe adjust later

		pneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.PollenAdultWorker + float64(s.pop.DronesInHive)*s.newCons.PollenAdultDrone
		for i := 0; i < 9; i++ {
			pneedAdult += s.newCons.PFPdrone / 9 * float64(s.inHive.Drones[i])
			s.nglobals.Total_pollen += s.newCons.PFPdrone / 9 * float64(s.inHive.Drones[i]) // assume that young drones get fed by nurse bees as well, but not the biggest priority when nurtients are scarce --> maybe change
		}
		s.nglobals.WorkerPriming = 0.
		for i := 0; i < 4; i++ {
			s.nglobals.WorkerPriming += s.newCons.PFPworker / 4 * float64(s.inHive.Workers[i]) // assume that young workers get fed by nurses as well. In times of high brood levels young adults do eat pollen themselves already though
			pneedAdult += s.nglobals.WorkerPriming
			s.nglobals.Total_pollen += s.nglobals.WorkerPriming
		}
		pconsumption := (pneedAdult + pneedLarvae) / 1000.0
		s.cons.PollenDaily = pconsumption
		s.stores.Pollen = math.Max(s.stores.Pollen-pconsumption, 0)
		s.stores.IdealPollen = math.Max(pconsumption*float64(s.storeParams.IdealPollenStoreDays), s.storeParams.MinIdealPollenStore)

		// REWPORK FROM HERE: ProteinFactorNurses

		if s.stores.Pollen > 0 { // REWORK STILL NECESSARY
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses + 1.0/s.storeParams.ProteinStoreNurse
		} else {
			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio // this will need to be reworked still
			workLoad := 0.0
			if maxBrood > 0 {
				workLoad = float64(s.pop.TotalBrood) / maxBrood
			}
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses - workLoad/s.storeParams.ProteinStoreNurse
		}
		s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, 1.0)

		// calculate  nurse intake capacity
		TotalNurseCap := 0.
		for i := 4; i >= s.nglobals.NurseAgeMax; i++ {
			TotalNurseCap += float64(s.inHive.Workers[i]) * s.newCons.MaxPollenNurse * s.newCons.Nursingcapabiliies[i]
		}
		s.nglobals.NurseWorkLoad = s.nglobals.Total_pollen / TotalNurseCap

		s.nglobals.SuffNurses = false // insufficient nurses; this makes young workers eat their own pollen and increaeses the nurse threshold next day
		if s.nglobals.NurseWorkLoad < 1.0 {
			s.nglobals.SuffNurses = true // we have sufficient nurses; this influences if nurses also eat pollen to prime young workers and does not increase nurse threshold next day
		}
	}
}

func (s *NurseConsumption) Finalize(w *ecs.World) {}
