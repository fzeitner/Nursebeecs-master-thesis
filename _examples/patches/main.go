// Demonstrates how to parametrize patches.
package main

import (
	"github.com/fzeitner/beecs_ecotox/comp"
	"github.com/fzeitner/beecs_ecotox/model"
	"github.com/fzeitner/beecs_ecotox/obs"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	// Get the default parameters.
	p := params.Default()

	// Change initial patches
	p.InitialPatches = params.InitialPatches{
		Patches: []comp.PatchConfig{
			// A patch with constant resource availability.
			{
				DistToColony: 1000,
				ConstantPatch: &comp.ConstantPatch{
					Nectar:               20,  // [L]
					Pollen:               1,   // [kg]
					NectarConcentration:  1.5, // [mumol/L]
					DetectionProbability: 0.2,
				},
			},
			// A patch with seasonal resource availability.
			{
				DistToColony: 1000,
				SeasonalPatch: &comp.SeasonalPatch{
					MaxNectar:            20,  // [L]
					MaxPollen:            1,   // [kg]
					NectarConcentration:  1.5, // [mumol/L]
					DetectionProbability: 0.2,

					SeasonShift: 20, // [d]
				},
			},
			// A patch with scripted resource availability.
			/*{
				DistToColony: 1000,
				ScriptedPatch: &comp.ScriptedPatch{
					Nectar: [][2]float64{
						{0, 0},
						{100, 20},
						{250, 0},
					},
					Pollen: [][2]float64{
						{0, 0},
						{100, 1},
						{250, 0},
					},
					NectarConcentration: [][2]float64{
						{0, 1.5},
					},
					DetectionProbability: [][2]float64{
						{0, 0.2},
					},
					Interpolation: interp.Step,
				},
			},*/ // FIGURE OUT WHY THIS ISNT WORKING LATER
		},
	}
	p.Termination.MaxTicks = 3650 // to check multiyear input of PPP

	p.ETOXparams = params.ETOXparams{
		Application:    true,
		PPPname:        "No applications", // Identifier for the PPP used.
		DT50:           1000.,             // Whole plant DT50 from residue studies [d].
		AppDay:         3,                 // Day of the year in which application starts [d].
		ExposurePeriod: 8,                 // Duration of exposure happening (irrespective of DT50) [d].
		SpinupPhase:    0,                 // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
		ExposurePhase:  3,                 // Number of years in which exposure takes place [y].

		PPPconcentrationNectar: 990,
		PPPconcentrationPollen: 26631,
		PPPcontactExposure:     0.3, // actually dont have a value for this

		ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
		DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
		ContactSum:                false, // Determines whether contact exposures of different flower visits shall be summed up.
		ContactExposureOneDay:     false,

		ForagerOralLD50:  1000., // Lethal oral dose for 50% mortality of foragers [µg/bee].
		ForagerOralSlope: 100.,  // Slope of the dose-response relationship (forager, oral) [ ].
		HSuptake:         0.1,   // Uptake of a given percentage of ai in the honey stomach by the forager bees

		ForagerContactLD50:  0.6,  // Lethal dose for 50% of foragers via contact exposure [µg/bee]
		ForagerContactSlope: 1.08, // Slope of the dose-response relationship (forager, contact) [ ]

		LarvaeOralLD50:  0.0014, // Lethal oral dose for 50% mortality of larvae [µg/larvae]
		LarvaeOralSlope: 1.6,    // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented

		NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
		NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]
		DT50honey:       60.,  // Honey DT50 [d]
		RUD:             21.,  // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
	}

	// Create a model with the default sub-models.
	m := model.Default(&p, nil)

	// Add a CSV outputs for patch nectar and pollen.
	m.AddSystem(&reporter.CSV{
		Observer: &obs.PatchNectar{},
		File:     "out/nectar.csv",
	})
	m.AddSystem(&reporter.CSV{
		Observer: &obs.PatchPPPNectar{},
		File:     "out/PPPnectar.csv",
	})

	m.AddSystem(&reporter.CSV{
		Observer: &obs.PatchPPPPollen{},
		File:     "out/PPPPollen.csv",
	})

	m.AddSystem(&reporter.CSV{
		Observer: &obs.PatchPPPcontact{},
		File:     "out/PPPcontact.csv",
	})

	m.AddSystem(&reporter.CSV{
		Observer: &obs.PatchPollen{},
		File:     "out/pollen.csv",
	})

	// Add a CSV outputs for patch nectar and pollen visits.
	m.AddSystem(&reporter.CSV{
		Observer: &obs.NectarVisits{},
		File:     "out/nectar-visits.csv",
	})
	m.AddSystem(&reporter.CSV{
		Observer: &obs.PollenVisits{},
		File:     "out/pollen-visits.csv",
	})

	// Run the model.
	m.Run()
}
