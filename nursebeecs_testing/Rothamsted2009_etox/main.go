package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/Nursebeecs-master-thesis/model_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/obs"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/params_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	p.ForagingPeriod = params.ForagingPeriod{
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}
	p.Termination.MaxTicks = 365

	pe := params_etox.Default_etox()
	pe.ETOXparams = params_etox.ETOXparams{
		Application:               false,
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,
		RealisticStoch:            false,
		ReworkedThermoETOX:        true,
		Nursebeefix:               true,
		HSUfix:                    true,

		PPPname:                "dimethoate", // Identifier for the PPP used.
		PPPconcentrationNectar: 990,
		PPPconcentrationPollen: 26631,
		PPPcontactExposure:     0.3, // contact exposure at patch

		AppDay:         217,   // Day of the year in which application starts [d].
		ExposurePeriod: 8,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	pe.Toxicityparams = params_etox.Toxicityparams{
		ForagerOralLD50:  1000., // fenoxycarb
		ForagerOralSlope: 100.,  // fenoxycarb
		HSuptake:         0.1,   //

		ForagerContactLD50:  193.92, // fenoxycarb
		ForagerContactSlope: 1.08,   // fenoxycarb

		LarvaeOralLD50:  0.0014, // fenoxycarb
		LarvaeOralSlope: 1.6,    // fenoxycarb

		NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ], 1 = no filtering effect, 0 = everything gets filtered
		NursebeesPollen: 0.25, // Factor describing the filter effect of nurse bees for pollen [ ], 1 = no filtering effect, 0 = everything gets filtered
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

	run_nbeecs := false // switch to run normal and/or nurse beecs
	if run_nbeecs {
		pe.Nursing.NewConsumption = true
		pe.Nursing.NewBroodCare = true
		pe.Nursing.Nursebeecsv1 = false
		pe.Nursing.ForesightedCannibalism = false

		for i := 0; i < 100; i++ {
			run_nursebeecs(app, i, &p, &pe)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)

	run_nbeecs2 := true // switch to run normal and/or nurse beecs
	if run_nbeecs2 {
		pe.Nursing.NewConsumption = true
		pe.Nursing.NewBroodCare = true
		pe.Nursing.Nursebeecsv0 = false
		pe.Nursing.Nursebeecsv1 = true
		pe.Nursing.ForesightedCannibalism = false

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
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default_nbeecs(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/oldbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default_nbeecs(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/newbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
