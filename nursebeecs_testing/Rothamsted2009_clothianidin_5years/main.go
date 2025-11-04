package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/beecs_masterthesis/model_etox"
	"github.com/fzeitner/beecs_masterthesis/obs"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	pe := params_etox.Default_etox()
	p.Termination.MaxTicks = 365 * 5

	p.Termination.WinterCritExtinction = true // let the hive die if below critical pop threshold
	p.Termination.CritColonySizeWinter = 4000

	pe.ETOXparams = params_etox.ETOXparams{
		Application:               true,
		ReworkedThermoETOX:        true,
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,

		PPPname:                "clothianidin", // Identifier for the PPP used.
		PPPconcentrationNectar: 990 / 100,
		PPPconcentrationPollen: 0, //27150 / 100,
		PPPcontactExposure:     0, //0.3 / 100, // kg/ha; contact exposure at patch; 0.035 kg/ha of clothianidin was advised somewhere against potato beeles; this is just a test

		AppDay:         166,   // Day of the year in which application starts [d]. --> just assumed 15. of june for fighting potato beetle
		ExposurePeriod: 8,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    1,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	pe.Toxicityparams = params_etox.Toxicityparams{ // clothianidin oral slope was estimated based on LD10 and LD50 from Sgolastra et al. 2017
		ForagerOralLD50:  0.00168, // clothianidin
		ForagerOralSlope: 3.28,    // clothianidin
		HSuptake:         0.1,     //

		// ContactLD50 is backed by PPDB; Slope is not estimated yet and larval data is missing completely atm
		ForagerContactLD50:  0.044, // clothianidin
		ForagerContactSlope: 3.28,  // clothianidin

		LarvaeOralLD50:  1000, // clothianidin
		LarvaeOralSlope: 1000, // clothianidin

		NursebeesNectar: 0.05, // Factor describing the filter effect of nurse bees for nectar [ ], 1 = no filtering effect, 0 = everything gets filtered
		NursebeesPollen: 0.05, // Factor describing the filter effect of nurse bees for pollen [ ], 1 = no filtering effect, 0 = everything gets filtered
	}

	p.ForagingPeriod = params.ForagingPeriod{
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}

	start := time.Now()

	run_beecs := true // switch to run normal and/or nurse beecs
	if run_beecs {
		for i := 0; i < 100; i++ {
			run(app, i, &p, &pe)
		}
	}
	dur := time.Since(start)
	fmt.Println(dur)

	run_nbeecs := true // switch to run normal and/or nurse beecs
	if run_nbeecs {
		pe.ConsumptionRework.Nursebeecs = true
		pe.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pe.Nursing.NewBroodCare = false

		for i := 0; i < 100; i++ {
			run_nursebeecs(app, i, &p, &pe)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)

	run_nbeecs2 := true // switch to run normal and/or nurse beecs
	if run_nbeecs2 {
		pe.ConsumptionRework.Nursebeecs = true
		pe.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pe.Nursing.NewBroodCare = true
		pe.Nursing.ScrambleComp = false

		for i := 0; i < 100; i++ {
			run_nursebeecs2(app, i, &p, &pe)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.NurseBeecsDefault(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/oldbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.NurseBeecsDefault(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursingEtox{},
		File:     fmt.Sprintf("out/newbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
