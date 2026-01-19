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
	p.Termination.MaxTicks = 365

	pe := params.DefaultEtox()
	pe.PPPApplication = params.PPPApplication{
		Application:               true,
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,
		RealisticStoch:            false,
		ReworkedThermoETOX:        false,
		Nursebeefix:               false,
		HSUfix:                    false,

		PPPname:                "fenoxycarb", // Identifier for the PPP used.
		PPPconcentrationNectar: 990,
		PPPconcentrationPollen: 27150,
		PPPcontactExposure:     0.3, // kg/ha; contact exposure at patch

		AppDay:         189,   // Day of the year in which application starts [d].
		ExposurePeriod: 8,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	pe.PPPToxicity = params.PPPToxicity{
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

	//pe.ConsumptionRework.Nursebeecs = false // this obviously does now work because Nursebeecs has to be run with Default_nbeecs model and default model has to be run with Nursebeecs false
	// --> maybe rethink structure of models and params to make this more intuitive ????

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p, &pe)
	}
	dur := time.Since(start)
	fmt.Println(dur)

	pe.PPPApplication.ReworkedThermoETOX = true
	for i := 0; i < 100; i++ {
		run2(app, i, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	pe.PPPApplication.Nursebeefix = true
	for i := 0; i < 100; i++ {
		run3(app, i, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	pe.PPPApplication.HSUfix = true
	for i := 0; i < 100; i++ {
		run4(app, i, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.PPPFateObs{},
		File:     fmt.Sprintf("out/baseline-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run2(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.PPPFateObs{},
		File:     fmt.Sprintf("out/thermofix-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run3(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.PPPFateObs{},
		File:     fmt.Sprintf("out/nursefix-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run4(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.PPPFateObs{},
		File:     fmt.Sprintf("out/HSUfix-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
