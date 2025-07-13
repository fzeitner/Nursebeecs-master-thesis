package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/beecs_ecotox/model"
	"github.com/fzeitner/beecs_ecotox/obs"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	p.Termination.MaxTicks = 365

	p.ETOXparams = params.ETOXparams{
		Application:            true,
		PPPname:                "fenoxycarb", // Identifier for the PPP used.
		PPPconcentrationNectar: 990,
		PPPconcentrationPollen: 27150,
		PPPcontactExposure:     0.3,   // actually dont have a value for this
		DT50:                   1000., // Whole plant DT50 from residue studies [d].
		AppDay:                 191,   // Day of the year in which application starts [d].
		ExposurePeriod:         8,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:            0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:          3,     // Number of years in which exposure takes place [y].

		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,

		ForagerOralLD50:  1000., // dimethoate
		ForagerOralSlope: 100.,  // dimethoate
		HSuptake:         0.1,   //

		ForagerContactLD50:  193.92, // dimethoate
		ForagerContactSlope: 1.08,   // dimethoate

		LarvaeOralLD50:  0.0014, // dimethoate
		LarvaeOralSlope: 1.6,    // dimethoate

		NursebeesNectar: 1.,  // Factor describing the filter effect of nurse bees for nectar [ ], 1 = no filtering effect, 0 = everything gets filtered
		NursebeesPollen: 1.,  // Factor describing the filter effect of nurse bees for pollen [ ], 1 = no filtering effect, 0 = everything gets filtered
		DT50honey:       60., // Honey DT50 [d]
		RUD:             21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	p.ForagingPeriod = params.ForagingPeriod{
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params) {
	app = model.Default(params, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugEcotox{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
