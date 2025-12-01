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
	p.Termination.MaxTicks = 365

	pe := params_etox.Default_etox()
	pe.ETOXparams = params_etox.ETOXparams{
		Application: true,
		//GUTS:                      true,
		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false,
		ContactExposureOneDay:     true,

		PPPname:                "dimethoate", // Identifier for the PPP used.
		PPPconcentrationNectar: 990,
		PPPconcentrationPollen: 26631,
		PPPcontactExposure:     0.3, // contact exposure at patch

		AppDay:         189,   // Day of the year in which application starts [d].
		ExposurePeriod: 8,     // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,     // Number of years in which exposure takes place [y].
		DT50:           1000., // Whole plant DT50 from residue studies [d].
		DT50honey:      60.,   // Honey DT50 [d]

		RUD: 21., // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	pe.Toxicityparams = params_etox.Toxicityparams{
		ForagerOralLD50:  0.127, // dimethoate
		ForagerOralSlope: 4.37,  // dimethoate
		HSuptake:         0.1,   //

		ForagerContactLD50:  0.169, // dimethoate
		ForagerContactSlope: 16.6,  // dimethoate

		LarvaeOralLD50:  0.24,  // dimethoate
		LarvaeOralSlope: 1.186, // dimethoate

		NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
		NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]
	}

	/*
		pe.GUTSParams = params_etox.GUTSParams{ // param estimates for dimethoate; taken from Baas et al. 2022
			Type: "SD",  // GUTS mode
			K_SR: 0.625, //default values taken from Baas et al. 2022
			K_CA: 0.4,   //default values taken from Baas et al. 2022
			T:    10,    // amount of timesteps per day for numeric GUTS approximation

			// IT params
			Kd_IT: 0.012,  // Dominant rate constant for the reduced-IT-model
			MW_IT: 0.0024, // Median of the distribution of thresholds for calculating IT threshold distribution
			F_S:   3.0,    // Fraction spread in distribution of thresholds; used to calculate beta for the threshold distribution calc

			// SD params
			Kd_SD: 0.36,  // Dominant rate constant for the reduced-SD-model
			MW_SD: 0.014, // Median of the distribution of thresholds for calculating h in the red-SD-model
			BW_SD: 21.,   // Killing rate for calculating h in the red-SD-model
		}*/

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p, &pe)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugEcotox{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
