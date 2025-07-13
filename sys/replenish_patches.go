package sys

// TODO: PPP input from read_in_file
// TExposure_at_patch_ETOX <- netlogo proc
import (
	"math"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// ReplenishPatches resets and replenishes flower patches to their current maximum nectar and pollen.
type ReplenishPatches struct {
	time   *resource.Tick
	filter *ecs.Filter3[comp.PatchProperties, comp.Resource, comp.Visits]

	etox          *params.ETOXparams
	energycontent *params.EnergyContent

	constantFilter *ecs.Filter3[comp.PatchProperties, comp.ConstantPatch, comp.Resource]
	seasonalFilter *ecs.Filter3[comp.PatchProperties, comp.SeasonalPatch, comp.Resource]
	scriptedFilter *ecs.Filter3[comp.PatchProperties, comp.ScriptedPatch, comp.Resource]
}

func (s *ReplenishPatches) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	s.filter = s.filter.New(w)

	s.etox = ecs.GetResource[params.ETOXparams](w)
	s.energycontent = ecs.GetResource[params.EnergyContent](w)

	s.constantFilter = s.constantFilter.New(w)
	s.seasonalFilter = s.seasonalFilter.New(w)
	s.scriptedFilter = s.scriptedFilter.New(w)

}

func (s *ReplenishPatches) Update(w *ecs.World) {
	if s.time.Tick > 0 {
		dayOfYear := int(s.time.Tick%365) - 1
		etox_year := int(s.time.Tick) / 365

		// TODO: water foraging, postponed because of lacking relevance

		constQuery := s.constantFilter.Query()
		for constQuery.Next() {
			props, con, res := constQuery.Get()

			props.MaxNectar = con.Nectar
			props.MaxPollen = con.Pollen

			props.DetectionProbability = con.DetectionProbability
			props.NectarConcentration = con.NectarConcentration

			props.PPPconcentrationNectar = res.PPPconcentrationNectar
			props.PPPconcentrationPollen = res.PPPconcentrationPollen
			props.PPPcontactDose = res.PPPcontactDose

			// first try of implementing a PPP application and managing its decay process
			// TODO: implement read in file with pesticide data and fit globals/parameters to this format; Water foraging
			if s.etox.Application && etox_year >= s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase ||
				s.etox.Application && props.PPPconcentrationNectar+props.PPPconcentrationPollen+props.PPPcontactDose > 0 { // if im not mistaken this cannot be accessed properly or is always wrong
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

			day := (s.time.Tick - 1 + int64(seas.SeasonShift)) % 365
			season := util.Season(day)

			props.MaxNectar = seas.MaxNectar * season
			props.MaxPollen = seas.MaxPollen * season

			props.DetectionProbability = seas.DetectionProbability
			props.NectarConcentration = seas.NectarConcentration

			props.PPPconcentrationNectar = res.PPPconcentrationNectar
			props.PPPconcentrationPollen = res.PPPconcentrationPollen
			props.PPPcontactDose = res.PPPcontactDose

			// first try of implementing a PPP application and managing its decay process
			// TODO: implement read in file with pesticide data and fit globals/parameters to this format; Water foraging
			if s.etox.Application && etox_year >= s.etox.SpinupPhase && etox_year < s.etox.SpinupPhase+s.etox.ExposurePhase ||
				s.etox.Application && props.PPPconcentrationNectar+props.PPPconcentrationPollen+props.PPPcontactDose > 0 { // if im not mistaken this cannot be accessed properly or is always wrong
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

		// TODO: STILL NEED TO FIGURE OUT SCRIPTED PATCHES
		day := s.time.Tick % 365
		scriptedQuery := s.scriptedFilter.Query()
		for scriptedQuery.Next() {
			props, scr, _ := scriptedQuery.Get()
			//_, con := constQuery.Get()

			props.MaxNectar = util.Interpolate(scr.Nectar, float64(day), scr.Interpolation)
			props.MaxPollen = util.Interpolate(scr.Pollen, float64(day), scr.Interpolation)

			props.DetectionProbability = util.Interpolate(scr.DetectionProbability, float64(day), scr.Interpolation)
			props.NectarConcentration = util.Interpolate(scr.NectarConcentration, float64(day), scr.Interpolation)

			/*
				props.PPPconcentrationNectar = res.PPPconcentrationNectar
				props.PPPconcentrationPollen = res.PPPconcentrationPollen
				props.PPPcontactDose = res.PPPcontactDose

					// first try of implementing a PPP application and managing its decay process, here still implemented with con, because I don´t really get the interpolation patch method yet
					if s.etox.Application && int(s.time.Tick) >= s.etox.SpinupPhase*365 && int(s.time.Tick) <= (s.etox.SpinupPhase+s.etox.ExposurePhase)*365 {
						if s.etox.AppDay+s.etox.ExposurePeriod >= dayOfYear && dayOfYear >= s.etox.AppDay {
							props.PPPconcentrationNectar = con.PPPconcentrationNectar
							props.PPPconcentrationPollen = con.PPPconcentrationPollen
						} else {
							props.PPPconcentrationNectar = props.PPPconcentrationNectar * math.Exp(-math.Log(2)/s.etox.DT50)
							props.PPPconcentrationPollen = props.PPPconcentrationPollen * math.Exp(-math.Log(2)/s.etox.DT50)
							props.PPPcontactDose = props.PPPcontactDose * math.Exp(-math.Log(2)/s.etox.DT50)
						}
					} else {
						props.PPPconcentrationNectar = 0
						props.PPPconcentrationPollen = 0
						props.PPPcontactDose = 0

					}*/
		}

		query := s.filter.Query()
		for query.Next() {
			conf, res, visits := query.Get()

			res.MaxNectar = conf.MaxNectar * 1000 * 1000
			res.MaxPollen = conf.MaxPollen * 1000

			res.Nectar = res.MaxNectar
			res.Pollen = res.MaxPollen

			visits.Nectar = 0
			visits.Pollen = 0

			res.PPPconcentrationNectar = conf.PPPconcentrationNectar
			res.PPPconcentrationPollen = conf.PPPconcentrationPollen
			res.PPPcontactDose = conf.PPPcontactDose
		}
	}
}

func (s *ReplenishPatches) Finalize(w *ecs.World) {}
