package sys_etox

// regulates compartimentalized storages of the _ecotox addition
// updates concentrations of PPP in nectar
// corresponding process in netlogo: TupdateInternalExposureNectar_ETOX

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type EtoxStorages struct {
	needs         *params.HoneyNeeds
	needspollen   *params.PollenNeeds
	workerDev     *params.WorkerDevelopment
	nurseParams   *params.Nursing
	energyParams  *params.EnergyContent
	storesParams  *params.Stores
	foragerParams *params.Foragers
	etox          *params_etox.ETOXparams
	toxic         *params_etox.Toxicityparams
	waterParams   *params_etox.WaterParams
	newCons       *params_etox.ConsumptionRework

	beecsstores *globals.Stores
	stores      *globals_etox.Storages_etox
	pop         *globals.PopulationStats
	etoxStats   *globals_etox.PopulationStats_etox
	waterNeeds  *globals_etox.WaterNeeds
	inHive      *globals.InHive
	inHive_etox *globals_etox.InHive_etox
	Larvae      *globals.Larvae
	Larvae_etox *globals_etox.Larvae_etox
	cons        *globals.ConsumptionStats

	foragerMapper  *ecs.Map1[comp_etox.PPPExpo]
	foragerFilter  *ecs.Filter1[comp_etox.PPPExpo]
	foragershuffle []ecs.Entity

	time *resource.Tick
	rng  *rand.Rand
}

func (s *EtoxStorages) Initialize(w *ecs.World) {
	s.needs = ecs.GetResource[params.HoneyNeeds](w)
	s.needspollen = ecs.GetResource[params.PollenNeeds](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.nurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storesParams = ecs.GetResource[params.Stores](w)
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)
	s.waterParams = ecs.GetResource[params_etox.WaterParams](w)
	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)

	s.beecsstores = ecs.GetResource[globals.Stores](w)
	s.stores = ecs.GetResource[globals_etox.Storages_etox](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.etoxStats = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	s.waterNeeds = ecs.GetResource[globals_etox.WaterNeeds](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHive_etox = ecs.GetResource[globals_etox.InHive_etox](w)
	s.Larvae = ecs.GetResource[globals.Larvae](w)
	s.Larvae_etox = ecs.GetResource[globals_etox.Larvae_etox](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)

	s.foragerMapper = s.foragerMapper.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	s.time = ecs.GetResource[resource.Tick](w)
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
}

func (s *EtoxStorages) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		thermoRegBrood := (s.needs.WorkerNurse - s.needs.WorkerResting) / s.nurseParams.MaxBroodNurseRatio
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood * 0.001 * s.energyParams.Honey

		s.stores.Pollenconcbeforeeating = s.stores.PPPInHivePollenConc // used in debugging and as a helpful metric
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D0       // used in debugging and as a helpful metric

		consumed_honey := 0. // tracker for total amount of honey consumed in this subsystem

		// foragers, pretty straigt forward
		forquery := s.foragerFilter.Query()
		for forquery.Next() {
			s.foragershuffle = append(s.foragershuffle, forquery.Entity())
		}
		s.rng.Shuffle(len(s.foragershuffle), func(i, j int) { s.foragershuffle[i], s.foragershuffle[j] = s.foragershuffle[j], s.foragershuffle[i] })

		forcount := len(s.foragershuffle) * 100
		for _, e := range s.foragershuffle {
			ppp := s.foragerMapper.Get(e)
			ppp.OralDose += s.stores.PPPInHivePollenConc * s.needspollen.Worker * 0.001 // * float64(s.foragerParams.SquadronSize)    // original model does not take sz into account, somehow PPPintake from pollen only assumes squadrons/cohorts to intake pollen once and not per individual

			ETOX_Consumed := s.needs.WorkerResting * 0.001 * s.energyParams.Honey * float64(s.foragerParams.SquadronSize)
			ETOX_Consumed += s.stores.ETOX_EnergyThermo
			s.stores.ETOX_EnergyThermo = 0.
			ppp.OralDose += s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize), false)

			consumed_honey += ETOX_Consumed
		}
		s.foragershuffle = s.foragershuffle[:0]

		if s.newCons.Nursebeecs { // if nursebeecs is turned on exosure gets calculated here
			// continue implementing nursebeecs here
		} else { // classic BEEHAVE_ecotox calculation of exposure

			// inhive bees, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
			c := 0
			h := 0.
			ETOX_Consumed_Honey := s.stores.ETOX_EnergyThermo // may already be set to 0 from forager query
			s.stores.ETOX_EnergyThermo = 0.
			s.etoxStats.CumDoseIHBees, c, h = s.CalcDosePerCohort(w, s.inHive.Workers, s.inHive_etox.WorkerCohortDose, ETOX_Consumed_Honey, s.needs.WorkerResting, s.needspollen.Worker, float64(1), float64(1))
			if s.pop.WorkersInHive > 0 {
				s.etoxStats.MeanDoseIHBees = s.etoxStats.CumDoseIHBees / float64(c)
			} else {
				s.etoxStats.MeanDoseIHBees = 0.
			}
			s.etoxStats.NumberIHbeeCohorts = c

			consumed_honey += h

			// inhive larvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
			// larvae exposure considers the nursebee-filtering effect
			ETOX_Consumed_Honey = 0.
			s.etoxStats.CumDoseLarvae, c, h = s.CalcDosePerCohort(w, s.Larvae.Workers, s.Larvae_etox.WorkerCohortDose, ETOX_Consumed_Honey, (s.needs.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), (s.needspollen.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
			if s.pop.WorkerLarvae > 0 {
				s.etoxStats.MeanDoseLarvae = s.etoxStats.CumDoseLarvae / float64(c)
			} else {
				s.etoxStats.MeanDoseLarvae = 0.
			}

			consumed_honey += h

			// inhive dronelarvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
			// larvae exposure considers the nursebee-filtering effect
			s.etoxStats.CumDoseDroneLarvae, c, h = s.CalcDosePerCohort(w, s.Larvae.Drones, s.Larvae_etox.DroneCohortDose, ETOX_Consumed_Honey, s.needs.DroneLarva, s.needspollen.DroneLarva, s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
			if s.pop.DroneLarvae > 0 {
				s.etoxStats.MeanDoseDroneLarvae = s.etoxStats.CumDoseDroneLarvae / float64(c)
			} else {
				s.etoxStats.MeanDoseDroneLarvae = 0.
			}

			consumed_honey += h

			// inhive drones, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
			s.etoxStats.CumDoseDrones, c, h = s.CalcDosePerCohort(w, s.inHive.Drones, s.inHive_etox.DroneCohortDose, ETOX_Consumed_Honey, s.needs.Drone, s.needspollen.Drone, float64(1), float64(1))
			if s.pop.DroneLarvae > 0 {
				s.etoxStats.MeanDoseDrones = s.etoxStats.CumDoseDrones / float64(c)
			} else {
				s.etoxStats.MeanDoseDrones = 0.
			}

			consumed_honey += h

			if s.etox.DegradationHoney {
				s.DegradeHoney(w)
			}

			// leftovers from debugging
			_ = s.pop.DroneLarvae + s.pop.DronesInHive + s.pop.WorkerLarvae + s.pop.WorkersForagers + s.pop.WorkersInHive + forcount
			// checkpoint for bugfixing honey consumption in etox
			if math.Round(consumed_honey) != math.Round(s.cons.HoneyDaily*0.001*s.energyParams.Honey) {
				panic("Fatal error in honey store dose calculations, model output will be wrong!")
			}

			s.ShiftHoney(w)
		}
	}
}

func (s *EtoxStorages) CalcDosePerCohort(w *ecs.World, coh []int, dose []float64, init_honey_need float64, honey_need float64, pollen_need float64, nursebeefactorHoney float64, nursebeefactorPollen float64) (CumDose float64, cohortcounter int, consumed float64) {
	CumDose = 0.
	cohortcounter = 0
	consumed = 0.

	order := rand.Perm(len(coh)) // randomize order to further emulate netlogo ask function
	for _, val := range order {
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honey_need

		if coh[val] != 0 {
			init_honey_need = 0.
			cohortcounter += 1

			ETOX_Consumed_Honey += honey_need * 0.001 * s.energyParams.Honey * float64(coh[val])
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[val]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage
			s.etoxStats.PPPNursebees += ETOX_PPPOralDose * (1 - nursebeefactorHoney)
			ETOX_PPPOralDose = ETOX_PPPOralDose * nursebeefactorHoney

			consumed += ETOX_Consumed_Honey

			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * nursebeefactorPollen // intake from pollen
			s.etoxStats.PPPNursebees += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * (1 - nursebeefactorPollen)

			dose[val] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose
		} else {
			dose[val] = 0
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

	if s.beecsstores.Honey <= 0 || s.stores.ETOX_HES_E_Capped < 0 {
		s.beecsstores.Honey = 0
		s.stores.ETOX_HES_E_Capped = 0
		s.stores.ETOX_HES_E_D0 = 0
		s.stores.ETOX_HES_E_D1 = 0
		s.stores.ETOX_HES_E_D2 = 0
		s.stores.ETOX_HES_E_D3 = 0
		s.stores.ETOX_HES_E_D4 = 0
	}

	// adjusted this panic to 0.1% acceptable deviation from the honey store in each timestep, because simple rounding triggered the error, probably because of floating point error
	if math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*1.001 <= math.Round(s.beecsstores.Honey) ||
		math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*0.999 >= math.Round(s.beecsstores.Honey) {
		panic("Fatal error in honey store dose calculations, model output will be wrong!")
	}

}

func (s *EtoxStorages) Finalize(w *ecs.World) {}
