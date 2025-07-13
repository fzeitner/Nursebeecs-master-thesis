package sys

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// MortalityForagers applies worker mortality, including
//   - background mortality from [params.WorkerMortality.InHive]
//   - removal of squadrons reaching [params.WorkerDevelopment.MaxLifespan]
//   - removal of squadrons exceeding [params.WorkerMortality.MaxMilage]
//   - mortality from PPP exposure if applicable
type MortalityForagers struct {
	rng           *resource.Rand
	time          *resource.Tick
	workerMort    *params.WorkerMortality
	workerDev     *params.WorkerDevelopment
	toRemove      []ecs.Entity
	foragerFilter *ecs.Filter3[comp.Age, comp.Milage, comp.PPPExpo]

	foragersFilterSimple *ecs.Filter0

	pop  *globals.PopulationStats
	etox *params.ETOXparams
}

func (s *MortalityForagers) Initialize(w *ecs.World) {
	s.rng = ecs.GetResource[resource.Rand](w)
	s.time = ecs.GetResource[resource.Tick](w)
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.foragerFilter = s.foragerFilter.New(w)

	s.foragersFilterSimple = ecs.NewFilter0(w).With(ecs.C[comp.Age]())

	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.etox = ecs.GetResource[params.ETOXparams](w)
}

func (s *MortalityForagers) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		r := rand.New(s.rng)
		query := s.foragerFilter.Query()
		s.pop.MeanDoseForager = 0.
		s.pop.CumDoseForagers = 0.

		for query.Next() {
			a, m, p := query.Get()

			// mortality from PPP exposition, simple dose response relationship depending on their susceptibility to the contaminant
			lethaldose := false
			if s.etox.Application {
				s.pop.CumDoseForagers += p.OralDose
				if p.OralDose > 1e-20 && p.OralDose < s.etox.ForagerOralLD50*1e5 {
					if p.RdmSurvivalOral < 1-(1/(1+math.Pow(p.OralDose/s.etox.ForagerOralLD50, s.etox.ForagerOralSlope))) {
						lethaldose = true
					}
				}
				if p.ContactDose > 0 {
					if p.RdmSurvivalContact < 1-(1/(1+math.Pow(p.ContactDose/s.etox.ForagerContactLD50, s.etox.ForagerContactSlope))) {
						lethaldose = true
					}
				}
			}
			p.OralDose = 0.    // exposure doses get reset to 0 every tick BEFORE the added dose from honey and pollen consumption
			p.ContactDose = 0. // gets taken into account, therefore exposure from foraging of the current day and food from the previous day is relevant

			if int(s.time.Tick-1)-a.DayOfBirth >= s.workerDev.MaxLifespan ||
				lethaldose ||
				m.Total >= s.workerMort.MaxMilage ||
				r.Float64() < s.workerMort.InHive {
				s.toRemove = append(s.toRemove, query.Entity())
			}
		}
	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.toRemove = s.toRemove[:0]

	querysimple := s.foragersFilterSimple.Query()
	c := querysimple.Count()
	querysimple.Close()

	if c > 0 {
		s.pop.MeanDoseForager = s.pop.CumDoseForagers / float64(c)
	} else {
		s.pop.MeanDoseForager = 0.
	}
}

func (s *MortalityForagers) Finalize(w *ecs.World) {}
