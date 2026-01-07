package sys_etox

import (
	"math"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/comp_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// MortalityForagers_etox applies worker mortality, including
//   - mortality from PPP exposure if applicable
type MortalityForagers_etox struct {
	rng                  *resource.Rand
	toRemove             []ecs.Entity
	foragerFilter        *ecs.Filter1[comp_etox.PPPExpo]
	foragersFilterSimple *ecs.Filter0

	etoxstats *globals_etox.PopulationStats_etox
	etox      *params_etox.ETOXparams
	toxic     *params_etox.Toxicityparams
}

func (s *MortalityForagers_etox) Initialize(w *ecs.World) {
	s.rng = ecs.GetResource[resource.Rand](w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.foragersFilterSimple = ecs.NewFilter0(w).With(ecs.C[comp.Age]())

	s.etoxstats = ecs.GetResource[globals_etox.PopulationStats_etox](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.toxic = ecs.GetResource[params_etox.Toxicityparams](w)
}

func (s *MortalityForagers_etox) Update(w *ecs.World) {
	query := s.foragerFilter.Query()
	s.etoxstats.MeanDoseForager = 0.
	s.etoxstats.CumDoseForagers = 0.
	c := query.Count()

	for query.Next() {
		p := query.Get()
		// mortality from PPP exposition, either dose-response relationship depending on their susceptibility to the contaminant or BeeGUTS can be called here
		lethaldose := false
		if s.etox.Application {
			s.etoxstats.CumDoseForagers += p.OralDose * 100
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
			p.OralDose = 0.    // exposure doses get reset to 0 every tick BEFORE the added dose from honey and pollen consumption gets taken into account,
			p.ContactDose = 0. // therefore exposure from foraging of the current day and exposure from food of the previous day is relevant for lethal effects only
		}
		if lethaldose {
			s.toRemove = append(s.toRemove, query.Entity())
		}
	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.toRemove = s.toRemove[:0]

	//querysimple := s.foragersFilterSimple.Query()
	//c := querysimple.Count()
	//querysimple.Close()

	if c > 0 {
		s.etoxstats.MeanDoseForager = s.etoxstats.CumDoseForagers / float64(c*100)
	} else {
		s.etoxstats.MeanDoseForager = 0.
	}
}

func (s *MortalityForagers_etox) Finalize(w *ecs.World) {}
