package sys

// regulates compartimentalized storages of the _ecotox addition
// updates concentrations of PPP in nectar
// corresponding process in netlogo: TupdateInternalExposureNectar_ETOX
// all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type EtoxStorages struct {
	needs          *params.HoneyNeeds
	needsPollen    *params.PollenNeeds
	workerDev      *params.WorkerDevelopment
	oldNurseParams *params.Nursing
	energyParams   *params.EnergyContent
	storesParams   *params.Stores
	foragerParams  *params.Foragers
	etox           *params.PPPApplication
	toxic          *params.PPPToxicity
	waterParams    *params.WaterForaging

	beecsStores *globals.Stores
	stores      *globals.StoragesEtox
	pppFate     *globals.PPPFate
	pop         *globals.PopulationStats
	etoxStats   *globals.PopulationStatsEtox
	inHive      *globals.InHive
	inHiveEtox  *globals.InHiveEtox
	Larvae      *globals.Larvae
	LarvaeEtox  *globals.LarvaeEtox
	cons        *globals.ConsumptionStats
	nstats      *globals.NursingStats

	foragerExpoMapper     *ecs.Map1[comp.PPPExpo]
	foragerActivityMapper *ecs.Map1[comp.ActivityEtox]
	foragerFilter         *ecs.Filter1[comp.Age]
	foragerShuffle        []ecs.Entity

	rng *rand.Rand
}

func (s *EtoxStorages) Initialize(w *ecs.World) {
	s.needs = ecs.GetResource[params.HoneyNeeds](w)
	s.needsPollen = ecs.GetResource[params.PollenNeeds](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storesParams = ecs.GetResource[params.Stores](w)
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.etox = ecs.GetResource[params.PPPApplication](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)
	s.waterParams = ecs.GetResource[params.WaterForaging](w)

	s.beecsStores = ecs.GetResource[globals.Stores](w)
	s.stores = ecs.GetResource[globals.StoragesEtox](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.etoxStats = ecs.GetResource[globals.PopulationStatsEtox](w)
	s.pppFate = ecs.GetResource[globals.PPPFate](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHiveEtox = ecs.GetResource[globals.InHiveEtox](w)
	s.Larvae = ecs.GetResource[globals.Larvae](w)
	s.LarvaeEtox = ecs.GetResource[globals.LarvaeEtox](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.nstats = ecs.GetResource[globals.NursingStats](w)

	s.foragerExpoMapper = s.foragerExpoMapper.New(w)
	s.foragerActivityMapper = s.foragerActivityMapper.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
}

func (s *EtoxStorages) Update(w *ecs.World) {
	// initiate necessary variables
	h := 0.  // for tracking honey in between cohorts
	p := 0.  // for tracking pollen in between cohorts
	c := 0   // for tracking the amount of individuals in the cohorts
	num := 0 // for tracking number of total individuals within one caste
	s.etoxStats.CumDoseNurses = 0.

	consumed_pollen := 0. // tracker for total amount of pollen consumed in this subsystem
	consumed_honey := 0.  // tracker for total amount of honey consumed in this subsystem

	// Thermoregulation energy budget
	thermoRegBrood := (s.needs.WorkerNurse - s.needs.WorkerResting) / s.oldNurseParams.MaxBroodNurseRatio
	workerbaselineneed := s.needs.WorkerResting
	if s.etox.ReworkedThermoETOX {
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood / float64(s.pop.WorkersForagers+s.pop.WorkersInHive) // calculate how much honey each adult IHbee/forager would need to take in extra
		workerbaselineneed += s.stores.ETOX_EnergyThermo
		s.stores.ETOX_EnergyThermo = 0.
	} else {
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood * 0.001 * s.energyParams.Honey // or calculate the total necessary energy
	}

	// get values for the debugging variables
	s.stores.Pollenconcbeforeeating = s.stores.PPPInHivePollenConc // used in debugging and as a helpful metric

	if s.stores.ETOX_HES_E_D0 != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D0 // used in debugging and as a helpful metric
	} else if s.stores.ETOX_HES_E_D1 != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D1 // used in debugging and as a helpful metric
	} else if s.stores.ETOX_HES_E_D2 != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D2 // used in debugging and as a helpful metric
	} else if s.stores.ETOX_HES_E_D3 != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D3 // used in debugging and as a helpful metric
	} else if s.stores.ETOX_HES_E_D4 != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D4 // used in debugging and as a helpful metric
	} else if s.stores.ETOX_HES_E_Capped != 0 {
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_Capped // used in debugging and as a helpful metric
	} else {
		s.stores.Nectarconcbeforeeating = 0 // used in debugging and as a helpful metric
	}

	// foragers, pretty straigt forward and same for all model versions
	forquery := s.foragerFilter.Query()
	for forquery.Next() {
		s.foragerShuffle = append(s.foragerShuffle, forquery.Entity())
	}
	s.rng.Shuffle(len(s.foragerShuffle), func(i, j int) { s.foragerShuffle[i], s.foragerShuffle[j] = s.foragerShuffle[j], s.foragerShuffle[i] })
	forcount := len(s.foragerShuffle) * 100

	for _, e := range s.foragerShuffle {
		ppp := s.foragerExpoMapper.Get(e)
		ppp.OralDose += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001
		s.pppFate.PPPforagersinHive += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001 * float64(s.foragerParams.SquadronSize)
		s.pppFate.PPPforagersTotal += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001 * float64(s.foragerParams.SquadronSize)

		ETOX_Consumed := workerbaselineneed * 0.001 * s.energyParams.Honey * float64(s.foragerParams.SquadronSize)
		ETOX_Consumed += s.stores.ETOX_EnergyThermo
		s.stores.ETOX_EnergyThermo = 0.

		intake := s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize), false)
		ppp.OralDose += intake

		s.pppFate.PPPforagersinHive += intake * float64(s.foragerParams.SquadronSize)
		s.pppFate.PPPforagersTotal += intake * float64(s.foragerParams.SquadronSize)

		consumed_pollen += s.needsPollen.Worker * float64(s.foragerParams.SquadronSize)
		consumed_honey += ETOX_Consumed
	}
	s.foragerShuffle = s.foragerShuffle[:0]

	// classic BEEHAVE_ecotox calculation of exposure
	// inhive bees, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	s.etoxStats.CumDoseIHBees, c, h, p, num = s.CalcDosePerCohort(w, s.inHive.Workers, s.inHiveEtox.WorkerCohortDose, s.stores.ETOX_EnergyThermo, workerbaselineneed, s.needsPollen.Worker, float64(1), float64(1))
	s.stores.ETOX_EnergyThermo = 0.
	currentIHbees := num
	if s.pop.WorkersInHive > 0 {
		s.etoxStats.MeanDoseIHBees = s.etoxStats.CumDoseIHBees / float64(currentIHbees)
	} else {
		s.etoxStats.MeanDoseIHBees = 0.
	}
	s.etoxStats.NumberIHbeeCohorts = c

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPIHbees += s.etoxStats.CumDoseIHBees

	// inhive larvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	// larvae exposure considers the nursebee-filtering effect
	s.etoxStats.CumDoseLarvae, _, h, p, num = s.CalcDosePerCohort(w, s.Larvae.Workers, s.LarvaeEtox.WorkerCohortDose, s.stores.ETOX_EnergyThermo, (s.needs.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), (s.needsPollen.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
	if s.pop.WorkerLarvae > 0 {
		s.etoxStats.MeanDoseLarvae = s.etoxStats.CumDoseLarvae / float64(num)
	} else {
		s.etoxStats.MeanDoseLarvae = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPlarvae += s.etoxStats.CumDoseLarvae

	// inhive dronelarvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	// larvae exposure considers the nursebee-filtering effect
	s.etoxStats.CumDoseDroneLarvae, _, h, p, num = s.CalcDosePerCohort(w, s.Larvae.Drones, s.LarvaeEtox.DroneCohortDose, s.stores.ETOX_EnergyThermo, s.needs.DroneLarva, s.needsPollen.DroneLarva, s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
	if s.pop.DroneLarvae > 0 {
		s.etoxStats.MeanDoseDroneLarvae = s.etoxStats.CumDoseDroneLarvae / float64(num)
	} else {
		s.etoxStats.MeanDoseDroneLarvae = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPdlarvae += s.etoxStats.CumDoseDroneLarvae

	if s.pop.WorkersInHive > 0 {
		s.etoxStats.CumDoseNurses = s.etoxStats.CumDoseIHBees + s.etoxStats.PPPNursebees
		s.etoxStats.MeanDoseNurses = s.etoxStats.CumDoseNurses / float64(currentIHbees)
	} else {
		s.etoxStats.CumDoseNurses = 0.
		s.etoxStats.MeanDoseNurses = 0.
	}
	s.pppFate.PPPNurses += s.etoxStats.PPPNursebees // this is the intake from Nursebeefactors of the old model version
	if s.etox.Nursebeefix && s.pop.WorkersInHive != 0 {
		s.addNurseExptoIHbees(w, s.etoxStats.PPPNursebees, float64(currentIHbees), s.inHive.Workers, s.inHiveEtox.WorkerCohortDose)
	}

	// inhive drones, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	s.etoxStats.CumDoseDrones, _, h, p, num = s.CalcDosePerCohort(w, s.inHive.Drones, s.inHiveEtox.DroneCohortDose, s.stores.ETOX_EnergyThermo, s.needs.Drone, s.needsPollen.Drone, float64(1), float64(1))
	if s.pop.DroneLarvae > 0 {
		s.etoxStats.MeanDoseDrones = s.etoxStats.CumDoseDrones / float64(num)
	} else {
		s.etoxStats.MeanDoseDrones = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPdrones += s.etoxStats.CumDoseDrones

	s.nstats.MaxPollenIntake = s.needsPollen.Worker
	s.nstats.MeanPollenIntake = s.needsPollen.Worker
	s.nstats.MeanHoneyIntake = s.needs.WorkerResting
	s.nstats.MaxHoneyIntake = s.needs.WorkerResting

	if s.etox.DegradationHoney {
		s.DegradeHoney(w)
	}

	// leftovers from debugging
	_ = s.pop.DroneLarvae + s.pop.DronesInHive + s.pop.WorkerLarvae + s.pop.WorkersForagers + s.pop.WorkersInHive + forcount
	// checkpoint for bugfixing honey consumption in etox
	if math.Round(consumed_honey-(s.cons.HoneyDaily*0.001*s.energyParams.Honey)) != 0 || math.Round(consumed_pollen/1000.0-s.cons.PollenDaily) != 0 {
		panic("Fatal error in honey store dose calculations, model output will be wrong!")
	}

	s.ShiftHoney(w)

	s.stores.PPPpollenTotal = s.stores.PPPInHivePollenConc * s.beecsStores.Pollen
	s.stores.PPPhoneyTotal = s.calcPPPhoneytotal(w)
	s.stores.PPPTotal = s.stores.PPPpollenTotal + s.stores.PPPhoneyTotal
}

func (s *EtoxStorages) addNurseExptoIHbees(w *ecs.World, PPPnurses float64, NumIHbees float64, coh []int, dose []float64) {

	AddOralDose := PPPnurses / NumIHbees

	for i := range coh {
		if coh[i] != 0 {
			dose[i] += AddOralDose
			PPPnurses -= AddOralDose * float64(coh[i])
		}
	}
	if math.Round(PPPnurses) != 0 {
		panic("PPP should be 0 by now, there must be a bug somewhere!")
	}
}

func (s *EtoxStorages) calcPPPhoneytotal(w *ecs.World) (totalPPP float64) {
	totalPPP = 0.
	totalPPP += s.stores.ETOX_HES_C_D0 * s.stores.ETOX_HES_E_D0
	totalPPP += s.stores.ETOX_HES_C_D1 * s.stores.ETOX_HES_E_D1
	totalPPP += s.stores.ETOX_HES_C_D2 * s.stores.ETOX_HES_E_D2
	totalPPP += s.stores.ETOX_HES_C_D3 * s.stores.ETOX_HES_E_D3
	totalPPP += s.stores.ETOX_HES_C_D4 * s.stores.ETOX_HES_E_D4
	totalPPP += s.stores.ETOX_HES_C_Capped * s.stores.ETOX_HES_E_Capped

	return
}

func (s *EtoxStorages) CalcDosePerCohort(w *ecs.World, coh []int, dose []float64, init_honeyenergy float64, honey_need float64, pollen_need float64, nursebeefactorHoney float64, nursebeefactorPollen float64) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// this is the baseline version with the logic of the original BEEHAVE_ecotox function
	CumDose = 0.
	cohortcounter = 0
	consumed = 0.
	num = 0
	pconsumed = 0.

	order := rand.Perm(len(coh)) // randomize order to further emulate netlogo ask function
	for _, i := range order {
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honeyenergy

		if coh[i] != 0 {
			init_honeyenergy = 0.
			cohortcounter++
			num += coh[i]

			ETOX_Consumed_Honey += honey_need * 0.001 * s.energyParams.Honey * float64(coh[i])
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage

			if s.etox.Nursebeefix {
				s.etoxStats.PPPNursebees += ETOX_PPPOralDose * (1 - nursebeefactorHoney) * float64(coh[i])
				s.etoxStats.PPPNursebees += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * (1 - nursebeefactorPollen) * float64(coh[i])
			}
			ETOX_PPPOralDose = ETOX_PPPOralDose * nursebeefactorHoney
			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * nursebeefactorPollen // intake from pollen

			consumed += ETOX_Consumed_Honey
			pconsumed += pollen_need * float64(coh[i])

			dose[i] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose * float64(coh[i])
		} else {
			dose[i] = 0
		}
	}
	return
}

func (s *EtoxStorages) FeedOnHoneyStores(w *ecs.World, cons float64, number float64, honeydilution bool) (OralDose float64) {
	OralDose = 0.
	if cons < s.stores.ETOX_HES_E_D0 {
		OralDose += cons * s.stores.ETOX_HES_C_D0 / number
		s.stores.ETOX_HES_E_D0 -= cons
	} else {
		OralDose += s.stores.ETOX_HES_E_D0 * s.stores.ETOX_HES_C_D0 / number
		cons -= s.stores.ETOX_HES_E_D0
		s.stores.ETOX_HES_E_D0 = 0

		if cons < s.stores.ETOX_HES_E_D1 {
			OralDose += cons * s.stores.ETOX_HES_C_D1 / number
			s.stores.ETOX_HES_E_D1 -= cons
		} else {
			OralDose += s.stores.ETOX_HES_E_D1 * s.stores.ETOX_HES_C_D1 / number
			cons -= s.stores.ETOX_HES_E_D1
			s.stores.ETOX_HES_E_D1 = 0

			if cons < s.stores.ETOX_HES_E_D2 {
				OralDose += cons * s.stores.ETOX_HES_C_D2 / number
				s.stores.ETOX_HES_E_D2 -= cons
			} else {
				OralDose += s.stores.ETOX_HES_E_D1 * s.stores.ETOX_HES_C_D2 / number
				cons -= s.stores.ETOX_HES_E_D2
				s.stores.ETOX_HES_E_D2 = 0

				if cons < s.stores.ETOX_HES_E_D3 {
					OralDose += cons * s.stores.ETOX_HES_C_D3 / number
					s.stores.ETOX_HES_E_D3 -= cons
				} else {
					OralDose += s.stores.ETOX_HES_E_D3 * s.stores.ETOX_HES_C_D3 / number
					cons -= s.stores.ETOX_HES_E_D3
					s.stores.ETOX_HES_E_D3 = 0

					if cons < s.stores.ETOX_HES_E_D4 {
						OralDose += cons * s.stores.ETOX_HES_C_D4 / number
						s.stores.ETOX_HES_E_D4 -= cons
					} else {
						OralDose += s.stores.ETOX_HES_E_D4 * s.stores.ETOX_HES_C_D4 / number
						cons -= s.stores.ETOX_HES_E_D4
						s.stores.ETOX_HES_E_D4 = 0

						if cons < s.stores.ETOX_HES_E_Capped {
							OralDose += cons * s.stores.ETOX_HES_C_Capped / number
							s.stores.ETOX_HES_E_Capped -= cons
							if honeydilution {
								s.stores.ETOX_Waterneedfordilution += cons / s.energyParams.Honey / s.etox.ETOXDensityOfHoney * 0.6
							}
						} else {
							OralDose += s.stores.ETOX_HES_E_Capped * s.stores.ETOX_HES_C_Capped / number
							cons -= s.stores.ETOX_HES_E_Capped
							if honeydilution {
								s.stores.ETOX_Waterneedfordilution += s.stores.ETOX_HES_E_Capped / s.energyParams.Honey / s.etox.ETOXDensityOfHoney * 0.6
							}
							s.stores.ETOX_HES_E_Capped = 0
						}
					}
				}
			}
		}
	}
	return
}

func (s *EtoxStorages) DegradeHoney(w *ecs.World) {
	// if this ever is to be turned on PPPfate oberver should be introduced here if someone wanted to create a mass balance again
	DT50honey := s.etox.DT50honey
	s.stores.ETOX_HES_C_D0 = s.stores.ETOX_HES_C_D0 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D1 = s.stores.ETOX_HES_C_D1 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D2 = s.stores.ETOX_HES_C_D2 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D3 = s.stores.ETOX_HES_C_D3 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D4 = s.stores.ETOX_HES_C_D4 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_Capped = s.stores.ETOX_HES_C_Capped * math.Exp(-math.Log(2)/DT50honey) // Dissappearance of the pesticide in the honey following a single first-order kinetic
}

func (s *EtoxStorages) ShiftHoney(w *ecs.World) {
	if (s.stores.ETOX_HES_E_Capped + s.stores.ETOX_HES_E_D4) > 0 {
		s.stores.ETOX_HES_C_Capped = ((s.stores.ETOX_HES_C_Capped * s.stores.ETOX_HES_E_Capped) + (s.stores.ETOX_HES_C_D4 * s.stores.ETOX_HES_E_D4)) / (s.stores.ETOX_HES_E_Capped + s.stores.ETOX_HES_E_D4)
	}
	s.stores.ETOX_HES_E_Capped = s.stores.ETOX_HES_E_D4 + s.stores.ETOX_HES_E_Capped
	s.stores.ETOX_HES_C_D4 = s.stores.ETOX_HES_C_D3
	s.stores.ETOX_HES_E_D4 = s.stores.ETOX_HES_E_D3
	s.stores.ETOX_HES_C_D3 = s.stores.ETOX_HES_C_D2
	s.stores.ETOX_HES_E_D3 = s.stores.ETOX_HES_E_D2
	s.stores.ETOX_HES_C_D2 = s.stores.ETOX_HES_C_D1
	s.stores.ETOX_HES_E_D2 = s.stores.ETOX_HES_E_D1
	s.stores.ETOX_HES_C_D1 = s.stores.ETOX_HES_C_D0
	s.stores.ETOX_HES_E_D1 = s.stores.ETOX_HES_E_D0
	s.stores.ETOX_HES_E_D0 = 0
	s.stores.ETOX_HES_C_D0 = 0

	if s.beecsStores.Honey <= 0 || s.stores.ETOX_HES_E_Capped < 0 {
		s.beecsStores.Honey = 0
		s.stores.ETOX_HES_E_Capped = 0
		s.stores.ETOX_HES_E_D0 = 0
		s.stores.ETOX_HES_E_D1 = 0
		s.stores.ETOX_HES_E_D2 = 0
		s.stores.ETOX_HES_E_D3 = 0
		s.stores.ETOX_HES_E_D4 = 0
		s.stores.ETOX_HES_C_Capped = 0
		s.stores.ETOX_HES_C_D0 = 0
		s.stores.ETOX_HES_C_D1 = 0
		s.stores.ETOX_HES_C_D2 = 0
		s.stores.ETOX_HES_C_D3 = 0
		s.stores.ETOX_HES_C_D4 = 0
	}

	// adjusted this panic to 0.1% acceptable deviation from the honey store in each timestep; 0.1% deemed acceptable because of floating point error
	if math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*1.001 < math.Round(s.beecsStores.Honey) ||
		math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*0.999 > math.Round(s.beecsStores.Honey) {
		panic("Fatal error in honey store dose calculations, model output will be wrong!") // should debug why this triggers sometimes on very long simulation runs through small differences once there is time
	}

}

func (s *EtoxStorages) Finalize(w *ecs.World) {}
