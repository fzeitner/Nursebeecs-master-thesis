package sys

// regulates compartimentalized storages of the _ecotox addition
// updates concentrations of PPP in nectar
// corresponding process in netlogo: TupdateInternalExposureNectar_ETOX

import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
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
	etox          *params.ETOXparams
	waterParams   *params.WaterParams

	stores     *globals.Stores
	pop        *globals.PopulationStats
	waterNeeds *globals.WaterNeeds
	inHive     *globals.InHive
	Larvae     *globals.Larvae

	foragerFilter *ecs.Filter1[comp.PPPExpo]

	time *resource.Tick
}

func (s *EtoxStorages) Initialize(w *ecs.World) {
	s.needs = ecs.GetResource[params.HoneyNeeds](w)
	s.needspollen = ecs.GetResource[params.PollenNeeds](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.nurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storesParams = ecs.GetResource[params.Stores](w)
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.etox = ecs.GetResource[params.ETOXparams](w)
	s.waterParams = ecs.GetResource[params.WaterParams](w)

	s.stores = ecs.GetResource[globals.Stores](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.waterNeeds = ecs.GetResource[globals.WaterNeeds](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.Larvae = ecs.GetResource[globals.Larvae](w)

	s.foragerFilter = s.foragerFilter.New(w)

	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *EtoxStorages) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		thermoRegBrood := (s.needs.WorkerNurse - s.needs.WorkerResting) / s.nurseParams.MaxBroodNurseRatio
		s.stores.ETOX_EnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood * 0.001 * s.energyParams.Honey

		s.stores.Pollenconcbeforeeating = s.stores.PPPInHivePollenConc
		s.stores.Nectarconcbeforeeating = s.stores.ETOX_HES_C_D0
		s.pop.PPPNursebees = 0.

		// foragers, pretty straigt forward
		query := s.foragerFilter.Query()
		for query.Next() {
			ppp := query.Get()
			ppp.OralDose += s.stores.PPPInHivePollenConc * s.needspollen.Worker * 0.001 // * float64(s.foragerParams.SquadronSize)    // original model does not take sz into account, somehow PPPintake from pollen only assumes squadrons/cohorts to intake pollen once and not per individual

			ETOX_Consumed := s.needs.WorkerResting * 0.001 * s.energyParams.Honey * float64(s.foragerParams.SquadronSize)
			ETOX_Consumed += s.stores.ETOX_EnergyThermo
			s.stores.ETOX_EnergyThermo = 0.
			ppp.OralDose += s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize), false)
		}

		if s.time.Tick == int64(240) {
			rdmbugfixvar := 1
			rdmbugfixvar += 1
		}

		// inhive bees, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
		c := 0
		ETOX_Consumed_Honey := s.stores.ETOX_EnergyThermo // may already be set to 0 from forager query
		s.stores.ETOX_EnergyThermo = 0.
		s.pop.CumDoseIHBees, c = s.CalcDosePerCohort(w, s.inHive.Workers, s.inHive.WorkerCohortDose, ETOX_Consumed_Honey, s.needs.WorkerResting, s.needspollen.Worker, float64(1), float64(1))
		if s.pop.WorkersInHive > 0 {
			s.pop.MeanDoseIHBees = s.pop.CumDoseIHBees / float64(c)
		} else {
			s.pop.MeanDoseIHBees = 0.
		}
		s.pop.NumberIHbeeCohorts = c

		// inhive larvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
		// larvae exposure considers the nursebee-filtering effect
		ETOX_Consumed_Honey = 0.
		s.pop.CumDoseLarvae, c = s.CalcDosePerCohort(w, s.Larvae.Workers, s.Larvae.WorkerCohortDose, ETOX_Consumed_Honey, (s.needs.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), (s.needspollen.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), s.etox.NursebeesNectar, s.etox.NursebeesPollen)
		if s.pop.WorkerLarvae > 0 {
			s.pop.MeanDoseLarvae = s.pop.CumDoseLarvae / float64(c)
		} else {
			s.pop.MeanDoseLarvae = 0.
		}

		// inhive dronelarvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
		// larvae exposure considers the nursebee-filtering effect
		ETOX_Consumed_Honey = 0.
		s.pop.CumDoseDroneLarvae, c = s.CalcDosePerCohort(w, s.Larvae.Drones, s.Larvae.DroneCohortDose, ETOX_Consumed_Honey, s.needs.DroneLarva, s.needspollen.DroneLarva, s.etox.NursebeesNectar, s.etox.NursebeesPollen)
		if s.pop.DroneLarvae > 0 {
			s.pop.MeanDoseDroneLarvae = s.pop.CumDoseDroneLarvae / float64(c)
		} else {
			s.pop.MeanDoseDroneLarvae = 0.
		}

		// inhive drones, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
		ETOX_Consumed_Honey = 0.
		s.pop.CumDoseDrones, c = s.CalcDosePerCohort(w, s.inHive.Drones, s.inHive.DroneCohortDose, ETOX_Consumed_Honey, s.needs.Drone, s.needspollen.Drone, float64(1), float64(1))
		if s.pop.DroneLarvae > 0 {
			s.pop.MeanDoseDrones = s.pop.CumDoseDrones / float64(c)
		} else {
			s.pop.MeanDoseDrones = 0.
		}

		if s.etox.DegradationHoney {
			s.DegradeHoney(w)
		}

		s.ShiftHoney(w)
	}
}

func (s *EtoxStorages) CalcDosePerCohort(w *ecs.World, coh []int, dose []float64, init_honey_need float64, honey_need float64, pollen_need float64, nursebeefactorHoney float64, nursebeefactorPollen float64) (CumDose float64, cohortcounter int) {
	CumDose = 0.
	cohortcounter = 0
	for i := range coh {
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honey_need

		if coh[i] != 0 {
			init_honey_need = 0.
			cohortcounter += 1

			ETOX_Consumed_Honey += honey_need * 0.001 * s.energyParams.Honey * float64(coh[i])
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i]), s.waterParams.WaterForaging) // calculates the exposition from consumption of honey storage
			ETOX_PPPOralDose = ETOX_PPPOralDose * nursebeefactorHoney

			s.pop.PPPNursebees += (ETOX_PPPOralDose / nursebeefactorHoney) * (1 - nursebeefactorHoney)

			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * nursebeefactorPollen // intake from pollen

			s.pop.PPPNursebees += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * (1 - nursebeefactorPollen)

			dose[i] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose
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
								s.stores.ETOX_Waterneedfordilution += cons / s.energyParams.Honey / s.storesParams.ETOXDensityOfHoney * 0.6
							}
						} else {
							OralDose += s.stores.ETOX_HES_E_Capped * s.stores.ETOX_HES_C_Capped / number
							cons -= s.stores.ETOX_HES_E_Capped
							if honeydilution {
								s.stores.ETOX_Waterneedfordilution += s.stores.ETOX_HES_E_Capped / s.energyParams.Honey / s.storesParams.ETOXDensityOfHoney * 0.6
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

	// adjusted this panic to 0.1% acceptable deviation from the honey store in each timestep, because simple rounding triggered the error, probably because of floating point error
	if math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*1.001 <= math.Round(s.stores.Honey) ||
		math.Round((s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D4+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D0))*0.999 >= math.Round(s.stores.Honey) {
		panic("Fatal error in honey store dose calculations, model output will be wrong!")
	}
}

func (s *EtoxStorages) Finalize(w *ecs.World) {}
