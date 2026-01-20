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

	// define default parameters; uses Berlin2000 as the default weather scenario. To change see beecs_examples.
	p := params.Default()
	p.Termination.MaxTicks = 365

	// define DefaultEtox parameters
	pe := params.DefaultEtox()

	// optionally, manually change the parameters for PPPApplication and PPPToxicity to simulate different pesticides
	// all parameters aside from the 4 booleans marked as new parameters below are directly taken from BEEHAVE_ecotox and work analogously
	pe.PPPApplication = params.PPPApplication{
		Application:               true,  // Determines whether an application shall take place at all; acts as an activator for all PPP related submodels and effects
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false, // Determines whether contact exposure shall be summed up when bees visit multiple patches; same as NetLogo version
		ContactExposureOneDay:     true,  // Determines whether contact exposure shall be limited to the day of applicaiton; same as NetLogo

		RealisticStoch:     false, // new parameter; changes the mortality calc when cohorts are very small
		ReworkedThermoETOX: false, // new parameter; fix for the thermoregulation energy in beecs_ecotox
		Nursebeefix:        true,  // new parameter; fix to ensure no PPP get lost when NursebeeFactors are used
		HSUfix:             true,  // new parameter; fix to ensure HSUptake does not let PPP vanish when foragers unload nectar

		PPPname:                "dimethoate", // Identifier for the PPP used.
		PPPconcentrationNectar: 1320,         // concentration in the nectar of the simulated patches [mug/kg]
		PPPconcentrationPollen: 36200,        // concentration in the pollen of the simulated patches [mug/kg]
		PPPcontactExposure:     0.4,          // contact exposure at patch [kg/ha]

		AppDay:         217,   // Day of the year on which application starts [d].
		ExposurePeriod: 9,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	// everything here works exactly like in BEEHAVE_ecotox
	pe.PPPToxicity = params.PPPToxicity{
		ForagerOralLD50:  0.127, // dimethoate
		ForagerOralSlope: 4.37,  // dimethoate
		HSuptake:         0.1,   // default value of BEEHAVE_ecotox

		ForagerContactLD50:  0.169, // dimethoate
		ForagerContactSlope: 16.6,  // dimethoate

		LarvaeOralLD50:  0.24,  // dimethoate
		LarvaeOralSlope: 1.186, // dimethoate

		NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]; default value of BEEHAVE_ecotox
		NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]; default value of BEEHAVE_ecotox
	}

	// you can also change any parameter directly without re-defining the complete parameter set above, like this:
	pe.PPPApplication.HSUfix = false
	pe.PPPApplication.Nursebeefix = false
	pe.PPPApplication.ReworkedThermoETOX = false
	// this deactivated all fixes I added to the ecotox model, meaning that the functinality is now identical to BEEHAVE_ecotox at release

	start := time.Now()

	// run beecs base model; note that ecotox-parameters are irrelevant here, because they are not needed/considered in the base model
	for i := 0; i < 100; i++ {
		runBaseModel(app, i, &p)
	}
	dur := time.Since(start)
	fmt.Println(dur)

	// run beecs_ecotox model; note that p (params) and pe (etox-params) are given to the model; runEtox is defined below
	for i := 0; i < 100; i++ {
		runEtox(app, i, &p, &pe)
	}

	dur = time.Since(start)
	fmt.Println(dur)
}

// define the functions that run the 2 model versions here. This can be copy-pasted easily to run various model versions
// note that it might be necessary to include more parameters for some model versions depending on your needs.
// also, you may want to change observer or create a new one depending on your needs. The observer defines which and how metrics are reported in the CSV-files.
// The observer also defines how the CSV-files will be names, see "File:" declaration below.

func runBaseModel(app *app.App, idx int, params params.Params) {
	// define the model version here by choosing from the 4 default models
	app = model.Default(params, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.Debug{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func runEtox(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	// define beecs_ecotox here by choosing model.DefaultEtox
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.Debug{},
		File:     fmt.Sprintf("out/beecsEtox-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

// the created CSV-files can be aggregated and visualized depending on your preferences and needs; rudimentary visualizations using Python's Matplotlib can be found
// in the py-folders of etox_validation_testing or nursebeecs_testing.
