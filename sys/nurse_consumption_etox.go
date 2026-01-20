package sys

import (
	"math"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/util"
	"github.com/mlange-42/ark/ecs"
)

// this regulates all consumption of honey and pollen in hive (non-foraging related) for nursebeecs_ecotox;
// therefore this is equivalent to honey_consumption.go and pollen_consumption.go within the beecs_ecotox model.

type NurseConsumptionEtox struct {
	honeyNeeds     *params.HoneyNeeds
	oldNurseParams *params.Nursing
	energyParams   *params.EnergyContent
	storeParams    *params.Stores

	newCons     *params.ConsumptionRework
	nurseParams *params.NursingRework
	toxic       *params.PPPToxicity
	larvae      *globals.Larvae
	inHive      *globals.InHive
	storesETOX  *globals.StoragesEtox

	nStats   *globals.NursingStats
	nGlobals *globals.NursingGlobals
	stores   *globals.Stores
	pop      *globals.PopulationStats
	cons     *globals.ConsumptionStats

	foragerfilter *ecs.Filter1[comp.ActivityEtox]
}

func (s *NurseConsumptionEtox) Initialize(w *ecs.World) {
	s.honeyNeeds = ecs.GetResource[params.HoneyNeeds](w)
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)

	s.newCons = ecs.GetResource[params.ConsumptionRework](w)
	s.nurseParams = ecs.GetResource[params.NursingRework](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.storesETOX = ecs.GetResource[globals.StoragesEtox](w)

	s.nStats = ecs.GetResource[globals.NursingStats](w)
	s.nGlobals = ecs.GetResource[globals.NursingGlobals](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)

	s.foragerfilter = s.foragerfilter.New(w)
}

func (s *NurseConsumptionEtox) Update(w *ecs.World) {
	// reduce maximum pollen intake here if last there was a HPGthreshold exceeded last step
	if s.nurseParams.HGEffects && s.nurseParams.HGFoodIntake {
		if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[2] {
			s.nGlobals.CurrentMaxPollenNurse = s.newCons.MaxPollenNurse * s.toxic.MaxPollenRed[2]
		} else if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[1] {
			s.nGlobals.CurrentMaxPollenNurse = s.newCons.MaxPollenNurse * s.toxic.MaxPollenRed[1]
		} else if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[0] {
			s.nGlobals.CurrentMaxPollenNurse = s.newCons.MaxPollenNurse * s.toxic.MaxPollenRed[0]
		} else {
			s.nGlobals.CurrentMaxPollenNurse = s.newCons.MaxPollenNurse
		}
	} else {
		s.nGlobals.CurrentMaxPollenNurse = s.newCons.MaxPollenNurse
	}

	// start by recalculating nursing metrics and total capacities
	TotalNurseCap, maxpollenpernurse := s.calcNursingMetrics(w)

	// and continue by calculating the total need of honey and pollen.
	// thermoRegBrood stays as an old process for now, this should probably be reworked eventually though (but not within the boundaries of this thesis)
	thermoRegBrood := (s.honeyNeeds.WorkerNurse - s.honeyNeeds.WorkerResting) / s.oldNurseParams.MaxBroodNurseRatio
	if s.pop.WorkersInHive+s.pop.WorkersForagers == 0 { // to prevent bugs; if there are no adults there cannot be honey used to warm brood; hive is dead anyways
		thermoRegBrood = 0
	}

	// initialize local and reset global variables
	s.resetNGlobals(w)
	hneedLarvae := 0.
	pneedLarvae := 0.
	DronePriming := 0.

	// larvae consumption first; gets halted if there are actually no nurses present (large scale death events)
	s.nGlobals.AbortNursing = false
	if s.nStats.TotalNurses == 0 { // means all nurses already died this timestep
		s.nGlobals.AbortNursing = true // stops all subsystems from calculating consumption of larvae
	} else {
		// larval honey and pollen consumption for worker larvae
		for i := 0; i < len(s.larvae.Workers); i++ {
			hneedLarvae += s.newCons.HoneyWorkerLarva[i] * float64(s.larvae.Workers[i])
			pneedLarvae += s.newCons.PollenWorkerLarva[i] * float64(s.larvae.Workers[i])
			if i > 2 {
				s.nGlobals.WLHoney += s.newCons.HoneyWorkerLarva[i] * float64(s.larvae.Workers[i]) * s.newCons.HoneyDirect    // assumes a fraction of honey gets taken in directly from age 3 onwards
				s.nGlobals.WLPollen += s.newCons.PollenWorkerLarva[i] * float64(s.larvae.Workers[i]) * s.newCons.PollenDirect // assumes a fraction of pollen gets taken in directly from age 3 onwards
			}
		}
		s.nGlobals.Total_honey = hneedLarvae - s.nGlobals.WLHoney     // other honey needs to be predigested by nurses
		s.nGlobals.Total_pollen = (pneedLarvae - s.nGlobals.WLPollen) // roughly 95% of pollen predigested by nurses

		// larval honey and pollen consumption for drone larvae
		hneedLarvae = 0.
		pneedLarvae = 0.
		for i := 0; i < len(s.larvae.Drones); i++ {
			hneedLarvae += s.newCons.HoneyDroneLarva[i] * float64(s.larvae.Drones[i])
			pneedLarvae += s.newCons.PollenDroneLarva[i] * float64(s.larvae.Drones[i])
			if i > 2 {
				s.nGlobals.DLHoney += s.newCons.HoneyDroneLarva[i] * float64(s.larvae.Drones[i]) * s.newCons.HoneyDirect    // assumes a fraction of honey gets taken in directly from age 3 onwards
				s.nGlobals.DLPollen += s.newCons.PollenDroneLarva[i] * float64(s.larvae.Drones[i]) * s.newCons.PollenDirect // assumes a fraction of pollen gets taken in directly from age 3 onwards
			}
		}
		s.nGlobals.Total_honey += (hneedLarvae - s.nGlobals.DLHoney)   // other honey needs to be predigested by nurses
		s.nGlobals.Total_pollen += (pneedLarvae - s.nGlobals.DLPollen) // assume 95% of pollen need gets predigested by nurses, technically only 4+ day old larvae any get pollen directly though, so maybe adjust later

		// increased needs of young adult drones
		for i := 0; i < 9; i++ {
			DronePriming += s.newCons.PFPdrone / 9 * float64(s.inHive.Drones[i])
		}
		s.nGlobals.Total_pollen += DronePriming // assume that young drones get fed by nurse bees as well, but not the biggest priority when nurtients are scarce --> maybe change
	}
	// adult honey consumption
	hneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.HoneyAdultWorker + float64(s.pop.DronesInHive)*s.newCons.HoneyAdultDrone

	// calculate total honey consumption and reduce storages
	hconsumption := hneedAdult + s.nGlobals.Total_honey + s.nGlobals.WLHoney + s.nGlobals.DLHoney + float64(s.pop.TotalBrood)*thermoRegBrood
	consumptionEnergy := 0.001 * hconsumption * s.energyParams.Honey

	s.stores.Honey -= consumptionEnergy
	s.cons.HoneyDaily = hconsumption

	// calculate adult pollen consumption
	pneedAdult := float64(s.pop.WorkersInHive+s.pop.WorkersForagers)*s.newCons.PollenAdultWorker + float64(s.pop.DronesInHive)*s.newCons.PollenAdultDrone

	for i := 0; i < 4; i++ {
		s.nGlobals.WorkerPriming += s.newCons.PFPworker / 4 * float64(s.inHive.Workers[i]) // assume that young workers get fed by nurses as well. In times of high brood levels young adults do eat pollen themselves already though
	}
	s.nGlobals.Total_pollen += s.nGlobals.WorkerPriming

	if s.stores.Pollen > s.nGlobals.PollenStoreLastDay { // pollen foraging added to pollen stores
		s.nGlobals.LastPollenInflux = 0
	} else {
		s.nGlobals.LastPollenInflux += 1
	}

	// calculate total pollen consumption and reduce storages
	pconsumption := (pneedAdult + s.nGlobals.Total_pollen + s.nGlobals.WLPollen + s.nGlobals.DLPollen) / 1000.0
	s.cons.PollenDaily = pconsumption
	s.stores.Pollen = math.Max(s.stores.Pollen-pconsumption, 0)
	s.stores.IdealPollen = math.Max(pconsumption*float64(s.storeParams.IdealPollenStoreDays), s.storeParams.MinIdealPollenStore)

	s.nGlobals.PollenStoreLastDay = s.stores.Pollen // note stores after consumption for next timestep

	// calculate NurseWorkLoad
	if TotalNurseCap != 0 {
		s.nGlobals.NurseWorkLoad = s.nGlobals.Total_pollen / TotalNurseCap
	} else {
		s.nGlobals.NurseWorkLoad = 0. //
	}

	// define if the model assumes nurse amount to be sufficient
	s.nGlobals.SuffNurses = false // insufficient nurses; this makes young workers eat their own pollen and increaeses the nurse threshold next day
	if s.nGlobals.NurseWorkLoad < 1.0 && (s.nGlobals.NurseWorkLoad != 0. || float64(s.nStats.WinterBees)/float64(s.nStats.TotalNurses) >= 0.2) {
		s.nGlobals.SuffNurses = true // we have sufficient nurses; this influences if nurses also eat pollen to prime young workers and does not increase nurse threshold next day
	} else {
		s.nGlobals.Total_pollen -= s.nGlobals.WorkerPriming
		if TotalNurseCap != 0. {
			s.nGlobals.NurseWorkLoad = s.nGlobals.Total_pollen / TotalNurseCap
		}
	}
	if s.nStats.TotalNurses != 0 && s.nGlobals.NurseWorkLoad != 0 {
		s.nStats.MeanPollenIntake = s.nGlobals.Total_pollen/float64(s.nStats.TotalNurses) + s.newCons.PollenAdultWorker
		s.nStats.MaxPollenIntake = maxpollenpernurse*s.nGlobals.NurseWorkLoad + s.newCons.PollenAdultWorker
		s.nStats.MeanHoneyIntake = s.nGlobals.Total_honey/float64(s.nStats.TotalNurses) + s.newCons.HoneyAdultWorker
		s.nStats.MaxHoneyIntake = (s.nStats.MaxPollenIntake-s.newCons.PollenAdultWorker)/s.nGlobals.Total_pollen*s.nGlobals.Total_honey + s.newCons.HoneyAdultWorker
	} else {
		s.nStats.MeanPollenIntake = s.newCons.PollenAdultWorker
		s.nStats.MaxPollenIntake = s.newCons.PollenAdultWorker
		s.nStats.MaxHoneyIntake = s.newCons.HoneyAdultWorker
		s.nStats.MeanHoneyIntake = s.newCons.HoneyAdultWorker
	}

	// is a reduction in the nursing force possible?
	s.nGlobals.SquadstoReduce = 0
	s.nGlobals.Reductionpossible = false

	foragernursefrac := 0.
	if s.nStats.TotalNurses > 0 {
		foragernursefrac = float64(s.nStats.WinterBees+s.nStats.RevertedForagers) / float64(s.nStats.TotalNurses)
	}
	if s.nGlobals.SuffNurses && foragernursefrac <= 0.1 && s.nStats.RevertedForagers == 0 { // is a reduction in the nursing force possible?
		TotalNurseCap_red := TotalNurseCap - float64(s.inHive.Workers[s.nGlobals.NurseAgeMax])*s.newCons.MaxPollenNurse*s.newCons.Nursingcapabiliies[s.nGlobals.NurseAgeMax] // could also use nonZeroCohort here for reducing to NurseAgeMax to that spot
		if TotalNurseCap_red >= s.nGlobals.Total_pollen {
			s.nGlobals.Reductionpossible = true
		}
	} else if s.nGlobals.SuffNurses && s.nStats.RevertedForagers != 0 {
		toReduce := TotalNurseCap - s.nGlobals.Total_pollen
		for i := 0; i <= s.nStats.RevertedForagers/100; i++ {
			toReduce -= s.newCons.MaxPollenNurse * 100
			if toReduce > 0 {
				s.nGlobals.SquadstoReduce++
			}
			if toReduce < 0 {
				break
			}
		}
	}

	// REWORKED: ProteinFactorNurses
	if s.nurseParams.Nursebeecsv1 {
		if s.stores.Pollen > 0 { // REWORKED to use NurseWorkload instead of overall colony size including foragers
			threshold := util.Clamp(s.nGlobals.NurseWorkLoad, s.nurseParams.MinimumTH, s.nurseParams.NurseWorkLoadTH)
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses+(threshold-s.nGlobals.NurseWorkLoad)/s.storeParams.ProteinStoreNurse, 0.0, 1.0) // increase of reservoir dependent on workload as well
		} else if s.stores.Pollen <= 0 {
			workLoad := util.Clamp(s.nGlobals.NurseWorkLoad, 0.0, 5.0)
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses-workLoad/s.storeParams.ProteinStoreNurse, 0.0, 1.0) // now uses NurseWorkLoad instead of old workLoad which was weirdly dependent on Foragers and thus overall colony size
		}
	} else if s.nurseParams.Nursebeecsv0 {
		// old version that leads to large fluctuations in larval abundance
		if s.stores.Pollen > 0 {
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses + (s.nurseParams.NurseWorkLoadTH-s.nGlobals.NurseWorkLoad)/s.storeParams.ProteinStoreNurse
		} else if s.stores.Pollen <= 0 {
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses - s.nGlobals.NurseWorkLoad/s.storeParams.ProteinStoreNurse // now uses NurseWorkLoad instead of old workLoad which was weirdly dependent on Foragers and thus overall colony size
		}
	} else { // old ProteinFactorNurses; dependent on total colony size
		if s.stores.Pollen > 0 {
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses + 1.0/s.storeParams.ProteinStoreNurse // this still makes sense
		} else {
			maxBrood := (float64(s.pop.WorkersInHive) + float64(s.pop.WorkersForagers)*s.oldNurseParams.ForagerNursingContribution) *
				s.oldNurseParams.MaxBroodNurseRatio
			workLoad := 0.0
			if maxBrood > 0 {
				workLoad = float64(s.pop.TotalBrood) / maxBrood
			}
			s.stores.ProteinFactorNurses = s.stores.ProteinFactorNurses - workLoad/s.storeParams.ProteinStoreNurse
		}
		s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, 1.0)
	}

	// effects on ProteinFactorNurses by reduced HG-activity based on Schott et al. 2021
	if s.nurseParams.HGEffects && !s.nurseParams.HGFoodIntake {
		if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[2] {
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, s.toxic.ProteinFactorNurseExposed[2])
		} else if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[1] {
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, s.toxic.ProteinFactorNurseExposed[1])
		} else if s.storesETOX.Nectarconcbeforeeating >= s.toxic.HGthreshold[0] {
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, s.toxic.ProteinFactorNurseExposed[0])
		} else {
			s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, 1.0)
		}
	} else {
		s.stores.ProteinFactorNurses = util.Clamp(s.stores.ProteinFactorNurses, 0.0, 1.0)
	}

}

func (s *NurseConsumptionEtox) Finalize(w *ecs.World) {}

func (s *NurseConsumptionEtox) calcNursingMetrics(w *ecs.World) (nursingcap float64, maxpollenpernurse float64) {
	s.nStats.IHbeeNurses = 0
	s.nStats.WinterBees = 0
	s.nStats.RevertedForagers = 0
	maxpollenpernurse = 0.
	for i := 4; i <= s.nGlobals.NurseAgeMax; i++ {
		s.nStats.IHbeeNurses += s.inHive.Workers[i]
		nursingcap += float64(s.inHive.Workers[i]) * s.nGlobals.CurrentMaxPollenNurse * s.newCons.Nursingcapabiliies[i]

		if s.inHive.Workers[i] != 0 && s.nGlobals.CurrentMaxPollenNurse*s.newCons.Nursingcapabiliies[i] > maxpollenpernurse { // just to get some nursing stats for analysis and debugging
			maxpollenpernurse = s.nGlobals.CurrentMaxPollenNurse * s.newCons.Nursingcapabiliies[i]
		}
	}
	// quick fix for the beginning and end of the year without IHbees via making starting foragers winterbees; also includes reverted foragers that might occur in times of great nursing needs or after mass death events
	s.nGlobals.WinterBees = s.nGlobals.WinterBees[:0]
	s.nGlobals.Reverted = s.nGlobals.Reverted[:0]
	query := s.foragerfilter.Query() // this is just a bandaid fix to make the model logic work in winterbee time; if there is enough time winterbees could be given actual "jobs"; right now all of them can feed AND forage at the same time
	for query.Next() {
		act := query.Get()
		if act.Winterbee {
			s.nStats.WinterBees += 100
			s.nGlobals.WinterBees = append(s.nGlobals.WinterBees, query.Entity())
		} else if act.Reverted {
			s.nStats.RevertedForagers += 100
			s.nGlobals.Reverted = append(s.nGlobals.Reverted, query.Entity())
		}
	}
	if s.nStats.WinterBees != 0 {
		maxpollenpernurse = s.nGlobals.CurrentMaxPollenNurse // winterbees always work at 100% for now; should be rethought down the line
	}
	nursingcap += float64(s.nStats.RevertedForagers+s.nStats.WinterBees) * s.nGlobals.CurrentMaxPollenNurse
	s.nStats.TotalNurses = s.nStats.IHbeeNurses + s.nStats.RevertedForagers + s.nStats.WinterBees // maybe ignore winterbees here as they are a bit of a special case?
	s.nStats.NurseFraction = (float64(s.nStats.TotalNurses) / float64(s.pop.WorkersInHive+s.pop.WorkersForagers))

	s.nStats.NonNurseIHbees = 0
	for i := 0; i < 4; i++ {
		s.nStats.NonNurseIHbees += s.inHive.Workers[i]
	}
	for i := s.nGlobals.NurseAgeMax + 1; i < len(s.inHive.Workers); i++ {
		s.nStats.NonNurseIHbees += s.inHive.Workers[i]
	}
	current_larvae := float64(s.pop.WorkerLarvae + s.pop.DroneLarvae)
	if current_larvae == 0. {
		current_larvae = 1.
	}
	s.nStats.NL_ratio = util.Clamp(float64(s.nStats.TotalNurses)/current_larvae, 0, 10) // give this bounds so that the graph actually tells something

	return
}

func (s *NurseConsumptionEtox) resetNGlobals(w *ecs.World) {
	s.nGlobals.Total_pollen = 0
	s.nGlobals.WLPollen = 0
	s.nGlobals.DLPollen = 0
	s.nGlobals.Total_honey = 0.
	s.nGlobals.WLHoney = 0
	s.nGlobals.DLHoney = 0
	s.nGlobals.WorkerPriming = 0.
}
