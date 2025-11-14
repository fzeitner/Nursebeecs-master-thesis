package sys_etox

import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/comp_etox"
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

	newCons     *params_etox.ConsumptionRework
	nurseParams *params_etox.Nursing
	larvae      *globals.Larvae
	inHive      *globals.InHive

	nstats   *globals_etox.Nursing_stats
	nglobals *globals_etox.Nursing_globals
	stores   *globals.Stores
	pop      *globals.PopulationStats
	cons     *globals.ConsumptionStats
	time     *resource.Tick

	foragerfilter *ecs.Filter1[comp_etox.Activity_etox]
}

func (s *NurseConsumption) Initialize(w *ecs.World) {
	s.hneeds = ecs.GetResource[params.HoneyNeeds](w)
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)

	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)
	s.nurseParams = ecs.GetResource[params_etox.Nursing](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)

	s.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)

	s.foragerfilter = s.foragerfilter.New(w)
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
		s.nglobals.WLHoney = 0
		for i := 0; i < len(s.larvae.Workers); i++ {
			hneedLarvae += s.newCons.HoneyWorkerLarva[i] * float64(s.larvae.Workers[i])
			if i > 2 {
				s.nglobals.WLHoney += s.newCons.HoneyWorkerLarva[i] * float64(s.larvae.Workers[i]) * s.newCons.HoneyDirect // assumes a fraction of honey gets taken in directly from age 3 onwards
			}
		}
		s.nglobals.Total_honey = hneedLarvae - s.nglobals.WLHoney // other honey needs to be predigested by nurses

		hneedLarvae = 0.
		s.nglobals.DLHoney = 0
		for i := 0; i < len(s.larvae.Drones); i++ {
			hneedLarvae += s.newCons.HoneyDroneLarva[i] * float64(s.larvae.Drones[i])
			if i > 2 {
				s.nglobals.DLHoney += s.newCons.HoneyDroneLarva[i] * float64(s.larvae.Drones[i]) * s.newCons.HoneyDirect // assumes a fraction of honey gets taken in directly from age 3 onwards
			}
		}
		s.nglobals.Total_honey += (hneedLarvae - s.nglobals.DLHoney) // other honey needs to be predigested by nurses

		hneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.HoneyAdultWorker + float64(s.pop.DronesInHive)*s.newCons.HoneyAdultDrone

		hconsumption := hneedAdult + s.nglobals.Total_honey + s.nglobals.WLHoney + s.nglobals.DLHoney + float64(s.pop.TotalBrood)*thermoRegBrood
		consumptionEnergy := 0.001 * hconsumption * s.energyParams.Honey

		s.stores.Honey -= consumptionEnergy
		s.cons.HoneyDaily = hconsumption

		// new pollen consumption
		pneedLarvae := 0.
		s.nglobals.WLPollen = 0
		for i := 0; i < len(s.larvae.Workers); i++ {
			pneedLarvae += s.newCons.PollenWorkerLarva[i] * float64(s.larvae.Workers[i])
			if i > 2 {
				s.nglobals.WLPollen += s.newCons.PollenWorkerLarva[i] * float64(s.larvae.Workers[i]) * s.newCons.PollenDirect // assumes a fraction of pollen gets taken in directly from age 3 onwards
			}
		}
		s.nglobals.Total_pollen = pneedLarvae - s.nglobals.WLPollen // roughly 95% of pollen predigested by nurses

		pneedLarvae = 0.
		s.nglobals.DLPollen = 0
		for i := 0; i < len(s.larvae.Drones); i++ {
			pneedLarvae += s.newCons.PollenDroneLarva[i] * float64(s.larvae.Drones[i])
			if i > 2 {
				s.nglobals.DLPollen += s.newCons.PollenDroneLarva[i] * float64(s.larvae.Drones[i]) * s.newCons.PollenDirect // assumes a fraction of pollen gets taken in directly from age 3 onwards
			}
		}
		s.nglobals.Total_pollen += (pneedLarvae - s.nglobals.DLPollen) // assume 95% of pollen need gets predigested by nurses, technically only 4+ day old larvae any get pollen directly though, so maybe adjust later

		pneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.PollenAdultWorker + float64(s.pop.DronesInHive)*s.newCons.PollenAdultDrone
		DronePriming := 0.
		for i := 0; i < 9; i++ {
			DronePriming += s.newCons.PFPdrone / 9 * float64(s.inHive.Drones[i])
		}
		s.nglobals.Total_pollen += DronePriming // assume that young drones get fed by nurse bees as well, but not the biggest priority when nurtients are scarce --> maybe change

		s.nglobals.WorkerPriming = 0.
		for i := 0; i < 4; i++ {
			s.nglobals.WorkerPriming += s.newCons.PFPworker / 4 * float64(s.inHive.Workers[i]) // assume that young workers get fed by nurses as well. In times of high brood levels young adults do eat pollen themselves already though
		}
		s.nglobals.Total_pollen += s.nglobals.WorkerPriming

		if s.stores.Pollen > s.nglobals.PollenStoreLastDay { // pollen foraging added to pollen stores
			s.nglobals.LastPollenInflux = 0
		} else {
			s.nglobals.LastPollenInflux += 1
		}

		pconsumption := (pneedAdult + s.nglobals.Total_pollen + s.nglobals.WLPollen + s.nglobals.DLPollen) / 1000.0
		s.cons.PollenDaily = pconsumption
		s.stores.Pollen = math.Max(s.stores.Pollen-pconsumption, 0)
		s.stores.IdealPollen = math.Max(pconsumption*float64(s.storeParams.IdealPollenStoreDays), s.storeParams.MinIdealPollenStore)

		s.nglobals.PollenStoreLastDay = s.stores.Pollen // note stores after consumption for next timestep

		// calculate  nurse intake capacity
		TotalNurseCap := 0.
		nonZeroCohort := 0 // to track how old the second oldest IHbee cohort is that is still alive
		s.nstats.TotalNurses = 0
		for i := 4; i <= s.nglobals.NurseAgeMax; i++ { // start at 5 instead of 4; assume there is always at least one cohort of nurses (aged 4 in this case)
			s.nstats.TotalNurses += s.inHive.Workers[i]
			if i != s.nglobals.NurseAgeMax && i != 4 && s.inHive.Workers[i] != 0 {
				nonZeroCohort = i
			}
			TotalNurseCap += float64(s.inHive.Workers[i]) * s.newCons.MaxPollenNurse * s.newCons.Nursingcapabiliies[i]
		}
		s.nstats.NurseFraction = (float64(s.nstats.TotalNurses) / float64(s.pop.TotalAdults)) * 100 // expressed in %
		current_larvae := float64(s.pop.WorkerLarvae + s.pop.DroneLarvae)
		if current_larvae == 0. {
			current_larvae = 1.
		}
		s.nstats.NL_ratio = util.Clamp(float64(s.nstats.TotalNurses)/current_larvae, 0, 10) // give this bounds so that the graph actually tells something
		// quick fix for the beginning of the year without IHbees via making starting foragers winterbees
		if s.nurseParams.StartWinterBees {
			s.nglobals.WinterBees = s.nglobals.WinterBees[:0]
			query := s.foragerfilter.Query() // this is just a bandaid fix to make the model logic work in winterbee time; if there is enough time winterbees could be given actual "jobs"; right now all of them can feed AND forage at the same time
			for query.Next() {
				act := query.Get()
				if act.Winterbee {
					s.nglobals.WinterBees = append(s.nglobals.WinterBees, query.Entity())
				}
			}
		}
		TotalNurseCap += float64(len(s.nglobals.WinterBees)) * 100 * s.newCons.MaxPollenNurse // * Winterbeefactor // could add a consumption capability factor here
		s.nglobals.NurseWorkLoad = s.nglobals.Total_pollen / TotalNurseCap

		// define if the model assumes nurse amount to be sufficient
		s.nglobals.SuffNurses = false // insufficient nurses; this makes young workers eat their own pollen and increaeses the nurse threshold next day
		if s.nglobals.NurseWorkLoad < 1.0 {
			s.nglobals.SuffNurses = true // we have sufficient nurses; this influences if nurses also eat pollen to prime young workers and does not increase nurse threshold next day
		}

		nonZeroCohort += 0                                            // just to be able to leave the code in for now
		if s.nglobals.SuffNurses && len(s.nglobals.WinterBees) == 0 { // is a reduction in the nursing force possible?
			TotalNurseCap_red := TotalNurseCap - float64(s.inHive.Workers[s.nglobals.NurseAgeMax])*s.newCons.MaxPollenNurse*s.newCons.Nursingcapabiliies[s.nglobals.NurseAgeMax] // could also use nonZeroCohort here for reducing to NurseAgeMax to that spot
			if TotalNurseCap_red >= s.nglobals.Total_pollen {
				s.nglobals.Reductionpossible = true
			} else {
				s.nglobals.Reductionpossible = false

			}
		} else {
			s.nglobals.Reductionpossible = false
		}

		if s.time.Tick > 204 {
			a := 1
			a++
		}

		// REWPORK FROM HERE: ProteinFactorNurses
		if s.stores.Pollen > 0 { // REWORK MAYBE NECESSARY; the idea behind this is to simulate a lack of protein based on pollen
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses + 1.0/s.storeParams.ProteinStoreNurse // this still makes sense
		} else {
			/*
				maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
					s.oldNurseParams.MaxBroodNurseRatio // this will probably need to be reworked still
				workLoad := 0.0 // not necessary anymore because workload is now defined directly via forced protein intake of nurses which adresses the same idea
				if maxBrood > 0 {
					workLoad = float64(s.pop.TotalBrood) / maxBrood
				}
				s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses - workLoad/s.storeParams.ProteinStoreNurse // now uses NurseWorkLoad instead of old workLoad which was weirdly dependent on Foragers and thus overall colony size
			*/
			workLoad := util.Clamp(s.nglobals.NurseWorkLoad, 0.0, 10.0)                                            // using values > 1 destabilizes model dynamics too much, maybe look for an alternative solution later
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses - workLoad/s.storeParams.ProteinStoreNurse // now uses NurseWorkLoad instead of old workLoad which was weirdly dependent on Foragers and thus overall colony size
		}
		s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, 1.0)
	}
}

func (s *NurseConsumption) Finalize(w *ecs.World) {}
