package sys

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"os"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/data"
	"github.com/fzeitner/Nursebeecs-master-thesis/enum/activity"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// InitCohorts initializes and adds the resources
// [globals.Eggs], [globals.Larvae], [globals.Pupae] and [globals.InHive].
type InitEtox struct {
	larvaeEtox globals.LarvaeEtox
	inHiveEtox globals.InHiveEtox
	etox       *params.PPPApplication

	foragerFilter    *ecs.Filter1[comp.Activity]
	patchFilter      *ecs.Filter1[comp.Coords]
	etoxAdder        *ecs.Map2[comp.KnownPatchEtox, comp.ActivityEtox]
	source           rand.Source
	foragerPPPmapper *ecs.Map2[comp.PPPExpo, comp.EtoxLoad]
	patchPPPmapper   *ecs.Map2[comp.PatchPropertiesEtox, comp.ResourceEtox]

	waterPeriodData globals.WaterForagingPeriodData
}

func (s *InitEtox) Initialize(w *ecs.World) {
	// initialize the globals for larvae/IHbee exposure
	aff := ecs.GetResource[params.AgeFirstForaging](w)
	workerDev := ecs.GetResource[params.WorkerDevelopment](w)
	droneDev := ecs.GetResource[params.DroneDevelopment](w)
	s.etox = ecs.GetResource[params.PPPApplication](w)

	s.larvaeEtox = globals.LarvaeEtox{
		WorkerCohortDose: make([]float64, workerDev.LarvaeTime),
		DroneCohortDose:  make([]float64, droneDev.LarvaeTime),
	}
	ecs.AddResource(w, &s.larvaeEtox)

	s.inHiveEtox = globals.InHiveEtox{
		WorkerCohortDose: make([]float64, aff.Max+1),
		DroneCohortDose:  make([]float64, droneDev.MaxLifespan),
	}
	ecs.AddResource(w, &s.inHiveEtox)

	// initialize ETOX storage globals
	init := ecs.GetResource[params.InitialStores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)
	storagesEtox := globals.StoragesEtox{
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
	ecs.AddResource(w, &storagesEtox)

	PPPfate := globals.PPPFate{}
	ecs.AddResource(w, &PPPfate)

	statsEtox := globals.PopulationStatsEtox{}
	ecs.AddResource(w, &statsEtox)

	forstatsEtox := globals.ForagingStatsEtox{}
	ecs.AddResource(w, &forstatsEtox)

	// add the PPPExpo component to all foragers
	s.source = rand.New(ecs.GetResource[resource.Rand](w))
	s.foragerPPPmapper = s.foragerPPPmapper.New(w)
	s.etoxAdder = s.etoxAdder.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	query := s.foragerFilter.Query()
	toAdd := []ecs.Entity{}

	for query.Next() {
		toAdd = append(toAdd, query.Entity())
	}
	// also adds GUTS related components here if GUTS is enabled
	rng := rand.New(s.source)
	for _, entity := range toAdd {
		s.foragerPPPmapper.Add(entity, &comp.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: rng.Float64(), RdmSurvivalOral: rng.Float64()}, &comp.EtoxLoad{PPPLoad: 0., EnergyUsed: 0.})

		s.etoxAdder.Add(entity, &comp.KnownPatchEtox{}, &comp.ActivityEtox{Current: activity.Resting, Winterbee: false})
	}
	toAdd = toAdd[:0]

	// add the PPP components to all patches
	s.patchPPPmapper = s.patchPPPmapper.New(w)
	s.patchFilter = s.patchFilter.New(w)

	pquery := s.patchFilter.Without(ecs.C[comp.ResourceEtox]()).Query()
	for pquery.Next() {
		toAdd = append(toAdd, pquery.Entity())
	}
	for _, entity := range toAdd {
		s.patchPPPmapper.Add(entity, &comp.PatchPropertiesEtox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.}, &comp.ResourceEtox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.})
	}

	// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day
	// water foraging from here on, does work with embedded files, but water foraging as a whole is not tested
	// and should not be turned on as of yet, might be deleted/never finished; time will tell
	waterParams := ecs.GetResource[params.WaterForaging](w)
	s.waterPeriodData = globals.WaterForagingPeriodData{}
	ecs.AddResource(w, &s.waterPeriodData)
	if waterParams.WaterForaging {
		waterPeriodParams := ecs.GetResource[params.WaterForagingPeriod](w)
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
				s.waterPeriodData.Years = append(s.waterPeriodData.Years, arr[year*365:(year+1)*365])
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
				s.waterPeriodData.Years = append(s.waterPeriodData.Years, arr[year*365:(year+1)*365])
			}
		}
	}
}

func (s *InitEtox) Update(w *ecs.World) {}

func (s *InitEtox) Finalize(w *ecs.World) {}
