package sys_etox

// TODO: PPP input from read_in_file
// TExposure_at_patch_ETOX <- netlogo proc
import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// ReplenishPatches resets and replenishes flower patches to their current maximum nectar and pollen.
type PPPApplication struct {
	time   *resource.Tick
	filter *ecs.Filter2[comp_etox.PatchProperties_etox, comp_etox.Resource_etox]

	etox          *params_etox.ETOXparams
	energycontent *params.EnergyContent

	constantFilter *ecs.Filter3[comp_etox.PatchProperties_etox, comp.ConstantPatch, comp_etox.Resource_etox]
	seasonalFilter *ecs.Filter3[comp_etox.PatchProperties_etox, comp.SeasonalPatch, comp_etox.Resource_etox]
	scriptedFilter *ecs.Filter3[comp_etox.PatchProperties_etox, comp.ScriptedPatch, comp_etox.Resource_etox]
}

func (s *PPPApplication) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	s.filter = s.filter.New(w)

	s.etox = ecs.GetResource[params_etox.ETOXparams](w)
	s.energycontent = ecs.GetResource[params.EnergyContent](w)

	s.constantFilter = s.constantFilter.New(w)
	s.seasonalFilter = s.seasonalFilter.New(w)
	s.scriptedFilter = s.scriptedFilter.New(w)

}

func (s *PPPApplication) Update(w *ecs.World) {
	if s.etox.Application {
		dayOfYear := int(s.time.Tick % 365)
		etox_year := int(s.time.Tick / 365)

		constQuery := s.constantFilter.Query()
		for constQuery.Next() {
			props, con, res := constQuery.Get()

			props.PPPconcentrationNectar = res.PPPconcentrationNectar
			props.PPPconcentrationPollen = res.PPPconcentrationPollen
			props.PPPcontactDose = res.PPPcontactDose

			if (etox_year >= s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase) ||
				props.PPPconcentrationNectar+props.PPPconcentrationPollen+props.PPPcontactDose > 0 {
				if s.etox.AppDay == dayOfYear && etox_year >= s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase {
					if con.NectarConcentration != 0 {
						props.PPPconcentrationNectar += ((s.etox.PPPconcentrationNectar / (1 - 0.1047*con.NectarConcentration)) / con.NectarConcentration) / (1000 * 1000 * s.energycontent.Sucrose) // looks complicated, simply adjusts the units properly to mug/kJ though depending on chemical properties
					} else {
						props.PPPconcentrationNectar += 0
					}
					props.PPPconcentrationPollen += s.etox.PPPconcentrationPollen / 1000 // mug/kg -> mug/g
					props.PPPcontactDose += s.etox.PPPcontactExposure * s.etox.RUD * 0.1 // [kg/ha] * [(ha*mg)/(kg*kg)] * [g]
				}
				if s.etox.ContactExposureOneDay && dayOfYear != s.etox.AppDay {
					props.PPPcontactDose = 0
				}
				if dayOfYear >= s.etox.AppDay+s.etox.ExposurePeriod || // TODO: ReadInFile adjustments
					etox_year*365+dayOfYear == s.etox.SpinupPhase*365+(s.etox.ExposurePhase-1)*365+s.etox.AppDay+s.etox.ExposurePeriod {
					props.PPPconcentrationNectar = 0
					props.PPPconcentrationPollen = 0
					props.PPPcontactDose = 0
				} else if dayOfYear != s.etox.AppDay {
					props.PPPconcentrationNectar *= math.Exp(-math.Log(2) / s.etox.DT50)
					props.PPPconcentrationPollen *= math.Exp(-math.Log(2) / s.etox.DT50)
					props.PPPcontactDose *= math.Exp(-math.Log(2) / s.etox.DT50)
				}
			}
		}

		seasonalQuery := s.seasonalFilter.Query()
		for seasonalQuery.Next() {
			props, seas, res := seasonalQuery.Get()

			props.PPPconcentrationNectar = res.PPPconcentrationNectar
			props.PPPconcentrationPollen = res.PPPconcentrationPollen
			props.PPPcontactDose = res.PPPcontactDose

			if (etox_year >= s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase) ||
				props.PPPconcentrationNectar+props.PPPconcentrationPollen+props.PPPcontactDose > 0 {
				if s.etox.AppDay == dayOfYear && etox_year > s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase {
					if seas.NectarConcentration != 0 {
						props.PPPconcentrationNectar += ((s.etox.PPPconcentrationNectar / (1 - 0.1047*seas.NectarConcentration)) / seas.NectarConcentration) / (1000 * 1000 * s.energycontent.Sucrose) // looks complicated, simply adjusts the units properly to mug/kJ though depending on chemical properties
					} else {
						props.PPPconcentrationNectar += 0
					}
					props.PPPconcentrationPollen += s.etox.PPPconcentrationPollen / 1000 // mug/kg -> mug/g
					props.PPPcontactDose += s.etox.PPPcontactExposure * s.etox.RUD * 0.1 // [kg/ha] * [(ha*mg)/(kg*kg)] * [g]
				}
				if s.etox.ContactExposureOneDay && dayOfYear != s.etox.AppDay {
					props.PPPcontactDose = 0
				}
				if dayOfYear >= s.etox.AppDay+s.etox.ExposurePeriod || // TODO: ReadInFile adjustments
					etox_year*365+dayOfYear == s.etox.SpinupPhase*365+(s.etox.ExposurePhase-1)*365+s.etox.AppDay+s.etox.ExposurePeriod {
					props.PPPconcentrationNectar = 0
					props.PPPconcentrationPollen = 0
					props.PPPcontactDose = 0
				} else if dayOfYear != s.etox.AppDay {
					props.PPPconcentrationNectar *= math.Exp(-math.Log(2) / s.etox.DT50)
					props.PPPconcentrationPollen *= math.Exp(-math.Log(2) / s.etox.DT50)
					props.PPPcontactDose *= math.Exp(-math.Log(2) / s.etox.DT50)
				}
			}
		}
	}
	// TODO: STILL NEED TO IMPLEMENT SCRIPTED PATCHES TO ENABLE ReadInFiles

	query := s.filter.Query()
	for query.Next() {
		conf, res := query.Get()

		res.PPPconcentrationNectar = conf.PPPconcentrationNectar
		res.PPPconcentrationPollen = conf.PPPconcentrationPollen
		res.PPPcontactDose = conf.PPPcontactDose
	}
}

func (s *PPPApplication) Finalize(w *ecs.World) {}
