package sys

import (
	"math/rand/v2"

	"github.com/fzeitner/Nursebeecs-master-thesis/comp"
	"github.com/fzeitner/Nursebeecs-master-thesis/enum/activity"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// InitCohorts initializes and adds the resources
// [globals.Eggs], [globals.Larvae], [globals.Pupae] and [globals.InHive].
type InitNursebeecs struct {
	nGlobals       *globals.NursingGlobals
	nurseCons      *params.ConsumptionRework
	nurseParams    *params.NursingRework
	oldNurseParams *params.Nursing

	foragersFilter *ecs.Filter0
	actAdder       *ecs.Map1[comp.ActivityEtox]
	source         rand.Source

	//waterperiodData globals.WaterForagingPeriodData
}

func (s *InitNursebeecs) Initialize(w *ecs.World) {
	// initialize the globals for larvae/IHbee exposure

	// nursebeecs consumption changes to larvae get initializd here
	s.nurseCons = ecs.GetResource[params.ConsumptionRework](w)
	// assume total honey need of 65.4 mg based on Rortais et al. 2005 and original BEEHAVE, but spread throughout the different stages
	// of development. Brouwers & Beetsma 1987 show the sugar and protein content of jelly provided to different larvae over time; multiple authors
	// provide information on exponential larval growth. This combines to the following intake estimates:

	// worker honey intake; estimated growth rate based on Bishop (1961) and data of Brouwers&Beetsma (1987)
	s.nurseCons.HoneyWorkerLarva[0] = 0.0036 * s.nurseCons.HWLtotal // <1% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.HoneyWorkerLarva[1] = 0.0091 * s.nurseCons.HWLtotal // +-1% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.HoneyWorkerLarva[2] = 0.0288 * s.nurseCons.HWLtotal // +-3% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.HoneyWorkerLarva[3] = 0.1953 * s.nurseCons.HWLtotal // +-20% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.HoneyWorkerLarva[4] = 0.5498 * s.nurseCons.HWLtotal // +-55% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.HoneyWorkerLarva[5] = 0.2134 * s.nurseCons.HWLtotal // +-21% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	// worker pollen intake; estimated growth rate based on Bishop (1961) and data of Brouwers&Beetsma (1987)
	s.nurseCons.PollenWorkerLarva[0] = 0.0148 * s.nurseCons.PWLtotal // +-1.5% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.PollenWorkerLarva[1] = 0.0454 * s.nurseCons.PWLtotal // +-4.5% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.PollenWorkerLarva[2] = 0.1320 * s.nurseCons.PWLtotal // +-13% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.PollenWorkerLarva[3] = 0.3905 * s.nurseCons.PWLtotal // +-39% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.PollenWorkerLarva[4] = 0.3299 * s.nurseCons.PWLtotal // +-33% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary
	s.nurseCons.PollenWorkerLarva[5] = 0.0874 * s.nurseCons.PWLtotal // +-9% of total intake on this day, estimate based on Brouwers&Beetsma (1987), Bishop (1961) and scientific guess of total amount necessary

	// drone honey intake; estimated from growth relative to worker larvae and data of Matsuka et al. (1973)
	s.nurseCons.HoneyDroneLarva[0] = 0.0024 * s.nurseCons.HDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[1] = 0.0049 * s.nurseCons.HDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[2] = 0.0229 * s.nurseCons.HDLtotal // +-2% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[3] = 0.0697 * s.nurseCons.HDLtotal // +-7% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[4] = 0.3303 * s.nurseCons.HDLtotal // +-33% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[5] = 0.4943 * s.nurseCons.HDLtotal // +-50% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.HoneyDroneLarva[6] = 0.0755 * s.nurseCons.HDLtotal // +-8% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	// drone pollen intake; estimated from growth relative to worker larvae and data of Matsuka et al. (1973)
	s.nurseCons.PollenDroneLarva[0] = 0.0060 * s.nurseCons.PDLtotal // <1% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[1] = 0.0253 * s.nurseCons.PDLtotal // +-2.5% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[2] = 0.0707 * s.nurseCons.PDLtotal // +-7% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[3] = 0.1927 * s.nurseCons.PDLtotal // +-20% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[4] = 0.3923 * s.nurseCons.PDLtotal // +-40% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[5] = 0.2637 * s.nurseCons.PDLtotal // +-25% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	s.nurseCons.PollenDroneLarva[6] = 0.0493 * s.nurseCons.PDLtotal // +-5% of total intake on this day, estimate based on Matsuka et al. (1973) and scientific guess of total amount necessary
	// define capabilities to digest pollen and nurse brood depending on the age (mostly based on Crailsheim et al. 1992 and Hrassnigg&Crailsheim 1998)
	if s.nurseCons.DynamicProteinNursing {
		for i := 4; i < 51; i++ {
			s.nurseCons.Nursingcapabiliies[i] = 0.9 // assume a maximum efficiency of 90% for returned nurse bees independent of age
		}
		s.nurseCons.Nursingcapabiliies[4] = 0.5   // HPG function is still ramping up
		s.nurseCons.Nursingcapabiliies[5] = 0.7   // HPG function is still ramping up
		s.nurseCons.Nursingcapabiliies[6] = 0.9   // HPG function is still ramping up
		s.nurseCons.Nursingcapabiliies[7] = 1     // peak of nursing ablity
		s.nurseCons.Nursingcapabiliies[8] = 1     // peak of nursing ablity
		s.nurseCons.Nursingcapabiliies[9] = 1     // peak of nursing ablity
		s.nurseCons.Nursingcapabiliies[10] = 0.95 // slowly decent to "normal" maximum
		s.nurseCons.Nursingcapabiliies[11] = 0.95 // slowly decent to "normal" maximum
	} else {
		for i := 0; i < 51; i++ {
			s.nurseCons.Nursingcapabiliies[i] = 1. // assume the same capability for nurse bees independent of age
		}
	}
	s.nurseCons.Nursingcapabiliies[0] = 0. // this needs to be 0 for a calc in nurse_consumption and freshly emerged bees never nurse immediately so this does make sense biologically as well

	s.nurseParams = ecs.GetResource[params.NursingRework](w)

	s.nGlobals = ecs.GetResource[globals.NursingGlobals](w)
	s.nGlobals.SuffNurses = true                           // assume that first stimestep there simply are enough nurses, maybe change at some point
	s.nGlobals.NurseAgeMax = s.nurseParams.NurseAgeCeiling // initialize with baseline nurse max age from params
	// rest probably won´t need to be initialized here because it should get set with a value before any other subsystem calls for a value

	statsEtox := globals.PopulationStatsEtox{}
	ecs.AddResource(w, &statsEtox)

	forstatsEtox := globals.ForagingStatsEtox{}
	ecs.AddResource(w, &forstatsEtox)

	// add the PPPExpo component to all foragers
	s.source = rand.New(ecs.GetResource[resource.Rand](w))
	s.actAdder = s.actAdder.New(w)
	s.foragersFilter = ecs.NewFilter0(w).With(ecs.C[comp.Age]())

	query := s.foragersFilter.Query()
	toAdd := []ecs.Entity{}

	for query.Next() {
		toAdd = append(toAdd, query.Entity())
	}

	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	// adds winterbee characteristic here if enabled
	for _, entity := range toAdd {
		if s.oldNurseParams.WinterBees {
			s.actAdder.Add(entity, &comp.ActivityEtox{Current: activity.Resting, Winterbee: true})
		} else {
			s.actAdder.Add(entity, &comp.ActivityEtox{Current: activity.Resting, Winterbee: false})
		}
	}

	/*
		// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day
		// water foraging from here on, does work with embedded files, but water foraging as a whole is not tested
		// and should not be turned on as of yet, might be deleted/never finished; time will tell
		waterParams := ecs.GetResource[params.WaterForaging](w)
		s.waterperiodData = globals.WaterForagingPeriodData{}
		ecs.AddResource(w, &s.waterperiodData)
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
	*/
}

func (s *InitNursebeecs) Update(w *ecs.World) {}

func (s *InitNursebeecs) Finalize(w *ecs.World) {}
