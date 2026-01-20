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

// The following function runs the default beecs_ecotox model version 100 times with 4 slightly different parameter settings.
// Specifically, it creates one baseline model run and then enables the fixes (ReworkedThermoETOX, Nursebeefix, HSUfix) introduced in my master thesis
// to enable proper functioning of the _ecotox submodels. This function creates CSV-files in the "out" folder that can be aggregated and analyzed using the
// mass_balance.ipynb notebook in this folder.
func main() {
	// create the application
	app := app.New()

	// define the parameters; most parameters are named analogously to Netlogo,
	// but some parameters are new and enable new functions. Explanations can either be taken
	// from my master thesis or can be found in the params folder. They are separated into different
	// files depending on the model version they are used for primarily.
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

	// timestamp to estimate runtimes of the mdoel
	start := time.Now()

	// run 100 baseline model runs
	name := "baseline"
	for i := 0; i < 100; i++ {
		run(app, i, name, &p, &pe)
	}
	dur := time.Since(start)
	fmt.Println(dur)

	// run 100 model runs with the thermofix enabled
	name = "thermofix"
	pe.PPPApplication.ReworkedThermoETOX = true
	for i := 0; i < 100; i++ {
		run(app, i, name, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// run 100 model runs with the previous fix and the nursebeefix enabled
	name = "nursefix"
	pe.PPPApplication.Nursebeefix = true
	for i := 0; i < 100; i++ {
		run(app, i, name, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// run 100 model runs with the previous fixes and the HSUfix enabled
	name = "HSUfix"
	pe.PPPApplication.HSUfix = true
	for i := 0; i < 100; i++ {
		run(app, i, name, &p, &pe)
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

// the run function runs the model according to the previous definitions
func run(app *app.App, idx int, filename string, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app) // this is the model version; for the various model versions check the model folder

	app.AddSystem(&reporter.CSV{
		Observer: &obs.PPPFateObs{}, // this is the observer that observes and notes PPP fate; for the various observers check the observer folder
		File:     fmt.Sprintf("Out/"+filename+"-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
