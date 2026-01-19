package sys

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// this subsystem calculates the amount of nurses and decides if the nursing age needs to be adjusted or if some foragers need to revert to nurses

type NursingNeeds struct {
	oldNurseParams *params.Nursing
	NurseParams    *params.NursingRework
	newCons        *params.ConsumptionRework
	affParams      *params.AgeFirstForaging

	newCohorts *globals.NewCohorts
	aff        *globals.AgeFirstForaging
	stores     *globals.Stores
	eggs       *globals.Eggs
	larvae     *globals.Larvae
	pupae      *globals.Pupae
	inHive     *globals.InHive
	nGlobals   *globals.NursingGlobals
	nStats     *globals.NursingStats
	pop        *globals.PopulationStats
	cons       *globals.ConsumptionStats
	time       *resource.Tick
	rng        *resource.Rand

	foragerFilter *ecs.Filter1[comp.ActivityEtox]
	foragerMapper *ecs.Map1[comp.Age]
	actMapper     *ecs.Map1[comp.ActivityEtox]

	toRevert []ecs.Entity // saves foragers squadrons that could potentially be reverted if the need arises
	ageSlice []int        // used to sort forager squadrons that are to be reverted to IHbees because of a mass death event leaving the hive without nurses
	sorted   []int        // used to sort forager squadrons that are to be reverted to IHbees because of a mass death event leaving the hive without nurses
}

func (s *NursingNeeds) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params.NursingRework](w)
	s.newCons = ecs.GetResource[params.ConsumptionRework](w)
	s.affParams = ecs.GetResource[params.AgeFirstForaging](w)

	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.nGlobals = ecs.GetResource[globals.NursingGlobals](w)
	s.nStats = ecs.GetResource[globals.NursingStats](w)

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.rng = ecs.GetResource[resource.Rand](w)

	s.foragerFilter = s.foragerFilter.New(w)
	s.foragerMapper = s.foragerMapper.New(w)
	s.actMapper = s.actMapper.New(w)
}

func (s *NursingNeeds) Update(w *ecs.World) {

	s.nGlobals.KillDrones = false
	TotalNursesLastDay := s.nStats.TotalNurses                      // in case of etox mortality events in between last day´s consumption proc and now we recount available nurses here
	MinimumNurses := max(int(0.5*float64(TotalNursesLastDay)), 200) // minimum of 200 estimated from Kama&Shpigler (2025) as the minimum taskforce able to rear a healthy queen larva

	s.calcNursingMetrics(w)
	if s.nStats.TotalNurses <= MinimumNurses {
		s.nGlobals.KillDrones = true
	}

	// un-revert redundant reverted foragers
	if s.nGlobals.SquadstoReduce > 0 {
		s.unRevertForagers(w, s.nGlobals.SquadstoReduce)
	}
	doy := s.time.Tick % 365

	WinterBeeRatio := 0.
	if s.nStats.TotalNurses > 0 {
		WinterBeeRatio = float64(s.nStats.WinterBees) / float64(s.nStats.TotalNurses)
	}

	if WinterBeeRatio <= 0.9 || s.nStats.TotalNurses <= MinimumNurses {
		// implement rules for nurse recruitment
		// regular nurse recruitment by moving age up or down by 1 depending on some metrics; down below there is the special case if there are far too little nurses

		NewNurseAgeMax := s.nGlobals.NurseAgeMax
		if !s.nGlobals.SuffNurses || (s.inHive.Workers[4] == 0 && s.inHive.Workers[min(s.nGlobals.NurseAgeMax+1, 50)] > 0) {
			NewNurseAgeMax++
		}
		if s.nGlobals.Reductionpossible { //|| s.stores.Pollen <= 0.5*s.stores.IdealPollen {
			NewNurseAgeMax--
		}
		s.nGlobals.NurseAgeMax = util.Clamp(NewNurseAgeMax, 5, s.aff.Aff-1)

		if doy == 0 { // reset NurseAgeMax to default value at the start of a new year
			s.nGlobals.NurseAgeMax = s.NurseParams.NurseAgeCeiling
		}

		/*
			// function to adjust NurseAgeMax to next nonzero cohort, this is still to be tested though
			if (!s.nGlobals.SuffNurses && s.nGlobals.NurseAgeMax < s.aff.Aff) || (s.inHive.Workers[4] == 0 && s.inHive.Workers[min(s.nGlobals.NurseAgeMax+1, 50)] > 0) {
				PufferCohortAvailable := false
				nextCohort := 50
				for i := s.nGlobals.NurseAgeMax + 1; i < s.aff.Aff; i++ {
					if s.inHive.Workers[i] != 0 {
						nextCohort = min(i, nextCohort)
						if nextCohort < i {
							PufferCohortAvailable = true
						}
					}
				}
				if PufferCohortAvailable && nextCohort != 50 {
					s.nGlobals.NurseAgeMax = util.Clamp(nextCohort, 5, s.aff.Aff)
				}
			} else if s.stores.Pollen <= 0 || s.nGlobals.Reductionpossible {
				nextCohort := 0
				for i := s.nGlobals.NurseAgeMax - 1; i > 4; i-- {
					if s.inHive.Workers[i] != 0 {
						nextCohort = max(i, nextCohort)
				}
				if nextCohort != 0 {
					s.nGlobals.NurseAgeMax = util.Clamp(nextCohort, 5, s.aff.Aff)
				}
			}
		*/

		// now evaluate current state of nursing metrics
		s.calcNursingMetrics(w)

		// special rules if all nurses die due to etox related mass deaths --> foragers get called back to make the hive continue somehow
		// may also make sense to have this activate if there are simply far too little nurses and a large fraction of foragers present
		// technically they should have a reduced nursing efficiency, but this has not been implemented as of yet
		totalAdults := s.pop.WorkersForagers + s.pop.WorkersInHive
		if (s.nStats.NurseFraction <= 0.1 || s.nStats.TotalNurses < MinimumNurses) && totalAdults > 0 {
			recruitedNurses := 0
			for i := s.nGlobals.NurseAgeMax; i < s.aff.Aff; i++ {
				if s.inHive.Workers[i] != 0 {
					recruitedNurses += s.inHive.Workers[i]
					s.nGlobals.NurseAgeMax = util.Clamp(i, 5, 50) // emergency increase to NurseAgeMax

					if float64(recruitedNurses+s.nStats.TotalNurses)/float64(totalAdults) >= 0.1 && recruitedNurses >= MinimumNurses { // reach 10% nurses and at least half of last dayl
						break
					}
				}
			}
			if (recruitedNurses < MinimumNurses || float64(recruitedNurses+s.nStats.TotalNurses)/float64(totalAdults) <= 0.1) && s.newCohorts.Foragers > 0 {
				s.nGlobals.NurseAgeMax = util.Clamp(s.aff.Aff, 5, 50)

				s.inHive.Workers[s.aff.Aff] = s.newCohorts.Foragers * 100
				s.aff.Aff++ // increase Aff by 1
				s.newCohorts.Foragers = 0
				recruitedNurses += s.inHive.Workers[s.nGlobals.NurseAgeMax]
			}
			if (recruitedNurses < MinimumNurses || float64(recruitedNurses+s.nStats.TotalNurses)/float64(totalAdults) <= 0.1) && s.pop.WorkersForagers > 0 {
				s.revertForagers(w, recruitedNurses, MinimumNurses)
			}
			s.calcNursingMetrics(w)
		}
	}
}

func (s *NursingNeeds) Finalize(w *ecs.World) {}

func (s *NursingNeeds) calcNursingMetrics(w *ecs.World) {
	s.nStats.IHbeeNurses = 0
	s.nStats.WinterBees = 0
	s.nStats.RevertedForagers = 0
	s.nStats.NonNurseIHbees = 0

	for i := 0; i < len(s.inHive.Workers); i++ {
		if i < 4 {
			s.nStats.NonNurseIHbees += s.inHive.Workers[i]
		} else if i >= 4 && i <= s.nGlobals.NurseAgeMax {
			s.nStats.IHbeeNurses += s.inHive.Workers[i]
		} else {
			s.nStats.NonNurseIHbees += s.inHive.Workers[i]
		}
	}
	// quick fix for the beginning and end of the year without IHbees via making starting foragers winterbees; also includes reverted foragers that might occur in times of great nursing needs or after mass death events
	s.nGlobals.WinterBees = s.nGlobals.WinterBees[:0]
	s.nGlobals.Reverted = s.nGlobals.Reverted[:0]
	query := s.foragerFilter.Query() // this is just a bandaid fix to make the model logic work in winterbee time; if there is enough time winterbees could be given actual "jobs"; right now all of them can feed AND forage at the same time
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
	s.nStats.TotalNurses = s.nStats.IHbeeNurses + s.nStats.RevertedForagers + s.nStats.WinterBees // maybe ignore winterbees here as they are a bit of a special case?
	s.nStats.NurseFraction = (float64(s.nStats.TotalNurses) / float64(s.pop.WorkersForagers+s.pop.WorkersInHive))
}

func (s *NursingNeeds) revertForagers(w *ecs.World, recruited int, recruitmenttarget int) {
	// this function reverts foragers to IHbees if they are aged below IHbee maximum age or turns them into winterbees
	// it will only revert as many squadrons as are necessary to meet the recruitment target
	// this should be adjusted once winter bees are modelled explicitly
	s.ageSlice = s.ageSlice[:0]
	s.sorted = s.sorted[:0]
	s.toRevert = s.toRevert[:0]

	query := s.foragerFilter.Query()
	for query.Next() {
		act := query.Get()
		if !act.Winterbee && !act.Reverted {
			s.toRevert = append(s.toRevert, query.Entity())
		}
	}
	doy := int((s.time.Tick) % 365)
	if len(s.toRevert)*100+recruited > recruitmenttarget { // sort squadrons by their age and have them revert until the target is met; only necessary if not all foragers should revert

		for _, e := range s.toRevert {
			age := s.foragerMapper.Get(e)
			adultcohortage := doy - age.DayOfBirth
			s.ageSlice = append(s.ageSlice, adultcohortage)
		}
		minAge := 1000
		for _, val := range s.ageSlice {
			if val < minAge {
				minAge = val
			}
		}
		for i := 0; i >= 0; i++ { // infinite loop, breaks once sorted is full
			for i, val := range s.ageSlice {
				if val == minAge {
					s.sorted = append(s.sorted, i)
				}
			}
			minAge++
			if len(s.sorted) == len(s.ageSlice) {
				break
			}
		}
		for _, i := range s.sorted {
			act := s.actMapper.Get(s.toRevert[i])
			act.Reverted = true
			recruited += 100

			if recruited >= recruitmenttarget {
				break
			}
		}
	} else { // just revert everyone
		for _, e := range s.toRevert {
			act := s.actMapper.Get(e)
			act.Reverted = true
			recruited += 100

			if recruited >= recruitmenttarget { // should never be reached tbh
				break
			}
		}
	}
	s.ageSlice = s.ageSlice[:0]
	s.sorted = s.sorted[:0]
	s.toRevert = s.toRevert[:0]
}

func (s *NursingNeeds) unRevertForagers(w *ecs.World, toRevert int) {
	// this function un-reverts reverted foragers back to normal foragers and prefers the oldest ones first
	// it will only revert as many squadrons as are necessary to meet the target

	if len(s.nGlobals.Reverted) == toRevert { // just un-revert everyone; no need to sort
		for _, e := range s.nGlobals.Reverted {
			act := s.actMapper.Get(e)
			act.Reverted = false
		}

	} else { // sorting will be necessary
		alreadyReverted := 0
		doy := int((s.time.Tick) % 365)

		for _, e := range s.nGlobals.Reverted {
			age := s.foragerMapper.Get(e)
			adultcohortage := doy - age.DayOfBirth
			s.ageSlice = append(s.ageSlice, adultcohortage)
		}
		maxAge := 0
		for _, val := range s.ageSlice { // this time sort for oldest first
			if val > maxAge {
				maxAge = val
			}
		}
		for i := 0; i >= 0; i++ { // infinite loop, breaks once sorted is full
			for i, val := range s.ageSlice {
				if val == maxAge {
					s.sorted = append(s.sorted, i)
				}
			}
			maxAge--
			if len(s.sorted) == len(s.ageSlice) {
				break
			}
		}
		for _, i := range s.sorted {
			act := s.actMapper.Get(s.nGlobals.Reverted[i])
			if act.Winterbee {
				panic("Reverted bees should not be able to be winterbees. There must be a bug in the code somewhere!")
			}
			act.Reverted = false
			alreadyReverted++

			if alreadyReverted == toRevert {
				break
			}
		}
		s.ageSlice = s.ageSlice[:0]
		s.sorted = s.sorted[:0]
		s.nGlobals.Reverted = s.nGlobals.Reverted[:0]
	}
}
