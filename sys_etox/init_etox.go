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
	nglobals    *globals_etox.Nursing_globals
	etox        *params_etox.ETOXparams
	nursecons   *params_etox.ConsumptionRework
	nurseparams *params_etox.Nursing

	foragerfilter    *ecs.Filter1[comp.Activity]
	patchfilter      *ecs.Filter1[comp.Coords]
	etoxAdder        *ecs.Map2[comp_etox.KnownPatch_etox, comp_etox.Activity_etox]
	source           rand.Source
	foragerPPPmapper *ecs.Map2[comp_etox.PPPExpo, comp_etox.EtoxLoad]
	patchPPPmapper   *ecs.Map2[comp_etox.PatchProperties_etox, comp_etox.Resource_etox]

	waterperiodData globals_etox.WaterForagingPeriodData
}

func (s *Init_etox) Initialize(w *ecs.World) {
	// initialize the globals for larvae/IHbee exposure
	aff := ecs.GetResource[params.AgeFirstForaging](w)
	workerDev := ecs.GetResource[params.WorkerDevelopment](w)
	droneDev := ecs.GetResource[params.DroneDevelopment](w)
	s.etox = ecs.GetResource[params_etox.ETOXparams](w)

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

	PPPfate := globals_etox.PPPfate{}
	ecs.AddResource(w, &PPPfate)

	// nursebeecs consumption changes to larvae get initializd here
	s.nursecons = ecs.GetResource[params_etox.ConsumptionRework](w)
	// assume total honey need of 65.4 mg based on Rortais et al. 2005 and original BEEHAVE, but spread throughout the different stages
	// of development. Brouwers & Beetsma 1987 show the sugar and protein content of jelly provided to different larvae over time; multiple authors
	// provide information on exponential larval growth. This combines to the following intake estimates:

	// worker honey intake; estimated growth rate based on Bishop (1961) and data of Brouwers&Beetsma (1987)
	s.nursecons.HoneyWorkerLarva[0] = 0.0036 * s.nursecons.HWLtotal // <1% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.HoneyWorkerLarva[1] = 0.0091 * s.nursecons.HWLtotal // +-1% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.HoneyWorkerLarva[2] = 0.0288 * s.nursecons.HWLtotal // +-3% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.HoneyWorkerLarva[3] = 0.1953 * s.nursecons.HWLtotal // +-20% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.HoneyWorkerLarva[4] = 0.5498 * s.nursecons.HWLtotal // +-55% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.HoneyWorkerLarva[5] = 0.2134 * s.nursecons.HWLtotal // +-21% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	// worker pollen intake; estimated growth rate based on Bishop (1961) and data of Brouwers&Beetsma (1987)
	s.nursecons.PollenWorkerLarva[0] = 0.0148 * s.nursecons.PWLtotal // +-1.5% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.PollenWorkerLarva[1] = 0.0454 * s.nursecons.PWLtotal // +-4.5% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.PollenWorkerLarva[2] = 0.1320 * s.nursecons.PWLtotal // +-13% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.PollenWorkerLarva[3] = 0.3905 * s.nursecons.PWLtotal // +-39% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.PollenWorkerLarva[4] = 0.3299 * s.nursecons.PWLtotal // +-33% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nursecons.PollenWorkerLarva[5] = 0.0874 * s.nursecons.PWLtotal // +-9% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary

	// drone honey intake; estimated from growth relative to worker larvae and data of Matsuka et al. (1973)
	s.nursecons.HoneyDroneLarva[0] = 0.0024 * s.nursecons.HDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[1] = 0.0049 * s.nursecons.HDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[2] = 0.0229 * s.nursecons.HDLtotal // +-2% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[3] = 0.0697 * s.nursecons.HDLtotal // +-7% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[4] = 0.3303 * s.nursecons.HDLtotal // +-33% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[5] = 0.4943 * s.nursecons.HDLtotal // +-50% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.HoneyDroneLarva[6] = 0.0755 * s.nursecons.HDLtotal // +-8% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	// drone pollen intake; estimated from growth relative to worker larvae and data of Matsuka et al. (1973)
	s.nursecons.PollenDroneLarva[0] = 0.0060 * s.nursecons.PDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[1] = 0.0253 * s.nursecons.PDLtotal // +-2.5% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[2] = 0.0707 * s.nursecons.PDLtotal // +-7% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[3] = 0.1927 * s.nursecons.PDLtotal // +-20% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[4] = 0.3923 * s.nursecons.PDLtotal // +-40% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[5] = 0.2637 * s.nursecons.PDLtotal // +-25% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nursecons.PollenDroneLarva[6] = 0.0493 * s.nursecons.PDLtotal // +-5% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	// define capabilities to digest pollen and nurse brood depending on the age (mostly based on Crailsheim et al. 1992 and Hrassnigg&Crailsheim 1998)
	if s.nursecons.DynamicProteinNursing {
		for i := 4; i < 51; i++ {
			s.nursecons.Nursingcapabiliies[i] = 0.9 // assume a maximum efficiency of 90% for returned nurse bees independent of age
		}
		s.nursecons.Nursingcapabiliies[4] = 0.5   // HPG function is still ramping up
		s.nursecons.Nursingcapabiliies[5] = 0.7   // HPG function is still ramping up
		s.nursecons.Nursingcapabiliies[6] = 0.9   // HPG function is still ramping up
		s.nursecons.Nursingcapabiliies[7] = 1     // peak of nursing ablity
		s.nursecons.Nursingcapabiliies[8] = 1     // peak of nursing ablity
		s.nursecons.Nursingcapabiliies[9] = 1     // peak of nursing ablity
		s.nursecons.Nursingcapabiliies[10] = 0.95 // slowly decent to "normal" maximum
		s.nursecons.Nursingcapabiliies[11] = 0.95 // slowly decent to "normal" maximum
	} else {
		for i := 0; i < 51; i++ {
			s.nursecons.Nursingcapabiliies[i] = 1. // assume the same capability for nurse bees independent of age
		}
	}
	s.nursecons.Nursingcapabiliies[0] = 0. // this needs to be 0 for a calc in nurse_consumption and freshly emerged bees never nurse immediately so this does make sense biologically as well

	s.nurseparams = ecs.GetResource[params_etox.Nursing](w)

	s.nglobals = ecs.GetResource[globals_etox.Nursing_globals](w)
	s.nglobals.SuffNurses = true                           // assume that first stimestep there simply are enough nurses, maybe change at some point
	s.nglobals.NurseAgeMax = s.nurseparams.NurseAgeCeiling // initialize with baseline nurse max age from params
	// rest probably won´t need to be initialized here because it should get set with a value before any other subsystem calls for a value

	stats_etox := globals_etox.PopulationStats_etox{}
	ecs.AddResource(w, &stats_etox)

	forstats_etox := globals_etox.ForagingStats_etox{}
	ecs.AddResource(w, &forstats_etox)

	// add the PPPExpo component to all foragers
	s.source = rand.New(ecs.GetResource[resource.Rand](w))
	s.foragerPPPmapper = s.foragerPPPmapper.New(w)
	s.etoxAdder = s.etoxAdder.New(w)
	s.foragerfilter = s.foragerfilter.New(w)

	query := s.foragerfilter.Query()
	toAdd := []ecs.Entity{}

	for query.Next() {
		toAdd = append(toAdd, query.Entity())
	}
	// also adds GUTS related components here if GUTS is enabled
	rng := rand.New(s.source)
	for _, entity := range toAdd {
		s.foragerPPPmapper.Add(entity, &comp_etox.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: rng.Float64(), RdmSurvivalOral: rng.Float64()}, &comp_etox.EtoxLoad{PPPLoad: 0., EnergyUsed: 0.})
		// add winterbee characteristic here depending on parameter value
		if s.nurseparams.WinterBees {
			s.etoxAdder.Add(entity, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting, Winterbee: true})
		} else {
			s.etoxAdder.Add(entity, &comp_etox.KnownPatch_etox{}, &comp_etox.Activity_etox{Current: activity.Resting, Winterbee: false})
		}
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
