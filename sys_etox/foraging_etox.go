package sys_etox

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"gonum.org/v1/gonum/stat/distuv"
)

// Foraging performs the complete foraging process of each day.
// It potentially performs multiple foraging rounds per day.
type Foraging_etox struct {
	rng  *rand.Rand
	time *resource.Tick

	foragerParams      *params.Foragers
	forageParams       *params.Foraging
	handlingTimeParams *params.HandlingTime
	danceParams        *params.Dance
	energyParams       *params.EnergyContent
	storeParams        *params.Stores
	storesParams       *params.Stores

	waterParams   *params_etox.WaterParams
	etox          *params_etox.ETOXparams
	toxic         *params_etox.Toxicityparams
	guts          *params_etox.GUTSParams
	nursingParams *params_etox.Nursing

	foragePeriod  *globals.ForagingPeriod
	stores        *globals.Stores
	stores_etox   *globals_etox.Storages_etox
	foragingStats *globals_etox.ForagingStats_etox
	//waterneeds    *globals_etox.WaterNeeds
	pop        *globals.PopulationStats
	newCohorts *globals.NewCohorts
	aff        *globals.AgeFirstForaging
	factory    *globals.ForagerFactory

	patches        []patchCandidate_etox
	toRemove       []ecs.Entity
	resting        []ecs.Entity
	foragershuffle []ecs.Entity
	dances         []ecs.Entity
	toAdd          []ecs.Entity

	patchResourceMapper   *ecs.Map1[comp.Resource]
	ageMapper             *ecs.Map1[comp.Age]
	patchVisitsMapper     *ecs.Map2[comp.Resource, comp.Visits]
	patchDanceMapper      *ecs.Map2[comp.Resource, comp.Dance]
	patchTripMapper       *ecs.Map1[comp.Trip]
	patchMortalityMapper  *ecs.Map1[comp.Mortality]
	patchConfigMapper     *ecs.Map2[comp.PatchProperties, comp.Trip]
	patchConfigMapperEtox *ecs.Map3[comp.PatchProperties, comp_etox.PatchProperties_etox, comp.Trip]
	foragerMapper         *ecs.Map2[comp_etox.Activity_etox, comp_etox.KnownPatch_etox]
	foragerLoadPPPMapper  *ecs.Map6[comp_etox.Activity_etox, comp_etox.KnownPatch_etox, comp.Milage, comp.NectarLoad, comp_etox.EtoxLoad, comp_etox.PPPExpo]
	pppExpoAdder          *ecs.Map2[comp_etox.PPPExpo, comp_etox.EtoxLoad]
	etoxPatchAdder        *ecs.Map2[comp_etox.KnownPatch_etox, comp_etox.Activity_etox]

	activityFilter       *ecs.Filter1[comp_etox.Activity_etox]
	ageFilter            *ecs.Filter1[comp.Age]
	loadFilter           *ecs.Filter3[comp_etox.Activity_etox, comp.NectarLoad, comp_etox.EtoxLoad]
	foragerFilter        *ecs.Filter3[comp_etox.Activity_etox, comp_etox.KnownPatch_etox, comp.Milage]
	foragerFilterLoadPPP *ecs.Filter6[comp_etox.Activity_etox, comp_etox.KnownPatch_etox, comp.Milage, comp.NectarLoad, comp_etox.EtoxLoad, comp_etox.PPPExpo]
	foragerFilterSimple  *ecs.Filter2[comp_etox.Activity_etox, comp_etox.KnownPatch_etox]
	patchFilter          *ecs.Filter2[comp.Resource, comp.PatchProperties]
	patchUpdateFilter    *ecs.Filter7[comp.PatchProperties, comp.PatchDistance, comp.Resource, comp.HandlingTime, comp.Trip, comp.Mortality, comp.Dance]

	maxHoneyStore float64
}

func (s *Foraging_etox) Initialize(w *ecs.World) {
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.forageParams = ecs.GetResource[params.Foraging](w)
	s.handlingTimeParams = ecs.GetResource[params.HandlingTime](w)
	s.danceParams = ecs.GetResource[params.Dance](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)
	s.storesParams = ecs.GetResource[params.Stores](w)

	s.waterParams = ecs.GetResource[params_etox.WaterParams](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)
	s.guts = ecs.GetResource[params_etox.GUTSParams](w)
	s.nursingParams = ecs.GetResource[params_etox.Nursing](w)

	s.foragingStats = ecs.GetResource[globals_etox.ForagingStats_etox](w)
	s.foragePeriod = ecs.GetResource[globals.ForagingPeriod](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.stores_etox = ecs.GetResource[globals_etox.Storages_etox](w)
	//s.waterneeds = ecs.GetResource[globals_etox.WaterNeeds](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.factory = ecs.GetResource[globals.ForagerFactory](w)

	s.activityFilter = s.activityFilter.New(w)
	s.ageFilter = s.ageFilter.New(w)
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
	s.patchConfigMapperEtox = s.patchConfigMapperEtox.New(w)
	s.foragerMapper = s.foragerMapper.New(w)
	s.ageMapper = s.ageMapper.New(w)
	s.foragerLoadPPPMapper = s.foragerLoadPPPMapper.New(w)
	s.pppExpoAdder = s.pppExpoAdder.New(w)
	s.etoxPatchAdder = s.etoxPatchAdder.New(w)

	storeParams := ecs.GetResource[params.Stores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)

	s.maxHoneyStore = storeParams.MaxHoneyStoreKg * 1000.0 * energyParams.Honey
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *Foraging_etox) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		s.foragingStats.Reset()

		s.newForagers(w) // here the foragers get initialized now; mimics BEEHAVE exactly.

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

		hangAroundDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity
		forageProb := s.calcForagingProb()
		s.foragingStats.Prob = forageProb // added these for debugging

		// TODO: Lazy winter bees.
		s.stores.DecentHoney = math.Max(float64(s.pop.WorkersInHive+s.pop.WorkersForagers), 1) * s.storesParams.DecentHoneyPerWorker * s.energyParams.Honey // added this because counting proc happens in between last decent honey calc and now --> recalc necessary
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

			s.foragingStats.SumDur = totalDuration // added these for debugging

			if totalDuration >= float64(s.foragePeriod.SecondsToday) {
				break
			}

			round++
		}
		query2 := s.loadFilter.Query() // changed this to a query that also checks load to track mean and max energy expended by foragers per day
		c := query2.Count()
		for query2.Next() {
			act, _, load := query2.Get()
			act.Current = activity.Resting

			if act.Reverted {
				continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
			}

			s.foragingStats.MeanEnergyExpenditure += load.EnergyUsed
			if load.EnergyUsed > s.foragingStats.MaxEnergyExpenditure {
				s.foragingStats.MaxEnergyExpenditure = load.EnergyUsed
			}
		}
		s.foragingStats.MeanEnergyExpenditure /= float64(c)
	}
}

func (s *Foraging_etox) Finalize(w *ecs.World) {}

func (s *Foraging_etox) newForagers(w *ecs.World) {
	if s.newCohorts.Foragers > 0 {
		s.factory.CreateSquadrons(s.newCohorts.Foragers, int(s.time.Tick-1)-s.aff.Aff)
	}
	s.newCohorts.Foragers = 0
	// adding etox components to the newly initialized forager entities
	agequery := s.ageFilter.Without(ecs.C[comp_etox.PPPExpo]()).Query() // this way only newly created squadrons get called by this query
	for agequery.Next() {
		s.toAdd = append(s.toAdd, agequery.Entity())
	}
	for _, e := range s.toAdd {
		if s.etox.GUTS && s.guts.Type == "IT" {
			s.pppExpoAdder.Add(e, &comp_etox.PPPExpo{OralDose: s.newCohorts.NewForOralDose, ContactDose: 0., C_i: s.newCohorts.NewForC_i, RmdSurvivalIT: s.newCohorts.NewForITthreshold}, &comp_etox.EtoxLoad{PPPLoad: 0., EnergyUsed: 0.})
		} else if s.etox.GUTS && s.guts.Type == "SD" {
			s.pppExpoAdder.Add(e, &comp_etox.PPPExpo{OralDose: s.newCohorts.NewForOralDose, ContactDose: 0., C_i: s.newCohorts.NewForC_i}, &comp_etox.EtoxLoad{PPPLoad: 0., EnergyUsed: 0.})
		} else {
			s.pppExpoAdder.Add(e, &comp_etox.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: s.rng.Float64(), RdmSurvivalOral: s.rng.Float64()}, &comp_etox.EtoxLoad{PPPLoad: 0., EnergyUsed: 0.})
		}

		squadAge := s.ageMapper.Get(e)
		if s.nursingParams.StartWinterBees && squadAge.DayOfBirth >= 205 && squadAge.DayOfBirth < 265 { // original BEEHAVE assumes starting foragers (=winter bees) are aged 100 - 160 days already; Aff + 21 = current age of the cohort; 21 = dev-time from egg - adult; Aff = adult time before foraging
			if s.rng.Float64() <= float64(1)/float64(60)*float64(squadAge.DayOfBirth-204) { // assume linear increase in likelihood to turn into winterbees
				s.etoxPatchAdder.Add(e, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting, Winterbee: true}) // assumes bees turning into foragers are winterbees again;
			} else {
				s.etoxPatchAdder.Add(e, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting})
			}
		} else if s.nursingParams.StartWinterBees && squadAge.DayOfBirth >= 265 { // original BEEHAVE assumes starting foragers are aged 100 - 160 days already !!!; this is just an estimate though, it would make a lot more sense to couple this to pop dynamic and nectar/pollen influxes
			s.etoxPatchAdder.Add(e, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting, Winterbee: true}) // assumes bees turning into foragers are winterbees again
			// aligns with literature assuming eggs from august - september start turning into winterbees (21 days for theses eggs to turn into IHbees + some more to turn into foragers --> roughly start of october)
			// there should eventually be a system introduced to actually differentiate between winterbees and summeerbees properly (mortalities, food demands, chance from egg onwards to turn into 1 of the 2, ...)
		} else {
			s.etoxPatchAdder.Add(e, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting})
		}
	}
	s.toAdd = s.toAdd[:0]
	if s.etox.GUTS {
		if s.guts.Type == "IT" {
			s.newCohorts.NewForITthreshold = 0.
		}
		s.newCohorts.NewForC_i = 0.
		s.newCohorts.NewForOralDose = 0.
	}
}

func (s *Foraging_etox) calcForagingProb() float64 {
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

func (s *Foraging_etox) foragingRound(w *ecs.World, forageProb float64) (duration float64, foragers int) {
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

func (s *Foraging_etox) updatePatches(w *ecs.World) {
	query := s.patchUpdateFilter.Query()
	for query.Next() {
		conf, dist, r, ht, trip, mort, dance := query.Get()

		if s.handlingTimeParams.ConstantHandlingTime {
			ht.Pollen = s.handlingTimeParams.PollenGathering
			ht.Nectar = s.handlingTimeParams.NectarGathering
		} else {
			ht.Pollen = s.handlingTimeParams.PollenGathering * r.MaxPollen / r.Pollen
			ht.Nectar = s.handlingTimeParams.NectarGathering * r.MaxNectar / r.Nectar
		}

		trip.CostNectar = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Nectar *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		trip.CostPollen = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Pollen *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		r.EnergyEfficiency = (conf.NectarConcentration*s.foragerParams.NectarLoad*s.energyParams.Sucrose - trip.CostNectar) / trip.CostNectar

		trip.DurationNectar = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Nectar
		trip.DurationPollen = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Pollen

		mort.Nectar = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationNectar))
		mort.Pollen = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationPollen))

		circ := r.EnergyEfficiency*s.danceParams.Slope + s.danceParams.Intercept
		dance.Circuits = util.Clamp(circ, 0, float64(s.danceParams.MaxCircuits))
	}
}

func (s *Foraging_etox) decisions(w *ecs.World, probForage, probCollectPollen float64) {
	query := s.foragerFilter.Query()
	for query.Next() {
		act, patch, milage := query.Get()

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
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

func (s *Foraging_etox) searching(w *ecs.World) {
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
		s.patches = append(s.patches, patchCandidate_etox{
			Patch:       patchQuery.Entity(),
			Probability: conf.DetectionProbability,
			HasNectar:   hasNectar,
			HasPollen:   hasPollen,
		})

		cumProb += conf.DetectionProbability
		nonDetectionProb *= 1.0 - conf.DetectionProbability
	}
	detectionProb := 1.0 - nonDetectionProb

	// TODO: test this foragershuffle
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Searching || act.Current == activity.Recruited {
			s.foragershuffle = append(s.foragershuffle, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.foragershuffle), func(i, j int) { s.foragershuffle[i], s.foragershuffle[j] = s.foragershuffle[j], s.foragershuffle[i] })

	for _, e := range s.foragershuffle {
		act, patch := s.foragerMapper.Get(e)

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
		}

		s.foragingStats.TotalSearches += 1

		if act.Current == activity.Searching {
			if s.rng.Float64() >= detectionProb {
				continue
			}
			p := s.rng.Float64() * cumProb
			cum := 0.0
			var selected patchCandidate_etox
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
	//s.foragingStats.TotalSearches = len(s.foragershuffle)
	s.patches = s.patches[:0]
	s.foragershuffle = s.foragershuffle[:0]
}

func (s *Foraging_etox) collecting(w *ecs.World) {
	sz := float64(s.foragerParams.SquadronSize)

	// TODO: water collecting here, postponed because water foraging seems basically irrelevant overall
	// TODO: test this foragershuffle
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Experienced || act.Current == activity.BringPollen || act.Current == activity.BringNectar {
			s.foragershuffle = append(s.foragershuffle, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.foragershuffle), func(i, j int) { s.foragershuffle[i], s.foragershuffle[j] = s.foragershuffle[j], s.foragershuffle[i] })

	for _, e := range s.foragershuffle {
		act, patch, milage, load, PPPload, PPPexpo := s.foragerLoadPPPMapper.Get(e)

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
		}

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

			s.foragingStats.Collectionflightstotal += 1

			conf, etoxprops, trip := s.patchConfigMapperEtox.Get(patch.Nectar)
			load.Energy = conf.NectarConcentration * s.foragerParams.NectarLoad * s.energyParams.Sucrose // --> kJ
			dist := trip.CostNectar / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			// exposition from nectar foraging
			PPPload.PPPLoad = load.Energy * etoxprops.PPPconcentrationNectar
			PPPexpo.OralDose += PPPload.PPPLoad * s.toxic.HSuptake
			PPPload.PPPLoad -= PPPload.PPPLoad * s.toxic.HSuptake

			if patch.VisitedthisDay {
				s.foragingStats.ContactExp_repeat++
			} else {
				s.foragingStats.ContactExp_once++
				patch.VisitedthisDay = true
			}
			if s.etox.AppDay == int(s.time.Tick) || !s.etox.ContactExposureOneDay { // this should a) always enable contactexp on AppDay and b) prevent exposure from diluting if ContactExposureOneDay == True for the GUTS model that does not reset ContactDose
				if PPPexpo.ContactDose > 0 {
					if s.etox.ContactSum {
						PPPexpo.ContactDose += etoxprops.PPPcontactDose
					} else {
						PPPexpo.ContactDose = (PPPexpo.ContactDose + etoxprops.PPPcontactDose) / 2
					}
				} else {
					PPPexpo.ContactDose += etoxprops.PPPcontactDose
				}
			}
		}

		if act.Current == activity.BringPollen {

			s.foragingStats.Collectionflightstotal += 1

			_, etoxprops, trip := s.patchConfigMapperEtox.Get(patch.Pollen)
			dist := trip.CostPollen / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			// exposition from pollen foraging
			PPPload.PPPLoad = s.foragerParams.PollenLoad * etoxprops.PPPconcentrationPollen

			if patch.VisitedthisDay {
				s.foragingStats.ContactExp_repeat++
			} else {
				s.foragingStats.ContactExp_once++
				patch.VisitedthisDay = true
			}
			if s.etox.AppDay == int(s.time.Tick) || !s.etox.ContactExposureOneDay { // this should a) always enable contactexp on AppDay and b) prevent exposure from diluting if ContactExposureOneDay == True for the GUTS model that does not reset ContactDose
				if PPPexpo.ContactDose > 0 {
					if s.etox.ContactSum {
						PPPexpo.ContactDose += etoxprops.PPPcontactDose
					} else {
						PPPexpo.ContactDose = (PPPexpo.ContactDose + etoxprops.PPPcontactDose) / 2
					}
				} else {
					PPPexpo.ContactDose += etoxprops.PPPcontactDose
				}
			}
		}
	}
	//s.foragingStats.Collectionflightstotal += len(s.foragershuffle)
	s.foragershuffle = s.foragershuffle[:0]
}

func (s *Foraging_etox) flightCost(w *ecs.World) (duration float64, foragers int) {
	duration = 0.0
	foragers = 0

	// TODO: flightCost for water foraging here, postponed ...

	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Searching || act.Current == activity.BringPollen || act.Current == activity.BringNectar {
			s.foragershuffle = append(s.foragershuffle, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.foragershuffle), func(i, j int) { s.foragershuffle[i], s.foragershuffle[j] = s.foragershuffle[j], s.foragershuffle[i] })

	for _, e := range s.foragershuffle {
		act, patch, milage, _, eload, ppp := s.foragerLoadPPPMapper.Get(e)

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
		}

		if act.Current == activity.Searching {
			dist := s.forageParams.SearchLength / 1000.0
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			en := s.forageParams.SearchLength * s.foragerParams.FlightCostPerM
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)
			eload.EnergyUsed += en

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
			eload.EnergyUsed += en
			ppp.OralDose += s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize), false)

			foragers++
		}
	}
	s.foragershuffle = s.foragershuffle[:0]

	return
}

func (s *Foraging_etox) mortality(w *ecs.World) {
	searchDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity

	// TODO: mortality for water foragers, postponed ..

	foragerQuery := s.foragerFilterLoadPPP.Query()
	for foragerQuery.Next() {
		act, patch, _, _, _, PPPexpo := foragerQuery.Get()

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
		}

		// Acute toxicity during flight
		lethaldose := false
		if s.etox.ForagerImmediateMortality { // always false for now; might as well be deactivateds
			if PPPexpo.RdmSurvivalOral < 1-(1/(1+math.Pow(PPPexpo.OralDose/s.toxic.ForagerOralLD50, s.toxic.ForagerOralSlope))) {
				lethaldose = true
			}
			if PPPexpo.RdmSurvivalContact < 1-(1/(1+math.Pow(PPPexpo.ContactDose/s.toxic.ForagerContactLD50, s.toxic.ForagerContactSlope))) {
				lethaldose = true
			}
		}

		if lethaldose {
			s.toRemove = append(s.toRemove, foragerQuery.Entity())
		} else if act.Current == activity.Searching {
			if s.rng.Float64() < 1-math.Pow(1-s.forageParams.MortalityPerSec, searchDuration) {
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
		}
	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.foragingStats.Foragerdied += len(s.toRemove)
	s.toRemove = s.toRemove[:0]
}

func (s *Foraging_etox) dancing(w *ecs.World) {
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()

		if act.Reverted {
			continue // reverted foragers get passed over; they do not act as foraging foragers but as nurses until reverted gets switched off again
		}

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

func (s *Foraging_etox) unloading(w *ecs.World) {

	// TODO: water unloading, postponed ...

	query := s.loadFilter.Query()
	for query.Next() {
		act, load, PPPload := query.Get()
		if act.Current == activity.BringNectar {

			s.stores.Honey += load.Energy * float64(s.foragerParams.SquadronSize)

			s.stores_etox.ETOX_HES_C_D0 = ((s.stores_etox.ETOX_HES_C_D0 * s.stores_etox.ETOX_HES_E_D0) + (PPPload.PPPLoad * (1 - s.toxic.HSuptake) * float64(s.foragerParams.SquadronSize))) / (s.stores_etox.ETOX_HES_E_D0 + (load.Energy * float64(s.foragerParams.SquadronSize))) // may need to readjust
			s.stores_etox.ETOX_HES_E_D0 += load.Energy * float64(s.foragerParams.SquadronSize)
			if s.stores.Honey > s.maxHoneyStore ||
				s.stores_etox.ETOX_HES_E_Capped+s.stores_etox.ETOX_HES_E_D0+s.stores_etox.ETOX_HES_E_D1+s.stores_etox.ETOX_HES_E_D2+s.stores_etox.ETOX_HES_E_D3+s.stores_etox.ETOX_HES_E_D4 > s.maxHoneyStore {
				s.stores.Honey = s.maxHoneyStore
				s.stores_etox.ETOX_HES_E_D0 = s.maxHoneyStore - (s.stores_etox.ETOX_HES_E_Capped + s.stores_etox.ETOX_HES_E_D1 + s.stores_etox.ETOX_HES_E_D2 + s.stores_etox.ETOX_HES_E_D3 + s.stores_etox.ETOX_HES_E_D4)
			}

			load.Energy = 0.
			PPPload.PPPLoad = 0.
			act.Current = activity.Experienced
		} else if act.Current == activity.BringPollen {
			s.stores_etox.PPPInHivePollenConc = ((s.stores_etox.PPPInHivePollenConc * s.stores.Pollen) + (PPPload.PPPLoad * float64(s.foragerParams.SquadronSize))) / (s.stores.Pollen + s.foragerParams.PollenLoad*float64(s.foragerParams.SquadronSize)) // may need to readjust

			s.stores.Pollen += s.foragerParams.PollenLoad * float64(s.foragerParams.SquadronSize)
			PPPload.PPPLoad = 0.
			act.Current = activity.Experienced

			s.foragingStats.Pollensuccess += 1
		}
	}
}

func (s *Foraging_etox) countForagers(w *ecs.World) {
	sz := s.foragerParams.SquadronSize
	round := globals_etox.ForagingRound_etox{}

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
			} else {
				round.Nectar += sz
			}
		default:
			panic(fmt.Sprintf("forager activity %d invalid at the end of a foraging round", act.Current))
		}
	}

	s.foragingStats.Rounds = append(s.foragingStats.Rounds, round)
}

type patchCandidate_etox struct {
	Patch       ecs.Entity
	Probability float64
	HasNectar   bool
	HasPollen   bool

	HasWater bool
}

// copy from etox_storages_consumption, because it should be more efficient to just use it here as a Foraging_etox function too than to get the necessary globals within the func with every call, right?
func (s *Foraging_etox) FeedOnHoneyStores(w *ecs.World, cons float64, number float64, honeydilution bool) (OralDose float64) {
	OralDose = 0.
	if cons < s.stores_etox.ETOX_HES_E_D0 {
		OralDose += cons * s.stores_etox.ETOX_HES_C_D0 / number
		s.stores_etox.ETOX_HES_E_D0 -= cons
	} else {
		OralDose += s.stores_etox.ETOX_HES_E_D0 * s.stores_etox.ETOX_HES_C_D0 / number
		cons -= s.stores_etox.ETOX_HES_E_D0
		s.stores_etox.ETOX_HES_E_D0 = 0

		if cons < s.stores_etox.ETOX_HES_E_D1 {
			OralDose += cons * s.stores_etox.ETOX_HES_C_D1 / number
			s.stores_etox.ETOX_HES_E_D1 -= cons
		} else {
			OralDose += s.stores_etox.ETOX_HES_E_D1 * s.stores_etox.ETOX_HES_C_D1 / number
			cons -= s.stores_etox.ETOX_HES_E_D1
			s.stores_etox.ETOX_HES_E_D1 = 0

			if cons < s.stores_etox.ETOX_HES_E_D2 {
				OralDose += cons * s.stores_etox.ETOX_HES_C_D2 / number
				s.stores_etox.ETOX_HES_E_D2 -= cons
			} else {
				OralDose += s.stores_etox.ETOX_HES_E_D1 * s.stores_etox.ETOX_HES_C_D2 / number
				cons -= s.stores_etox.ETOX_HES_E_D2
				s.stores_etox.ETOX_HES_E_D2 = 0

				if cons < s.stores_etox.ETOX_HES_E_D3 {
					OralDose += cons * s.stores_etox.ETOX_HES_C_D3 / number
					s.stores_etox.ETOX_HES_E_D3 -= cons
				} else {
					OralDose += s.stores_etox.ETOX_HES_E_D3 * s.stores_etox.ETOX_HES_C_D3 / number
					cons -= s.stores_etox.ETOX_HES_E_D3
					s.stores_etox.ETOX_HES_E_D3 = 0

					if cons < s.stores_etox.ETOX_HES_E_D4 {
						OralDose += cons * s.stores_etox.ETOX_HES_C_D4 / number
						s.stores_etox.ETOX_HES_E_D4 -= cons
					} else {
						OralDose += s.stores_etox.ETOX_HES_E_D4 * s.stores_etox.ETOX_HES_C_D4 / number
						cons -= s.stores_etox.ETOX_HES_E_D4
						s.stores_etox.ETOX_HES_E_D4 = 0

						if cons < s.stores_etox.ETOX_HES_E_Capped {
							OralDose += cons * s.stores_etox.ETOX_HES_C_Capped / number
							s.stores_etox.ETOX_HES_E_Capped -= cons
							if honeydilution {
								s.stores_etox.ETOX_Waterneedfordilution += cons / s.energyParams.Honey / s.etox.ETOXDensityOfHoney * 0.6
							}
						} else {
							OralDose += s.stores_etox.ETOX_HES_E_Capped * s.stores_etox.ETOX_HES_C_Capped / number
							cons -= s.stores_etox.ETOX_HES_E_Capped
							if honeydilution {
								s.stores_etox.ETOX_Waterneedfordilution += s.stores_etox.ETOX_HES_E_Capped / s.energyParams.Honey / s.etox.ETOXDensityOfHoney * 0.6
							}
							s.stores_etox.ETOX_HES_E_Capped = 0
						}
					}
				}
			}
		}
	}
	return
}
