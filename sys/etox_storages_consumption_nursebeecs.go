package sys

// regulates compartimentalized storages of the _ecotox addition
// updates concentrations of PPP in nectar
// corresponding process in NetLogo: TupdateInternalExposureNectar_ETOX
// and also calculates exposure for nursebeecs

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type EtoxStoragesNbeecs struct {
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
	newCons        *params.ConsumptionRework
	nursing        *params.NursingRework

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
	nglobals    *globals.NursingGlobals
	nstats      *globals.NursingStats

	foragerExpoMapper     *ecs.Map1[comp.PPPExpo]
	foragerActivityMapper *ecs.Map1[comp.ActivityEtox]
	foragerFilter         *ecs.Filter1[comp.Age]
	foragerShuffle        []ecs.Entity

	rng *rand.Rand
}

func (s *EtoxStoragesNbeecs) Initialize(w *ecs.World) {
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
	s.newCons = ecs.GetResource[params.ConsumptionRework](w)
	s.nursing = ecs.GetResource[params.NursingRework](w)

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
	s.nglobals = ecs.GetResource[globals.NursingGlobals](w)
	s.nstats = ecs.GetResource[globals.NursingStats](w)

	s.foragerExpoMapper = s.foragerExpoMapper.New(w)
	s.foragerActivityMapper = s.foragerActivityMapper.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
}

func (s *EtoxStoragesNbeecs) Update(w *ecs.World) {
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
	workerbaselineneed := s.newCons.HoneyAdultWorker
	if s.etox.ReworkedThermoETOX {
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood / float64(s.pop.WorkersForagers+s.pop.WorkersInHive) // calculate how much honey each adult IHbee/forager would need to take in extra
		workerbaselineneed += s.stores.ETOX_EnergyThermo
		s.stores.ETOX_EnergyThermo = 0.
	} else {
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood * 0.001 * s.energyParams.Honey // or calculate the total necessary energy
	}

	// get values for some observing/debugging variables
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

	// forager oral intake, pretty straigt forward and same for all model versions
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

		act := s.foragerActivityMapper.Get(e)
		if act.Winterbee || act.Reverted {
			s.etoxStats.CumDoseNurses += intake * float64(s.foragerParams.SquadronSize)
			s.etoxStats.CumDoseNurses += s.stores.PPPInHivePollenConc * s.newCons.PollenAdultWorker * 0.001 * float64(s.foragerParams.SquadronSize)
		}

		s.pppFate.PPPforagersinHive += intake * float64(s.foragerParams.SquadronSize)
		s.pppFate.PPPforagersTotal += intake * float64(s.foragerParams.SquadronSize)

		consumed_pollen += s.needsPollen.Worker * float64(s.foragerParams.SquadronSize)
		consumed_honey += ETOX_Consumed
	}
	s.foragerShuffle = s.foragerShuffle[:0]

	// new nursebeecs consumption based calculation of exposure
	if s.nglobals.Total_pollen > 0 {
		// these for loops are only relevant to calculate the nursing activities of foragers (winter/reverted); their own needs were calculated already further above
		for _, e := range s.nglobals.WinterBees {
			ppp := s.foragerExpoMapper.Get(e)
			ppp.OralDose += s.stores.PPPInHivePollenConc * s.nglobals.CurrentMaxPollenNurse * s.nglobals.NurseWorkLoad * 0.001 // should not matter because there is seldom a PPP application scenario in winter/early spring anyways

			pollentoeat := s.nglobals.CurrentMaxPollenNurse * s.nglobals.NurseWorkLoad * float64(s.foragerParams.SquadronSize)
			s.etoxStats.CumDoseNurses += pollentoeat * s.stores.PPPInHivePollenConc * 0.001

			honeytoeat := (s.nglobals.Total_honey * (pollentoeat / s.nglobals.Total_pollen)) // consumed honey is calculated via the fraction of total pollen that this squadron ate because NurseWorkLoad is only coupled to consumed pollen

			s.nglobals.Total_honey -= honeytoeat
			s.nglobals.Total_pollen -= pollentoeat

			ETOX_Consumed := honeytoeat * 0.001 * s.energyParams.Honey
			ppphoney := s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize), false)
			ppp.OralDose += ppphoney

			s.pppFate.PPPNurses += ppphoney * float64(s.foragerParams.SquadronSize)
			s.etoxStats.CumDoseNurses += ppphoney * float64(s.foragerParams.SquadronSize)

			consumed_pollen += pollentoeat
			consumed_honey += honeytoeat * 0.001 * s.energyParams.Honey
		}
		for _, e := range s.nglobals.Reverted { // and the reverted foragers here
			ppp := s.foragerExpoMapper.Get(e)
			ppp.OralDose += s.stores.PPPInHivePollenConc * s.nglobals.CurrentMaxPollenNurse * s.nglobals.NurseWorkLoad * 0.001

			pollentoeat := s.nglobals.CurrentMaxPollenNurse * s.nglobals.NurseWorkLoad * float64(s.foragerParams.SquadronSize)
			s.etoxStats.CumDoseNurses += pollentoeat * s.stores.PPPInHivePollenConc * 0.001

			honeytoeat := (s.nglobals.Total_honey * (pollentoeat / s.nglobals.Total_pollen)) // consumed honey is calculated via the fraction of total pollen that this squadron ate because NurseWorkLoad is only coupled to consumed pollen

			s.nglobals.Total_honey -= honeytoeat
			s.nglobals.Total_pollen -= pollentoeat

			ETOX_Consumed := honeytoeat * 0.001 * s.energyParams.Honey
			ppphoney := s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize), false)
			ppp.OralDose += ppphoney

			s.pppFate.PPPNurses += ppphoney * float64(s.foragerParams.SquadronSize)
			s.etoxStats.CumDoseNurses += ppphoney * float64(s.foragerParams.SquadronSize)

			consumed_pollen += pollentoeat
			consumed_honey += honeytoeat * 0.001 * s.energyParams.Honey
		}
	}

	// non-nurse inhive bees first
	ETOX_consumed := s.stores.ETOX_EnergyThermo
	s.stores.ETOX_EnergyThermo = 0.
	s.etoxStats.CumDoseIHBees, c, h, p, s.nstats.NonNurseIHbees = s.CalcDosePerCohortHPGWorkers(w, s.inHive.Workers, s.inHiveEtox.WorkerCohortDose, ETOX_consumed, workerbaselineneed, s.newCons.PollenAdultWorker, s.nglobals.SuffNurses)
	consumed_honey += h
	consumed_pollen += p
	s.etoxStats.NumberIHbeeCohorts = c
	s.pppFate.PPPIHbees += s.etoxStats.CumDoseIHBees

	if s.nstats.NonNurseIHbees != 0 {
		ETOX_consumed = 0.
	}
	// nurse specific consumption here
	IHnurseIntake := 0.
	IHnurseIntake, _, h, p, s.nstats.IHbeeNurses = s.CalcDosePerCohortNursing(w, s.inHive.Workers, s.inHiveEtox.WorkerCohortDose, ETOX_consumed, workerbaselineneed, s.nglobals.Total_honey, s.nglobals.Total_pollen)
	consumed_honey += h
	consumed_pollen += p
	s.etoxStats.CumDoseNurses += IHnurseIntake
	s.pppFate.PPPNurses += IHnurseIntake

	s.etoxStats.MeanDoseIHBees = 0.
	s.etoxStats.MeanDoseNurses = 0.
	if s.nstats.TotalNurses > 0 {
		s.etoxStats.MeanDoseNurses = s.etoxStats.CumDoseNurses / float64(s.nstats.TotalNurses)
	}
	if s.nstats.NonNurseIHbees > 0 {
		s.etoxStats.MeanDoseIHBees = s.etoxStats.CumDoseIHBees / float64(s.nstats.NonNurseIHbees)
	}

	if !s.nglobals.AbortNursing { // no feeding anymore if there are no nurses
		// continue with larvae here
		s.etoxStats.CumDoseLarvae, _, h, p, num = s.CalcDosePerCohortNursingWLarvae(w, s.Larvae.Workers, s.LarvaeEtox.WorkerCohortDose, s.nglobals.WLHoney, s.nglobals.WLPollen)
		if s.pop.WorkerLarvae > 0 {
			s.etoxStats.MeanDoseLarvae = s.etoxStats.CumDoseLarvae / float64(num)
		} else {
			s.etoxStats.MeanDoseLarvae = 0.
		}
		consumed_honey += h
		consumed_pollen += p
		s.pppFate.PPPlarvae += s.etoxStats.CumDoseLarvae

		// and drone larvae
		s.etoxStats.CumDoseDroneLarvae, _, h, p, num = s.CalcDosePerCohortNursingDLarvae(w, s.Larvae.Drones, s.LarvaeEtox.DroneCohortDose, s.nglobals.DLHoney, s.nglobals.DLPollen)
		if s.pop.DroneLarvae > 0 {
			s.etoxStats.MeanDoseDroneLarvae = s.etoxStats.CumDoseDroneLarvae / float64(num)
		} else {
			s.etoxStats.MeanDoseDroneLarvae = 0.
		}
		consumed_honey += h
		consumed_pollen += p
		s.pppFate.PPPdlarvae += s.etoxStats.CumDoseDroneLarvae
	}
	// and drones
	s.etoxStats.CumDoseDrones, _, h, p, num = s.CalcDosePerCohort(w, s.inHive.Drones, s.inHiveEtox.DroneCohortDose, 0, s.newCons.HoneyAdultDrone, s.newCons.PollenAdultDrone, float64(1), float64(1))
	if s.pop.DroneLarvae > 0 {
		s.etoxStats.MeanDoseDrones = s.etoxStats.CumDoseDrones / float64(num)
	} else {
		s.etoxStats.MeanDoseDrones = 0.
	}
	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPdrones += s.etoxStats.CumDoseDrones

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

func (s *EtoxStoragesNbeecs) calcPPPhoneytotal(w *ecs.World) (totalPPP float64) {
	totalPPP = 0.
	totalPPP += s.stores.ETOX_HES_C_D0 * s.stores.ETOX_HES_E_D0
	totalPPP += s.stores.ETOX_HES_C_D1 * s.stores.ETOX_HES_E_D1
	totalPPP += s.stores.ETOX_HES_C_D2 * s.stores.ETOX_HES_E_D2
	totalPPP += s.stores.ETOX_HES_C_D3 * s.stores.ETOX_HES_E_D3
	totalPPP += s.stores.ETOX_HES_C_D4 * s.stores.ETOX_HES_E_D4
	totalPPP += s.stores.ETOX_HES_C_Capped * s.stores.ETOX_HES_E_Capped

	return
}

func (s *EtoxStoragesNbeecs) CalcDosePerCohortNursing(w *ecs.World, coh []int, dose []float64, init_honeyenergy float64, ownHoneyNeed float64, total_honey float64, total_pollen float64) (CumDose float64, Ncohortcounter int, consumed float64, pconsumedtotal float64, num int) {
	// dose calculation for nurses dependent on their own intake and the honey/pollen they additionally eat to provide nutrient secretions to larvae and other young adults
	CumDose = 0.
	Ncohortcounter = 0
	consumed = 0.       // tracker for total honey consumed
	hconsumed := 0.     // tracker for honey consumed only for nursing other bees/larvae
	pconsumedtotal = 0. // tracker for pollen consumed only for nursing other bees/larvae

	for i := 4; i <= s.nglobals.NurseAgeMax; i++ {
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honeyenergy

		if coh[i] != 0 {
			Ncohortcounter++
			num += coh[i]
			init_honeyenergy = 0.

			pconsumed := s.nglobals.CurrentMaxPollenNurse * s.newCons.Nursingcapabiliies[i] * s.nglobals.NurseWorkLoad * float64(coh[i])
			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * 0.001 * (s.nglobals.CurrentMaxPollenNurse*s.newCons.Nursingcapabiliies[i]*s.nglobals.NurseWorkLoad + s.newCons.PollenAdultWorker) // intake from pollen
			pconsumedtotal += (pconsumed + s.newCons.PollenAdultWorker*float64(coh[i]))

			fraction_consumed := 0.
			if total_pollen != 0 {
				fraction_consumed = pconsumed / total_pollen
			}
			hconsumed += fraction_consumed * total_honey

			ETOX_Consumed_Honey += (fraction_consumed*total_honey + ownHoneyNeed*float64(coh[i])) * 0.001 * s.energyParams.Honey
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage
			consumed += ETOX_Consumed_Honey

			dose[i] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose * float64(coh[i])
		} else {
			dose[i] = 0
		}
	}
	total_pollen += s.newCons.PollenAdultWorker * float64(num)
	if math.Round(pconsumedtotal-total_pollen) != 0 || math.Round(hconsumed-total_honey) != 0 {
		panic("Fatal error in dose calculations, model output will be wrong!")
	}
	return
}

func (s *EtoxStoragesNbeecs) CalcDosePerCohortHPGWorkers(w *ecs.World, coh []int, dose []float64, init_honeyenergy float64, honey_need float64, pollen_need float64, SuffNurses bool) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// dose calculation for worker aged <4 days when nurses cannot provide enough protein for them and they eat increased pollen themselves
	CumDose = 0.
	cohortcounter = 0
	consumed = 0. // tracker for total honey consumed
	num = 0
	pconsumed = 0.

	order := rand.Perm(len(coh)) // randomize order to further emulate NetLogo ask function
	for _, i := range order {
		if i >= 4 && i <= s.nglobals.NurseAgeMax { // exclude nurses here
			continue
		}
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honeyenergy
		pollentoeat := pollen_need
		if coh[i] != 0 {
			num += coh[i]
			cohortcounter += 1
			init_honeyenergy = 0.

			ETOX_Consumed_Honey += honey_need * 0.001 * s.energyParams.Honey * float64(coh[i])
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage

			consumed += ETOX_Consumed_Honey

			if i < 4 && !SuffNurses {
				pollentoeat += s.newCons.PFPworker / 4
			}
			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * pollentoeat * 0.001 // intake from pollen
			pconsumed += pollentoeat * float64(coh[i])

			dose[i] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose * float64(coh[i])
		} else {
			dose[i] = 0
		}
	}
	return
}

func (s *EtoxStoragesNbeecs) CalcDosePerCohortNursingWLarvae(w *ecs.World, coh []int, dose []float64, honey float64, pollen float64) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// dose calculation for worker larvae with nursing
	CumDose = 0.
	consumed = 0.
	pconsumed = 0.
	num = 0

	order := rand.Perm(len(coh)) // randomize order to further emulate NetLogo ask function
	for _, i := range order {    // range of a slice iterates over 2 values, first the index and then a copy of the actual val that we use as the index within the for loop
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := 0.
		num += coh[i]

		if coh[i] != 0 {
			cohortcounter += 1
			if i > 2 {

				ETOX_Consumed_Honey += s.newCons.HoneyWorkerLarva[i] * float64(coh[i]) * s.newCons.HoneyDirect * 0.001 * s.energyParams.Honey
				ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage

				consumed += ETOX_Consumed_Honey
				pconsumed += s.newCons.PollenWorkerLarva[i] * s.newCons.PollenDirect * float64(coh[i])

				ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * s.newCons.PollenWorkerLarva[i] * s.newCons.PollenDirect * 0.001 // intake from pollen

				dose[i] = ETOX_PPPOralDose
				CumDose += ETOX_PPPOralDose * float64(coh[i])
			} else {
				dose[i] = 0.
			}
		} else {
			dose[i] = 0.
		}

	}
	if math.Round((consumed/(0.001*s.energyParams.Honey))-honey) != 0 || math.Round(pconsumed-pollen) != 0 {
		panic("Fatal error in dose calculations, model output will be wrong!")
	}
	return
}

func (s *EtoxStoragesNbeecs) CalcDosePerCohortNursingDLarvae(w *ecs.World, coh []int, dose []float64, honey float64, pollen float64) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// dose calculation for drone larvae with nursing
	CumDose = 0.
	consumed = 0.
	pconsumed = 0.
	num = 0

	order := rand.Perm(len(coh)) // randomize order to further emulate NetLogo ask function
	for _, i := range order {    // range of a slice iterates over 2 values, first the index and then a copy of the actual val that we use as the index within the for loop
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := 0.
		num += coh[i]

		if coh[i] != 0 {
			cohortcounter += 1

			if i > 2 {
				ETOX_Consumed_Honey += s.newCons.HoneyDroneLarva[i] * float64(coh[i]) * s.newCons.HoneyDirect * 0.001 * s.energyParams.Honey
				ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage

				consumed += ETOX_Consumed_Honey
				pconsumed += s.newCons.PollenDroneLarva[i] * s.newCons.PollenDirect * float64(coh[i])

				ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * s.newCons.PollenDroneLarva[i] * s.newCons.PollenDirect * 0.001 // intake from pollen

				dose[i] = ETOX_PPPOralDose
				CumDose += ETOX_PPPOralDose * float64(coh[i])
			} else {
				dose[i] = 0.
			}
		} else {
			dose[i] = 0.
		}
	}
	if math.Round((consumed/(0.001*s.energyParams.Honey))-honey) != 0 || math.Round(pconsumed-pollen) != 0 {
		panic("Fatal error in dose calculations, model output will be wrong!")
	}
	return
}

func (s *EtoxStoragesNbeecs) CalcDosePerCohort(w *ecs.World, coh []int, dose []float64, init_honeyenergy float64, honey_need float64, pollen_need float64, nursebeefactorHoney float64, nursebeefactorPollen float64) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// this is the baseline version with the logic of the original BEEHAVE_ecotox function
	CumDose = 0.
	cohortcounter = 0
	consumed = 0.
	num = 0
	pconsumed = 0.

	order := rand.Perm(len(coh)) // randomize order to further emulate NetLogo ask function
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

func (s *EtoxStoragesNbeecs) FeedOnHoneyStores(w *ecs.World, cons float64, number float64, honeydilution bool) (OralDose float64) {
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

func (s *EtoxStoragesNbeecs) DegradeHoney(w *ecs.World) {
	// if this ever is to be turned on PPPfate oberver should be introduced here if someone wanted to create a mass balance again
	DT50honey := s.etox.DT50honey
	s.stores.ETOX_HES_C_D0 = s.stores.ETOX_HES_C_D0 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D1 = s.stores.ETOX_HES_C_D1 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D2 = s.stores.ETOX_HES_C_D2 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D3 = s.stores.ETOX_HES_C_D3 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_D4 = s.stores.ETOX_HES_C_D4 * math.Exp(-math.Log(2)/DT50honey)         // Dissappearance of the pesticide in the honey following a single first-order kinetic
	s.stores.ETOX_HES_C_Capped = s.stores.ETOX_HES_C_Capped * math.Exp(-math.Log(2)/DT50honey) // Dissappearance of the pesticide in the honey following a single first-order kinetic
}

func (s *EtoxStoragesNbeecs) ShiftHoney(w *ecs.World) {
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

func (s *EtoxStoragesNbeecs) Finalize(w *ecs.World) {}
