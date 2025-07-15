package sys_etox

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"os"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/fzeitner/beecs_masterthesis/data"
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// InitCohorts initializes and adds the resources
// [globals.Eggs], [globals.Larvae], [globals.Pupae] and [globals.InHive].
type Init_etox struct {
	larvae_etox globals_etox.Larvae_etox
	inHive_etox globals_etox.InHive_etox

	foragerfilter    *ecs.Filter1[comp.Activity]
	patchfilter      *ecs.Filter1[comp.Coords]
	etoxExchanger    *ecs.Exchange2[comp_etox.KnownPatch_etox, comp_etox.Activity_etox]
	source           rand.Source
	foragerPPPmapper *ecs.Map2[comp_etox.PPPExpo, comp_etox.PPPLoad]
	patchPPPmapper   *ecs.Map2[comp_etox.PatchProperties_etox, comp_etox.Resource_etox]

	waterperiodData globals_etox.WaterForagingPeriodData
}

func (s *Init_etox) Initialize(w *ecs.World) {
	// initialize the globals for larvae/IHbee exposure
	aff := ecs.GetResource[params.AgeFirstForaging](w)
	workerDev := ecs.GetResource[params.WorkerDevelopment](w)
	droneDev := ecs.GetResource[params.DroneDevelopment](w)

	s.larvae_etox = globals_etox.Larvae_etox{
		WorkerCohortDose: make([]float64, workerDev.LarvaeTime),
		DroneCohortDose:  make([]float64, droneDev.LarvaeTime),
	}
	ecs.AddResource(w, &s.larvae_etox)

	s.inHive_etox = globals_etox.InHive_etox{
		WorkerCohortDose: make([]float64, aff.Max+1),
		DroneCohortDose:  make([]float64, droneDev.MaxLifespan),
	}
	ecs.AddResource(w, &s.inHive_etox)

	// initialize ETOX storage globals
	init := ecs.GetResource[params.InitialStores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)
	storages_etox := globals_etox.Storages_etox{
		ETOX_HES_E_Capped: init.Honey * 1000.0 * energyParams.Honey,
		ETOX_HES_C_Capped: 0,
		ETOX_HES_E_D0:     0,
		ETOX_HES_C_D0:     0,
		ETOX_HES_E_D1:     0,
		ETOX_HES_C_D1:     0,
		ETOX_HES_E_D2:     0,
		ETOX_HES_C_D2:     0,
		ETOX_HES_E_D3:     0,
		ETOX_HES_C_D3:     0,
		ETOX_HES_E_D4:     0,
		ETOX_HES_C_D4:     0,
	}
	ecs.AddResource(w, &storages_etox)

	stats_etox := globals_etox.PopulationStats_etox{}
	ecs.AddResource(w, &stats_etox)

	forstats_etox := globals_etox.ForagingStats_etox{}
	ecs.AddResource(w, &forstats_etox)

	// add the PPPExpo component for all foragers
	s.foragerfilter = s.foragerfilter.New(w)
	s.source = rand.New(ecs.GetResource[resource.Rand](w))
	s.foragerPPPmapper = s.foragerPPPmapper.New(w)
	s.etoxExchanger = s.etoxExchanger.New(w)

	query := s.foragerfilter.Without(ecs.C[comp_etox.PPPExpo]()).Query()
	toAdd := []ecs.Entity{}
	for query.Next() {
		toAdd = append(toAdd, query.Entity())
	}

	exchange := s.etoxExchanger.Removes(ecs.C[comp.KnownPatch](), ecs.C[comp.Activity]())
	rng := rand.New(s.source)
	for _, entity := range toAdd {
		s.foragerPPPmapper.Add(entity, &comp_etox.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: rng.Float64(), RdmSurvivalOral: rng.Float64()}, &comp_etox.PPPLoad{PPPLoad: 0.})
		exchange.Exchange(entity, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting})
	}
	toAdd = toAdd[:0]

	// add the PPP components to all patches
	s.patchPPPmapper = s.patchPPPmapper.New(w)
	s.patchfilter = s.patchfilter.New(w)

	pquery := s.patchfilter.Without(ecs.C[comp_etox.Resource_etox]()).Query()
	for pquery.Next() {
		toAdd = append(toAdd, pquery.Entity())
	}
	for _, entity := range toAdd {
		s.patchPPPmapper.Add(entity, &comp_etox.PatchProperties_etox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.}, &comp_etox.Resource_etox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.})
	}

	// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day
	// water foraging from here on, does work with embedded files, but water foraging as a whole is not tested
	// and should not be turned on as of yet, might be deleted/never finished; time will tell
	waterParams := ecs.GetResource[params_etox.WaterParams](w)
	s.waterperiodData = globals_etox.WaterForagingPeriodData{}
	ecs.AddResource(w, &s.waterperiodData)
	if waterParams.WaterForaging {
		waterPeriodParams := ecs.GetResource[params_etox.WaterForagingPeriod](w)
		var WaterfileSys fs.FS = data.WaterNeedsDaily
		if !waterPeriodParams.Builtin {
			wd := ecs.GetResource[params.WorkingDirectory](w).Path
			WaterfileSys = os.DirFS(wd)
		}

		// fill from data provided directly
		for _, arr := range waterPeriodParams.Years {
			if len(arr)%365 != 0 {
				log.Fatal(fmt.Errorf("foraging period entries requires multiple of 365 values, parameters have %d", len(arr)))
			}
			years := len(arr) / 365
			for year := 0; year < years; year++ {
				s.waterperiodData.Years = append(s.waterperiodData.Years, arr[year*365:(year+1)*365])
			}
		}

		// fill from files
		for _, f := range waterPeriodParams.Files {
			arr, err := util.FloatArrayFromFile(WaterfileSys, f)
			if err != nil {
				log.Fatal(fmt.Errorf("error reading foraging period file '%s': %s", f, err.Error()))
			}
			if len(arr)%365 != 0 {
				log.Fatal(fmt.Errorf("foraging period file requires multiple of 365 values, '%s' has %d", f, len(arr)))
			}
			years := len(arr) / 365
			for year := 0; year < years; year++ {
				s.waterperiodData.Years = append(s.waterperiodData.Years, arr[year*365:(year+1)*365])
			}
		}
	}
}

func (s *Init_etox) Update(w *ecs.World) {}

func (s *Init_etox) Finalize(w *ecs.World) {}
