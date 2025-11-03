package sys_etox

import (
	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// this subsystem calculates the amount of nurses and decides if the nursing age needs to be adjusted or if some foragers need to revert to nurses

type NursingNeeds struct {
	oldNurseParams *params.Nursing
	NurseParams    *params_etox.Nursing
	newCons        *params_etox.ConsumptionRework
	affParams      *params.AgeFirstForaging

	newCohorts *globals.NewCohorts
	aff        *globals.AgeFirstForaging
	stores     *globals.Stores
	eggs       *globals.Eggs
	larvae     *globals.Larvae
	pupae      *globals.Pupae
	inHive     *globals.InHive
	nglobals   *globals_etox.Nursing_globals
	nstats     *globals_etox.Nursing_stats
	pop        *globals.PopulationStats
	cons       *globals.ConsumptionStats
	time       *resource.Tick
	rng        *resource.Rand

	foragerfilter *ecs.Filter1[comp_etox.Activity_etox]
	foragermapper *ecs.Map1[comp.Age]
	actMapper     *ecs.Map1[comp_etox.Activity_etox]

	toRevert []ecs.Entity
	ageslice []int // experimental; used to sort forager squadrons that are to be reverted to IHbees because of a mass death event leaving the hive without nurses
	sorted   []int // experimental; used to sort forager squadrons that are to be reverted to IHbees because of a mass death event leaving the hive without nurses
}

func (s *NursingNeeds) Initialize(w *ecs.World) {
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.NurseParams = ecs.GetResource[params_etox.Nursing](w)
	s.newCons = ecs.GetResource[params_etox.ConsumptionRework](w)
	s.affParams = ecs.GetResource[params.AgeFirstForaging](w)

	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.eggs = ecs.GetResource[globals.Eggs](w)
	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.pupae = ecs.GetResource[globals.Pupae](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	s.nstats = ecs.GetResource[globals_etox.Nursing_stats](w)

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.rng = ecs.GetResource[resource.Rand](w)

	s.foragerfilter = s.foragerfilter.New(w)
	s.foragermapper = s.foragermapper.New(w)
	s.actMapper = s.actMapper.New(w)
}

func (s *NursingNeeds) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		// implement rules for nurse recruitment

		// regular nurse recruitment by moving age up or down by 1 depending on some metrics; down below there is the special case if there are far too little nurses
		if (!s.nglobals.SuffNurses && s.nglobals.NurseAgeMax < s.aff.Aff) || (s.inHive.Workers[4] == 0 && s.inHive.Workers[min(s.nglobals.NurseAgeMax+1, 50)] > 0) {
			// && float64(NonNurseIHbees/s.pop.TotalAdults) >= 0.1  && float64(s.nstats.TotalNurses/s.pop.TotalAdults) < 0.5 {
			s.nglobals.NurseAgeMax = util.Clamp(s.nglobals.NurseAgeMax+1, 5, 50)
		} else if (s.stores.ProteinFactorNurses < 1. && len(s.nglobals.WinterBees) == 0) || s.nglobals.Reductionpossible {
			s.nglobals.NurseAgeMax = util.Clamp(s.nglobals.NurseAgeMax-1, 5, 50)
		}
		// maybe it is a better idea to calulate every step if nursing demands could still be fulfilled if nurse max age was lowered by the next nonzero-cohort and then lower NurseAgeMax to that age

		// un-revert redundant reverted foragers
		if s.nglobals.SquadstoReduce > 0 {
			s.unRevertForagers(w, s.nglobals.SquadstoReduce)
		}

		TotalNursesLastDay := s.nstats.TotalNurses // in case of etox mortality events in between last day´s consumption proc and now we recount available nurses here
		// now evaluate current state of nursing metrics
		s.calcNursingMetrics(w)

		// special rules if all nurses die due to etox related mass deaths --> foragers get called back to make the hive continue somehow
		// may also make sense to have this activate if there are simply far too little nurses and a large fraction of foragers present
		if (s.nstats.NurseFraction <= 0.05 && s.pop.TotalAdults > 0) || s.nstats.TotalNurses == 0 {
			recruitedNurses := 0
			for i := s.nglobals.NurseAgeMax; i < s.aff.Aff; i++ {
				if s.inHive.Workers[i] != 0 {
					recruitedNurses += s.inHive.Workers[i]
					s.nglobals.NurseAgeMax = util.Clamp(i, 5, 50) // emergency increase to NurseAgeMax

					if float64(recruitedNurses+s.nstats.TotalNurses)/float64(s.pop.TotalAdults) >= 0.05 && float64(recruitedNurses) >= 0.5*float64(TotalNursesLastDay) && s.nstats.TotalNurses != 0 { // reach 5% nurses and at least half of last day; this is experimental
						break
					}
				}
			}
			if (float64(recruitedNurses) < 0.5*float64(TotalNursesLastDay) || float64(recruitedNurses+s.nstats.TotalNurses)/float64(s.pop.TotalAdults) <= 0.05 || s.nstats.TotalNurses == 0) && s.newCohorts.Foragers > 0 {
				s.nglobals.NurseAgeMax = util.Clamp(s.nglobals.NurseAgeMax+1, 5, 50)

				s.inHive.Workers[s.aff.Aff] = s.newCohorts.Foragers * 100
				s.aff.Aff++ // experimental
				s.newCohorts.Foragers = 0
				recruitedNurses += s.inHive.Workers[s.nglobals.NurseAgeMax]
			}
			if (float64(recruitedNurses) < 0.5*float64(TotalNursesLastDay) || float64(recruitedNurses+s.nstats.TotalNurses)/float64(s.pop.TotalAdults) <= 0.05 || s.nstats.TotalNurses == 0) && s.pop.WorkersForagers > 0 {
				s.revertForagers(w, recruitedNurses, TotalNursesLastDay)
			}
		}
		s.calcNursingMetrics(w)
	}
}

func (s *NursingNeeds) Finalize(w *ecs.World) {}

func (s *NursingNeeds) calcNursingMetrics(w *ecs.World) {
	s.nstats.IHbeeNurses = 0
	s.nstats.WinterBees = 0
	s.nstats.RevertedForagers = 0
	for i := 4; i <= s.nglobals.NurseAgeMax; i++ {
		s.nstats.IHbeeNurses += s.inHive.Workers[i]
	}
	// quick fix for the beginning and end of the year without IHbees via making starting foragers winterbees; also includes reverted foragers that might occur in times of great nursing needs or after mass death events
	s.nglobals.WinterBees = s.nglobals.WinterBees[:0]
	s.nglobals.Reverted = s.nglobals.Reverted[:0]
	query := s.foragerfilter.Query() // this is just a bandaid fix to make the model logic work in winterbee time; if there is enough time winterbees could be given actual "jobs"; right now all of them can feed AND forage at the same time
	for query.Next() {
		act := query.Get()
		if act.Winterbee {
			s.nstats.WinterBees += 100
			s.nglobals.WinterBees = append(s.nglobals.WinterBees, query.Entity())
		} else if act.Reverted {
			s.nstats.RevertedForagers += 100
			s.nglobals.Reverted = append(s.nglobals.Reverted, query.Entity())
		}
	}
	s.nstats.TotalNurses = s.nstats.IHbeeNurses + s.nstats.RevertedForagers + s.nstats.WinterBees // maybe ignore winterbees here as they are a bit of a special case?
	s.nstats.NurseFraction = (float64(s.nstats.TotalNurses) / float64(s.pop.TotalAdults)) * 100   // expressed in %

	s.nstats.NonNurseIHbees = 0
	for i := 0; i < 4; i++ {
		s.nstats.NonNurseIHbees += s.inHive.Workers[i]
	}
	for i := s.nglobals.NurseAgeMax + 1; i < len(s.inHive.Workers); i++ {
		s.nstats.NonNurseIHbees += s.inHive.Workers[i]
	}

}

func (s *NursingNeeds) revertForagers(w *ecs.World, recruited int, recruitmenttarget int) {
	// this function reverts foragers to IHbees if they are aged below IHbee maximum age or turns them into winterbees
	// it will only revert as many squadrons as are necessary to meet the recruitment target
	// this should be adjusted once winter bees are modelled explicitly
	query := s.foragerfilter.Query()
	for query.Next() {
		act := query.Get()
		if !act.Winterbee && !act.Reverted {
			s.toRevert = append(s.toRevert, query.Entity())
		}
	}
	doy := int((s.time.Tick - 1) % 365)
	if len(s.toRevert)*100+recruited > recruitmenttarget { // sort squadrons by their age and have them revert until the target is met; only necessary if not all foragers should revert

		for _, e := range s.toRevert {
			age := s.foragermapper.Get(e)
			adultcohortage := doy - age.DayOfBirth
			s.ageslice = append(s.ageslice, adultcohortage)
		}
		minAge := 1000
		for _, val := range s.ageslice {
			if val < minAge {
				minAge = val
			}
		}
		for i := 0; i >= 0; i++ { // infinite loop, breaks once sorted is full
			for i, val := range s.ageslice {
				if val == minAge {
					s.sorted = append(s.sorted, i)
				}
			}
			minAge++
			if len(s.sorted) == len(s.ageslice) {
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
	} else {
		for _, e := range s.toRevert {
			act := s.actMapper.Get(e)
			act.Reverted = true
			recruited += 100

			if recruited >= recruitmenttarget { // should never be reached tbh
				break
			}
		}
	}
	s.ageslice = s.ageslice[:0]
	s.sorted = s.sorted[:0]
	s.toRevert = s.toRevert[:0]
}

func (s *NursingNeeds) unRevertForagers(w *ecs.World, toRevert int) {
	// this function un-reverts reverted foragers back to normal foragers and prefers the oldest ones first
	// it will only revert as many squadrons as are necessary to meet the target

	if len(s.nglobals.Reverted) == toRevert { // just un-revert everyone; no need to sort
		for _, e := range s.nglobals.Reverted {
			act := s.actMapper.Get(e)
			act.Reverted = false
		}

	} else { // sorting will be necessary
		alreadyReverted := 0
		doy := int((s.time.Tick - 1) % 365)

		for _, e := range s.nglobals.Reverted {
			age := s.foragermapper.Get(e)
			adultcohortage := doy - age.DayOfBirth
			s.ageslice = append(s.ageslice, adultcohortage)
		}
		maxAge := 0
		for _, val := range s.ageslice { // this time sort for oldest first
			if val > maxAge {
				maxAge = val
			}
		}
		for i := 0; i >= 0; i++ { // infinite loop, breaks once sorted is full
			for i, val := range s.ageslice {
				if val == maxAge {
					s.sorted = append(s.sorted, i)
				}
			}
			maxAge--
			if len(s.sorted) == len(s.ageslice) {
				break
			}
		}
		for _, i := range s.sorted {
			act := s.actMapper.Get(s.nglobals.Reverted[i])
			if act.Winterbee {
				panic("Reverted bees should not be able to be winterbees. There must be a bug in the code somewhere!")
			}
			act.Reverted = false
			alreadyReverted++

			if alreadyReverted == toRevert {
				break
			}
		}
		s.ageslice = s.ageslice[:0]
		s.sorted = s.sorted[:0]
		s.nglobals.Reverted = s.nglobals.Reverted[:0]
	}
}
