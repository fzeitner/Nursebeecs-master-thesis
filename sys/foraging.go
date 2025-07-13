package sys

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_ecotox/comp"
	"github.com/fzeitner/beecs_ecotox/enum/activity"
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/fzeitner/beecs_ecotox/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"gonum.org/v1/gonum/stat/distuv"
)

// Foraging performs the complete foraging process of each day.
// It potentially performs multiple foraging rounds per day.
type Foraging struct {
	rng *rand.Rand

	foragerParams      *params.Foragers
	forageParams       *params.Foraging
	handlingTimeParams *params.HandlingTime
	danceParams        *params.Dance
	energyParams       *params.EnergyContent
	storeParams        *params.Stores
	storesParams       *params.Stores

	waterParams *params.WaterParams
	etox        *params.ETOXparams

	foragePeriod  *globals.ForagingPeriod
	stores        *globals.Stores
	popStats      *globals.PopulationStats
	foragingStats *globals.ForagingStats
	waterneeds    *globals.WaterNeeds

	patches  []patchCandidate
	toRemove []ecs.Entity
	resting  []ecs.Entity
	dances   []ecs.Entity

	patchResourceMapper  *ecs.Map1[comp.Resource]
	patchVisitsMapper    *ecs.Map2[comp.Resource, comp.Visits]
	patchDanceMapper     *ecs.Map2[comp.Resource, comp.Dance]
	patchTripMapper      *ecs.Map1[comp.Trip]
	patchMortalityMapper *ecs.Map1[comp.Mortality]
	patchConfigMapper    *ecs.Map2[comp.PatchProperties, comp.Trip]
	foragerMapper        *ecs.Map2[comp.Activity, comp.KnownPatch]

	activityFilter       *ecs.Filter1[comp.Activity]
	loadFilter           *ecs.Filter2[comp.Activity, comp.NectarLoad]
	foragerFilter        *ecs.Filter3[comp.Activity, comp.KnownPatch, comp.Milage]
	foragerFilterLoadPPP *ecs.Filter5[comp.Activity, comp.KnownPatch, comp.Milage, comp.NectarLoad, comp.PPPExpo]
	foragerFilterSimple  *ecs.Filter2[comp.Activity, comp.KnownPatch]
	patchFilter          *ecs.Filter2[comp.Resource, comp.PatchProperties]
	patchUpdateFilter    *ecs.Filter7[comp.PatchProperties, comp.PatchDistance, comp.Resource, comp.HandlingTime, comp.Trip, comp.Mortality, comp.Dance]

	maxHoneyStore float64
	time          *resource.Tick
}

func (s *Foraging) Initialize(w *ecs.World) {
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.forageParams = ecs.GetResource[params.Foraging](w)
	s.handlingTimeParams = ecs.GetResource[params.HandlingTime](w)
	s.danceParams = ecs.GetResource[params.Dance](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)
	s.storesParams = ecs.GetResource[params.Stores](w)

	s.waterParams = ecs.GetResource[params.WaterParams](w)
	s.etox = ecs.GetResource[params.ETOXparams](w)

	s.popStats = ecs.GetResource[globals.PopulationStats](w)
	s.foragingStats = ecs.GetResource[globals.ForagingStats](w)
	s.foragePeriod = ecs.GetResource[globals.ForagingPeriod](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.waterneeds = ecs.GetResource[globals.WaterNeeds](w)

	s.activityFilter = s.activityFilter.New(w)
	s.loadFilter = s.loadFilter.New(w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.foragerFilterLoadPPP = s.foragerFilterLoadPPP.New(w)
	s.foragerFilterSimple = s.foragerFilterSimple.New(w)
	s.patchFilter = s.patchFilter.New(w)
	s.patchUpdateFilter = s.patchUpdateFilter.New(w)

	s.patchResourceMapper = s.patchResourceMapper.New(w)
	s.patchVisitsMapper = s.patchVisitsMapper.New(w)
	s.patchDanceMapper = s.patchDanceMapper.New(w)
	s.patchTripMapper = s.patchTripMapper.New(w)
	s.patchMortalityMapper = s.patchMortalityMapper.New(w)
	s.patchConfigMapper = s.patchConfigMapper.New(w)
	s.foragerMapper = s.foragerMapper.New(w)

	storeParams := ecs.GetResource[params.Stores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)

	s.maxHoneyStore = storeParams.MaxHoneyStoreKg * 1000.0 * energyParams.Honey
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))

	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *Foraging) Update(w *ecs.World) {
	s.foragingStats.Reset()

	if s.foragePeriod.SecondsToday <= 0 ||
		(s.stores.Honey >= 0.95*s.maxHoneyStore && s.stores.Pollen >= s.stores.IdealPollen) {
		return
	}

	query := s.foragerFilter.Query()
	for query.Next() {
		_, patch, milage := query.Get()
		milage.Today = 0
		patch.VisitedthisDay = false
	}

	if s.time.Tick > 0 {
		hangAroundDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity
		forageProb := s.calcForagingProb()

		// TODO: Lazy winter bees.

		round := 0
		totalDuration := 0.0
		for {
			duration, foragers := s.foragingRound(w, forageProb)
			meanDuration := 0.0
			if foragers > 0 {
				meanDuration = duration / float64(foragers)
			} else {
				meanDuration = hangAroundDuration
			}
			totalDuration += meanDuration

			if totalDuration > float64(s.foragePeriod.SecondsToday) {
				break
			}

			round++
		}

		query = s.foragerFilter.Query()
		for query.Next() {
			act, _, _ := query.Get()
			act.Current = activity.Resting
		}
	}
}

func (s *Foraging) Finalize(w *ecs.World) {}

func (s *Foraging) calcForagingProb() float64 {
	if s.stores.Pollen/s.stores.IdealPollen > 0.5 && s.stores.Honey/s.stores.DecentHoney > 1 {
		return 0
	}
	prob := s.forageParams.ProbBase
	if s.stores.Pollen/s.stores.IdealPollen < 0.2 || s.stores.Honey/s.stores.DecentHoney < 0.5 {
		prob = s.forageParams.ProbHigh
	}
	if s.stores.Honey/s.stores.DecentHoney < 0.2 {
		prob = s.forageParams.ProbEmergency
	}
	return prob
}

func (s *Foraging) foragingRound(w *ecs.World, forageProb float64) (duration float64, foragers int) {
	probCollectPollen := (1.0 - s.stores.Pollen/s.stores.IdealPollen) * s.danceParams.MaxProportionPollenForagers

	if s.stores.Honey/s.stores.DecentHoney < 0.5 {
		probCollectPollen *= s.stores.Honey / s.stores.DecentHoney
	}

	s.updatePatches(w)
	s.decisions(w, forageProb, probCollectPollen)
	s.searching(w)
	s.collecting(w)
	duration, foragers = s.flightCost(w)
	s.mortality(w)
	s.dancing(w)
	s.unloading(w)
	s.countForagers(w)
	return
}

func (s *Foraging) updatePatches(w *ecs.World) {
	query := s.patchUpdateFilter.Query()
	for query.Next() {
		conf, dist, r, ht, trip, mort, dance := query.Get()

		if s.handlingTimeParams.ConstantHandlingTime {
			ht.Pollen = s.handlingTimeParams.PollenGathering
			ht.Nectar = s.handlingTimeParams.NectarGathering
			ht.Water = s.handlingTimeParams.ETOX_handlingTimeWater
		} else {
			ht.Pollen = s.handlingTimeParams.PollenGathering * r.MaxPollen / r.Pollen
			ht.Nectar = s.handlingTimeParams.NectarGathering * r.MaxNectar / r.Nectar
			ht.Water = s.handlingTimeParams.ETOX_handlingTimeWater
		}

		trip.CostNectar = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Nectar *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		trip.CostPollen = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Pollen *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		trip.CostWater = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) // for some reason does not have the same addition as the ones above

		r.EnergyEfficiency = (conf.NectarConcentration*s.foragerParams.NectarLoad*s.energyParams.Sucrose - trip.CostNectar) / trip.CostNectar

		trip.DurationNectar = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Nectar
		trip.DurationPollen = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Pollen
		trip.DurationWater = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Water

		mort.Nectar = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationNectar))
		mort.Pollen = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationPollen))
		mort.Water = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationWater))

		circ := r.EnergyEfficiency*s.danceParams.Slope + s.danceParams.Intercept
		dance.Circuits = util.Clamp(circ, 0, float64(s.danceParams.MaxCircuits))
	}
}

func (s *Foraging) decisions(w *ecs.World, probForage, probCollectPollen float64) {
	ETOX_foragersneeded := 0.
	if s.stores.ETOX_Waterneedfordilution > 0 {
		ETOX_foragersneeded = s.stores.ETOX_Waterneedfordilution / (s.waterParams.ETOX_cropvolume_water * s.waterParams.ETOX_Watertripsperh * float64(s.foragePeriod.SecondsToday) / 3600)
	}
	if s.waterneeds.ETOX_Waterneedforcooling > 0 {
		ETOX_foragersneeded += s.waterneeds.ETOX_Waterneedforcooling / (s.waterParams.ETOX_cropvolume_water * s.waterParams.ETOX_Watertripsperh * float64(s.foragePeriod.SecondsToday) / 3600)
	}
	sumofwaterforagers := 0. // unsure if the original model made a mistake here and actually intended to use the global "ETOX_NumberofWaterforagers" instead, need to check up on this at some point
	ETOX_ProbWatercollection := 0.

	query := s.foragerFilter.Query()
	for query.Next() {
		act, patch, milage := query.Get()

		if s.waterParams.WaterForaging { // doesn´t get accessed atm, because WaterForaging should not be turned on, seems rather irrelevant to the original model as well
			if act.Current != activity.WaterForaging {
				if sumofwaterforagers > ETOX_foragersneeded {
					act.Current = activity.Resting
				}
			}
			if sumofwaterforagers < ETOX_foragersneeded { // This code is super weird; sumofwaterforagers never gets reassigned any value other than 0, checked the original BEEHAVE multiple times, unsure if I need to fix smth here atm
				ETOX_foragersneeded -= sumofwaterforagers
				ETOX_ProbWatercollection = ETOX_foragersneeded / (float64(s.popStats.WorkersForagers)) // BEEHAVE uses SQUADRON_SIZE * count foragerSquadrons, should equate to s.popStats.WorkersForagers

				if s.rng.Float64() < ETOX_ProbWatercollection {
					act.Current = activity.WaterForaging
					act.WaterForager = true
					act.PollenForager = false
				}
			}
		}

		if act.Current != activity.Recruited {
			act.PollenForager = s.rng.Float64() < probCollectPollen
		}

		if act.Current != activity.Recruited &&
			act.Current != activity.Resting &&
			act.Current != activity.Lazy {
			if s.rng.Float64() < s.forageParams.StopProbability {
				act.Current = activity.Resting
			}
		}

		if !act.PollenForager && !patch.Nectar.IsZero() {
			res := s.patchResourceMapper.Get(patch.Nectar)
			if s.rng.Float64() < 1.0/res.EnergyEfficiency &&
				s.rng.Float64() < s.stores.Honey/s.stores.DecentHoney {

				patch.Nectar = ecs.Entity{}
				if act.Current != activity.Resting && act.Current != activity.Lazy {
					act.Current = activity.Searching
				}
			}
		}

		if !patch.Pollen.IsZero() && act.PollenForager {
			trip := s.patchTripMapper.Get(patch.Pollen)
			if s.rng.Float64() < 1-math.Pow(1-s.forageParams.AbandonPollenPerSec, trip.DurationPollen) {
				patch.Nectar = ecs.Entity{}
				if act.Current != activity.Resting && act.Current != activity.Lazy {
					act.Current = activity.Searching
				}
			}
		}

		if act.Current == activity.Resting {
			if s.rng.Float64() < probForage {
				if act.PollenForager {
					if patch.Pollen.IsZero() {
						act.Current = activity.Searching
					} else {
						act.Current = activity.Experienced
					}
				} else {
					if patch.Nectar.IsZero() {
						act.Current = activity.Searching
					} else {
						act.Current = activity.Experienced
					}
				}
			}
		}

		if milage.Today > float32(s.foragerParams.MaxKmPerDay) {
			act.Current = activity.Resting
		}
	}
}

func (s *Foraging) searching(w *ecs.World) {
	cumProb := 0.0
	nonDetectionProb := 1.0

	// TODO: water foraging search here, postponed because module seems to be rather irrelevant

	sz := float64(s.foragerParams.SquadronSize)
	patchQuery := s.patchFilter.Query()
	for patchQuery.Next() {
		res, conf := patchQuery.Get()
		hasNectar := res.Nectar >= s.foragerParams.NectarLoad*sz
		hasPollen := res.Pollen >= s.foragerParams.PollenLoad*sz
		if !hasNectar && !hasPollen {
			continue
		}
		s.patches = append(s.patches, patchCandidate{
			Patch:       patchQuery.Entity(),
			Probability: conf.DetectionProbability,
			HasNectar:   hasNectar,
			HasPollen:   hasPollen,
		})

		cumProb += conf.DetectionProbability
		nonDetectionProb *= 1.0 - conf.DetectionProbability
	}
	detectionProb := 1.0 - nonDetectionProb

	// TODO: shuffle foragers
	foragerQuery := s.foragerFilterSimple.Query()
	for foragerQuery.Next() {
		act, patch := foragerQuery.Get()

		if act.Current == activity.Searching {
			if s.rng.Float64() >= detectionProb {
				continue
			}
			p := s.rng.Float64() * cumProb
			cum := 0.0
			var selected patchCandidate
			for _, pch := range s.patches {
				cum += pch.Probability
				if cum >= p {
					selected = pch
					break
				}
			}
			if act.PollenForager {
				if selected.HasPollen {
					patch.Pollen = selected.Patch
					act.Current = activity.BringPollen
					res, vis := s.patchVisitsMapper.Get(selected.Patch)
					res.Pollen -= s.foragerParams.PollenLoad * sz
					vis.Pollen += s.foragerParams.SquadronSize
				} else {
					patch.Pollen = ecs.Entity{}
				}
			} else {
				if selected.HasNectar {
					patch.Nectar = selected.Patch
					act.Current = activity.BringNectar
					res, vis := s.patchVisitsMapper.Get(selected.Patch)
					res.Nectar -= s.foragerParams.NectarLoad * sz
					vis.Nectar += s.foragerParams.SquadronSize
				} else {
					patch.Nectar = ecs.Entity{}
				}
			}
		}

		if act.Current != activity.Recruited {
			continue
		}

		if !act.PollenForager && !patch.Nectar.IsZero() {
			success := false
			if s.rng.Float64() < s.danceParams.FindProbability {
				res, vis := s.patchVisitsMapper.Get(patch.Nectar)
				if res.Nectar >= s.foragerParams.NectarLoad*sz {
					res.Nectar -= s.foragerParams.NectarLoad * sz
					vis.Nectar += s.foragerParams.SquadronSize
					act.Current = activity.BringNectar
					success = true
				}
			}
			if !success {
				act.Current = activity.Searching
				patch.Nectar = ecs.Entity{}
			}
		}

		if act.PollenForager && !patch.Pollen.IsZero() {
			success := false
			if s.rng.Float64() < s.danceParams.FindProbability {
				res, vis := s.patchVisitsMapper.Get(patch.Pollen)
				if res.Pollen >= s.foragerParams.PollenLoad*sz {
					res.Pollen -= s.foragerParams.PollenLoad * sz
					vis.Pollen += s.foragerParams.SquadronSize
					act.Current = activity.BringPollen
					success = true
				}
			}
			if !success {
				act.Current = activity.Searching
				patch.Pollen = ecs.Entity{}
			}
		}
	}

	s.patches = s.patches[:0]
}

func (s *Foraging) collecting(w *ecs.World) {
	sz := float64(s.foragerParams.SquadronSize)
	foragerQuery := s.foragerFilterLoadPPP.Query()

	// TODO: water collecting here, postponed because water foraging basically irrelevant overall

	for foragerQuery.Next() {
		act, patch, milage, load, PPPexpo := foragerQuery.Get()

		if act.Current == activity.Experienced {
			if act.PollenForager {
				if patch.Pollen.IsZero() {
					act.Current = activity.Resting
				} else {
					res, vis := s.patchVisitsMapper.Get(patch.Pollen)
					if res.Pollen >= s.foragerParams.PollenLoad*sz {
						act.Current = activity.BringPollen
						res.Pollen -= s.foragerParams.PollenLoad * sz
						vis.Pollen += s.foragerParams.SquadronSize
					} else {
						act.Current = activity.Searching
						patch.Pollen = ecs.Entity{}
					}
				}
			} else {
				if patch.Nectar.IsZero() {
					act.Current = activity.Resting
				} else {
					res, vis := s.patchVisitsMapper.Get(patch.Nectar)
					if res.Nectar >= s.foragerParams.NectarLoad*sz {
						act.Current = activity.BringNectar
						res.Nectar -= s.foragerParams.NectarLoad * sz
						vis.Nectar += s.foragerParams.SquadronSize
					} else {
						act.Current = activity.Searching
						patch.Nectar = ecs.Entity{}
					}
				}
			}
		}

		if act.Current == activity.BringNectar {
			conf, trip := s.patchConfigMapper.Get(patch.Nectar)
			load.Energy = conf.NectarConcentration * s.foragerParams.NectarLoad * s.energyParams.Sucrose // --> kJ

			// exposition from nectar foraging
			load.PPPLoad = load.Energy * conf.PPPconcentrationNectar
			PPPexpo.OralDose += load.PPPLoad * s.etox.HSuptake
			load.PPPLoad -= load.PPPLoad * s.etox.HSuptake

			if patch.VisitedthisDay {
				s.foragingStats.ContactExp_repeat++
			} else {
				s.foragingStats.ContactExp_once++
				patch.VisitedthisDay = true
			}

			if PPPexpo.ContactDose > 0 {
				if s.etox.ContactSum {
					PPPexpo.ContactDose += conf.PPPcontactDose
				} else {
					PPPexpo.ContactDose = (PPPexpo.ContactDose + conf.PPPcontactDose) / 2
				}
			} else {
				PPPexpo.ContactDose += conf.PPPcontactDose
			}

			dist := trip.CostNectar / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)
		}

		if act.Current == activity.BringPollen {
			conf, trip := s.patchConfigMapper.Get(patch.Pollen)

			// exposition from pollen foraging
			load.PPPLoad = s.foragerParams.PollenLoad * conf.PPPconcentrationPollen

			if patch.VisitedthisDay {
				s.foragingStats.ContactExp_repeat++
			} else {
				s.foragingStats.ContactExp_once++
				patch.VisitedthisDay = true
			}

			if PPPexpo.ContactDose > 0 {
				if s.etox.ContactSum {
					PPPexpo.ContactDose += conf.PPPcontactDose
				} else {
					PPPexpo.ContactDose = (PPPexpo.ContactDose + conf.PPPcontactDose) / 2
				}
			} else {
				PPPexpo.ContactDose += conf.PPPcontactDose
			}

			dist := trip.CostPollen / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)
		}
	}
}

func (s *Foraging) flightCost(w *ecs.World) (duration float64, foragers int) {
	duration = 0.0
	foragers = 0

	// TODO: flightCost for water foraging here, postponed ...

	query := s.foragerFilterLoadPPP.Query()
	for query.Next() {
		act, patch, milage, _, ppp := query.Get()

		if act.Current == activity.Searching {
			dist := s.forageParams.SearchLength / 1000.0
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			en := s.forageParams.SearchLength * s.foragerParams.FlightCostPerM
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)

			ppp.OralDose += s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize), false)

			duration += s.forageParams.SearchLength / s.foragerParams.FlightVelocity
			foragers++
		} else if act.Current == activity.BringNectar || act.Current == activity.BringPollen {
			en := 0.0
			if act.PollenForager {
				trip := s.patchTripMapper.Get(patch.Pollen)
				duration += trip.DurationPollen + s.handlingTimeParams.PollenUnloading
				en = trip.CostPollen
			} else {
				trip := s.patchTripMapper.Get(patch.Nectar)
				duration += trip.DurationNectar + s.handlingTimeParams.NectarUnloading
				en = trip.CostNectar
			}
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)

			ppp.OralDose += s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize), false)

			foragers++
		} else if s.waterParams.WaterForaging && (act.Current == activity.BringWater || act.Current == activity.UnsuccessfulWater) {
			en := 0.0
			trip := s.patchTripMapper.Get(patch.Water)
			duration += trip.DurationWater + s.handlingTimeParams.NectarUnloading // WRONG; ADJUSTMENT NEEDED original model assumes Nectar unloading time for water as well (or forgot to change)
			en = trip.CostWater

			//if act.Current == activity.UnsuccessfulWater {
			//	patch.Water =                       // have to mark knownWaterPatch here if waterforaging will ever be relevant enough to implement
			//}
			act.Current = activity.WaterForaging
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)

			ppp.OralDose += s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize), false)

			foragers++
		}
	}
	return
}

func (s *Foraging) mortality(w *ecs.World) {
	searchDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity

	// TODO: mortality for water foragers, postponed ..

	foragerQuery := s.foragerFilterLoadPPP.Query()
	for foragerQuery.Next() {
		act, patch, _, _, PPPexpo := foragerQuery.Get()

		// Acute toxicity during flight
		lethaldose := false
		if s.etox.ForagerImmediateMortality {
			if PPPexpo.RdmSurvivalOral < 1-(1/(1+math.Pow(PPPexpo.OralDose/s.etox.ForagerOralLD50, s.etox.ForagerOralSlope))) {
				lethaldose = true
			}
			if PPPexpo.RdmSurvivalContact < 1-(1/(1+math.Pow(PPPexpo.ContactDose/s.etox.ForagerContactLD50, s.etox.ForagerContactSlope))) {
				lethaldose = true
			}
		}

		if act.Current == activity.Searching {
			if s.rng.Float64() < 1-math.Pow(1-s.forageParams.MortalityPerSec, searchDuration) || lethaldose {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
			}
		} else if act.Current == activity.BringNectar {
			m := s.patchMortalityMapper.Get(patch.Nectar)
			if s.rng.Float64() < m.Nectar || lethaldose {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
			}
		} else if act.Current == activity.BringPollen {
			m := s.patchMortalityMapper.Get(patch.Pollen)
			if s.rng.Float64() < m.Pollen || lethaldose {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
			}
		} //else if act.Current == activity.BringWater {}   // postponed because irrelevant

	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.toRemove = s.toRemove[:0]
}

func (s *Foraging) dancing(w *ecs.World) {
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Resting {
			s.resting = append(s.resting, activityQuery.Entity())
		} else if act.Current == activity.BringNectar || act.Current == activity.BringPollen {
			s.dances = append(s.dances, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.resting), func(i, j int) { s.resting[i], s.resting[j] = s.resting[j], s.resting[i] })
	s.rng.Shuffle(len(s.dances), func(i, j int) { s.dances[i], s.dances[j] = s.dances[j], s.dances[i] })

	for _, e := range s.dances {
		act, patch := s.foragerMapper.Get(e)

		if act.Current != activity.BringNectar && act.Current != activity.BringPollen {
			continue
		}

		if act.Current == activity.BringNectar {
			patchRes, dance := s.patchDanceMapper.Get(patch.Nectar)
			danceEEF := patchRes.EnergyEfficiency

			rPoisson := distuv.Poisson{
				Src:    &util.RandWrapper{Src: s.rng},
				Lambda: dance.Circuits * 0.05,
			}
			danceFollowers := int(rPoisson.Rand())

			if danceFollowers == 0 {
				continue
			}
			if len(s.resting) < danceFollowers {
				continue
			}

			for i := 0; i < danceFollowers; i++ {
				follower := s.resting[len(s.resting)-1]
				fAct, fPatch := s.foragerMapper.Get(follower)

				if fPatch.Nectar.IsZero() {
					fPatch.Nectar = patch.Nectar
					fAct.Current = activity.Recruited
					fAct.PollenForager = false
				} else {
					knownRes := s.patchResourceMapper.Get(fPatch.Nectar)
					if danceEEF > knownRes.EnergyEfficiency {
						fPatch.Nectar = patch.Nectar
						fAct.Current = activity.Recruited
						fAct.PollenForager = false
					} else {
						// TODO: really? was resting before!
						fAct.Current = activity.Experienced
					}
				}

				s.resting = s.resting[:len(s.resting)-1]
			}
		}

		if act.Current == activity.BringPollen {
			trip := s.patchTripMapper.Get(patch.Pollen)
			danceTripDuration := trip.DurationPollen

			danceFollowers := s.danceParams.PollenDanceFollowers

			if len(s.resting) < danceFollowers {
				continue
			}

			for i := 0; i < danceFollowers; i++ {
				follower := s.resting[len(s.resting)-1]
				fAct, fPatch := s.foragerMapper.Get(follower)

				if fPatch.Pollen.IsZero() {
					fPatch.Pollen = patch.Pollen
					fAct.Current = activity.Recruited
					fAct.PollenForager = true
				} else {
					knownTrip := s.patchTripMapper.Get(fPatch.Pollen)
					if danceTripDuration < knownTrip.DurationPollen {
						fPatch.Pollen = patch.Pollen
						fAct.Current = activity.Recruited
						fAct.PollenForager = true
					} else {
						// TODO: really? was resting before!
						fAct.Current = activity.Experienced
					}
				}

				s.resting = s.resting[:len(s.resting)-1]
			}
		}
	}

	s.resting = s.resting[:0]
	s.dances = s.dances[:0]
}

func (s *Foraging) unloading(w *ecs.World) {

	// TODO: water unloading, postponed ...

	query := s.loadFilter.Query()
	for query.Next() {
		act, load := query.Get()
		if act.Current == activity.BringNectar {

			s.stores.Honey += load.Energy * float64(s.foragerParams.SquadronSize)

			s.stores.ETOX_HES_C_D0 = ((s.stores.ETOX_HES_C_D0 * s.stores.ETOX_HES_E_D0) + (load.PPPLoad * (1 - s.etox.HSuptake) * float64(s.foragerParams.SquadronSize))) / (s.stores.ETOX_HES_E_D0 + (load.Energy * float64(s.foragerParams.SquadronSize))) // may need to readjust
			s.stores.ETOX_HES_E_D0 += load.Energy * float64(s.foragerParams.SquadronSize)

			// if s.etox.WaterForaging {} // continue here when Water foraging

			if s.stores.Honey > s.maxHoneyStore ||
				s.stores.ETOX_HES_E_Capped+s.stores.ETOX_HES_E_D0+s.stores.ETOX_HES_E_D1+s.stores.ETOX_HES_E_D2+s.stores.ETOX_HES_E_D3+s.stores.ETOX_HES_E_D4 > s.maxHoneyStore {
				s.stores.Honey = s.maxHoneyStore
				s.stores.ETOX_HES_E_D0 = s.maxHoneyStore - (s.stores.ETOX_HES_E_Capped + s.stores.ETOX_HES_E_D1 + s.stores.ETOX_HES_E_D2 + s.stores.ETOX_HES_E_D3 + s.stores.ETOX_HES_E_D4)
			}

			load.Energy = 0.
			load.PPPLoad = 0.
			act.Current = activity.Experienced
		} else if act.Current == activity.BringPollen {
			s.stores.PPPInHivePollenConc = ((s.stores.PPPInHivePollenConc * s.stores.Pollen) + (load.PPPLoad * float64(s.foragerParams.SquadronSize))) / (s.stores.Pollen + s.foragerParams.PollenLoad*float64(s.foragerParams.SquadronSize)) // may need to readjust

			s.stores.Pollen += s.foragerParams.PollenLoad * float64(s.foragerParams.SquadronSize)
			act.Current = activity.Experienced
			load.PPPLoad = 0.
		}
	}
}

func (s *Foraging) countForagers(w *ecs.World) {
	sz := s.foragerParams.SquadronSize
	round := globals.ForagingRound{}

	query := s.activityFilter.Query()
	for query.Next() {
		act := query.Get()

		switch act.Current {
		case activity.Lazy:
			round.Lazy += sz
		case activity.Resting:
			round.Resting += sz
		case activity.Searching:
			round.Searching += sz
		case activity.Recruited:
			round.Recruited += sz
		case activity.WaterForaging:
			round.Water += sz
		case activity.UnsuccessfulWater:
			round.Water += sz
		case activity.Experienced:
			if act.PollenForager {
				round.Pollen += sz
			} else if act.WaterForager {
				round.Water += sz
			} else {
				round.Nectar += sz
			}
		default:
			panic(fmt.Sprintf("forager activity %d invalid at the end of a foraging round", act.Current))
		}
	}

	s.foragingStats.Rounds = append(s.foragingStats.Rounds, round)
}

type patchCandidate struct {
	Patch       ecs.Entity
	Probability float64
	HasNectar   bool
	HasPollen   bool

	HasWater bool
}

// copy from etox_storages_consumption, because I had no simple way or headspace to deal with cyclic import issues atm
func (s *Foraging) FeedOnHoneyStores(w *ecs.World, cons float64, number float64, honeydilution bool) (OralDose float64) {
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
