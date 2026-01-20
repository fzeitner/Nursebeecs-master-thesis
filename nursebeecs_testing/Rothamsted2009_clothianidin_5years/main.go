package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/Nursebeecs-master-thesis/model"
	"github.com/fzeitner/Nursebeecs-master-thesis/obs"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	pe := params.DefaultEtox()
	pn := params.DefaultNursebeecs()
	p.Termination.MaxTicks = 365 * 8

	p.Termination.OnExtinction = false
	p.Termination.WinterCritExtinction = true // let the hive die if below critical pop threshold
	p.Termination.CritColonySizeWinter = 4000

	pe.PPPApplication = params.PPPApplication{
		Application:               true,
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,
		RealisticStoch:            false,
		ReworkedThermoETOX:        true,
		HSUfix:                    true,
		Nursebeefix:               true,

		PPPname:                "clothianidin", // Identifier for the PPP used.
		PPPconcentrationNectar: 3,              //990 / 100,
		PPPconcentrationPollen: 0,              //27150 / 100,
		PPPcontactExposure:     0,              //0.3 / 100, // kg/ha; contact exposure at patch; 0.035 kg/ha of clothianidin was advised somewhere against potato beeles; this is just a test

		AppDay:         182,   // Day of the year in which application starts [d]. --> just assumed 01. of july for now; might change. There is a large window for fighting potato beetle for example (june to august at least)
		ExposurePeriod: 30,    // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    2,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  6,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	pe.PPPToxicity = params.PPPToxicity{ // clothianidin oral slope was estimated based on LD10 and LD50 from Sgolastra et al. 2017
		ForagerOralLD50:  0.00168, // clothianidin
		ForagerOralSlope: 3.28,    // clothianidin
		HSuptake:         0.1,     //

		// ContactLD50 is backed by PPDB; Slope is not estimated yet and larval data is missing completely atm
		ForagerContactLD50:  0.044, // clothianidin
		ForagerContactSlope: 1,     // clothianidin

		// larval toxdata is missing as well atm
		LarvaeOralLD50:  1000, // clothianidin
		LarvaeOralSlope: 1,    // clothianidin

		NursebeesNectar: 0.05, // Factor describing the filter effect of nurse bees for nectar [ ], 1 = no filtering effect, 0 = everything gets filtered
		NursebeesPollen: 0.05, // Factor describing the filter effect of nurse bees for pollen [ ], 1 = no filtering effect, 0 = everything gets filtered

		HGthreshold: []float64{0.000727 / 12.78, 0.000727 / 12.78 * 10, 0.000727 / 12.78 * 100}, // authors used 1 mug/L Clothianidin in 400 ml of Apiinvert; Apiinvert has 1 kg of sugar per Liter and a concentration of 72.7% (https://www.beefeed.com/en/apiinvert/) --> 1 l of Apiinvert should be equivalent to 1/0.727 = 1.376 kg
		// therefore 1 mug/L (w/v) should be equivalent to 0.727 mug/kg (w/w) in apiinvert; 0.727 mug/kg = 0.727 ng/g = 0.000727 mug/g; BEEHAVE needs values in weight per kJ, therefore we need to consider honey energy content of 12.78 kJ/g
		ProteinFactorNurseExposed: []float64{0.65, 0.65, 0.65}, // very much experimental; straight up taken from Schott et al. 2021
		MaxPollenRed:              []float64{0.3, 0.3, 0.3},    // no clue yet if this makes any sense
	}

	p.ForagingPeriod = params.ForagingPeriod{
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}

	start := time.Now()

	run_beecs := true // switch to run normal and/or nurse beecs
	if run_beecs {
		pn.NursingRework.NewBroodCare = true
		pn.NursingRework.Nursebeecsv1 = true

		for i := 0; i < 100; i++ {
			run(app, i, &p, &pe, &pn)
		}
	}
	dur := time.Since(start)
	fmt.Println(dur)

	run_nbeecs := true // switch to run normal and/or nurse beecs
	if run_nbeecs {
		pn.NursingRework.NewBroodCare = true
		pn.NursingRework.Nursebeecsv1 = true

		pn.NursingRework.HGEffects = true
		pn.NursingRework.HGFoodIntake = false

		for i := 0; i < 100; i++ {
			run_nursebeecs(app, i, &p, &pe, &pn)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)

	run_nbeecs2 := true // switch to run normal and/or nurse beecs
	if run_nbeecs2 {
		pn.NursingRework.NewBroodCare = true
		pn.NursingRework.Nursebeecsv1 = true

		pn.NursingRework.HGEffects = true
		pn.NursingRework.HGFoodIntake = true

		for i := 0; i < 100; i++ {
			run_nursebeecs2(app, i, &p, &pe, &pn)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox, paramsNbeecs params.ParamsNursebeecs) {
	app = model.DefaultNbeecsEtox(params, paramsEtox, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/nbeecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox, paramsNbeecs params.ParamsNursebeecs) {
	app = model.DefaultNbeecsEtox(params, paramsEtox, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/nbeecsHG-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox, paramsNbeecs params.ParamsNursebeecs) {
	app = model.DefaultNbeecsEtox(params, paramsEtox, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/nbeecsHGFood-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
