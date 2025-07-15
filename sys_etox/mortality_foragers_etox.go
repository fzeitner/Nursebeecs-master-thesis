package sys_etox

import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// MortalityForagers applies worker mortality, including
//   - background mortality from [params.WorkerMortality.InHive]
//   - removal of squadrons reaching [params.WorkerDevelopment.MaxLifespan]
//   - removal of squadrons exceeding [params.WorkerMortality.MaxMilage]
//   - mortality from PPP exposure if applicable
type MortalityForagers_etox struct {
	rng                  *resource.Rand
	toRemove             []ecs.Entity
	foragerFilter        *ecs.Filter1[comp_etox.PPPExpo]
	foragersFilterSimple *ecs.Filter0

	etoxstats *globals_etox.PopulationStats_etox
	etox      *params_etox.ETOXparams
	toxic     *params_etox.Toxicityparams

	time *resource.Tick
}

func (s *MortalityForagers_etox) Initialize(w *ecs.World) {
	s.rng = ecs.GetResource[resource.Rand](w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.foragersFilterSimple = ecs.NewFilter0(w).With(ecs.C[comp.Age]())

	s.etoxstats = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)

	s.time = ecs.GetResource[resource.Tick](w)

}

func (s *MortalityForagers_etox) Update(w *ecs.World) {
	if s.time.Tick > 0 {

		query := s.foragerFilter.Query()
		s.etoxstats.MeanDoseForager = 0.
		s.etoxstats.CumDoseForagers = 0.

		for query.Next() {
			p := query.Get()

			// mortality from PPP exposition, simple dose response relationship depending on their susceptibility to the contaminant
			lethaldose := false
			if s.etox.Application {
				s.etoxstats.CumDoseForagers += p.OralDose
				if p.OralDose > 1e-20 && p.OralDose < s.toxic.ForagerOralLD50*1e5 {
					if p.RdmSurvivalOral < 1-(1/(1+math.Pow(p.OralDose/s.toxic.ForagerOralLD50, s.toxic.ForagerOralSlope))) {
						lethaldose = true
					}
				}
				if p.ContactDose > 0 {
					if p.RdmSurvivalContact < 1-(1/(1+math.Pow(p.ContactDose/s.toxic.ForagerContactLD50, s.toxic.ForagerContactSlope))) {
						lethaldose = true
					}
				}
			}
			p.OralDose = 0.    // exposure doses get reset to 0 every tick BEFORE the added dose from honey and pollen consumption
			p.ContactDose = 0. // gets taken into account, therefore exposure from foraging of the current day and food from the previous day is relevant

			if lethaldose {
				s.toRemove = append(s.toRemove, query.Entity())
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
			s.etoxstats.MeanDoseForager = s.etoxstats.CumDoseForagers / float64(c)
		} else {
			s.etoxstats.MeanDoseForager = 0.
		}
	}
}

func (s *MortalityForagers_etox) Finalize(w *ecs.World) {}
