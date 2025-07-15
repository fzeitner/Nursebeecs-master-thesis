package params_etox

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/mlange-42/ark/ecs"
)

// Params is an interface for parameter sets.
type Params_etox interface {
	// Apply the parameters to a world.
	Apply(world *ecs.World)
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSONFile(path string) error
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSON(data []byte) error
}

// DefaultParams contains all default parameters of BEEHAVE.
//
// DefaultParams implements [Params].
type DefaultParams_etox struct {
	ETOXparams          ETOXparams
	WaterParams         WaterParams
	WaterForagingPeriod WaterForagingPeriod
	Toxicityparams      Toxicityparams
}

// Default returns the complete default parameter set of BEEHAVE.
func Default_etox() DefaultParams_etox {
	return DefaultParams_etox{
		ETOXparams: ETOXparams{
			Application:               false,
			ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
			DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
			ContactSum:                false, // Determines whether contact exposure should be summed up per visit to a patch (true) or if the mean should be calculated whenever a new patch is visited (false)
			ContactExposureOneDay:     false, // Determines whether contact exposure should only be possible on the day of application

			PPPname:                "No applications", // Identifier for the PPP used.
			PPPconcentrationNectar: 990,
			PPPconcentrationPollen: 26631,
			PPPcontactExposure:     0.3, // actually dont have a value for this

			AppDay:         189,   // Day of the year in which application starts [d].
			ExposurePeriod: 8,     // Duration of exposure happening (irrespective of DT50) [d].
			SpinupPhase:    0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
			ExposurePhase:  3,     // Number of years in which exposure takes place [y].
			DT50:           1000., // Whole plant DT50 from residue studies [d].
			RUD:            21.,   // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
			DT50honey:      60.,   // Honey DT50 [d]

			ETOXDensityOfHoney: 1.4, // [kg/l]
		},
		Toxicityparams: Toxicityparams{
			ForagerOralLD50:  1000., // Lethal oral dose for 50% mortality of foragers [µg/bee].
			ForagerOralSlope: 100.,  // Slope of the dose-response relationship (forager, oral) [ ].
			HSuptake:         0.1,   // Uptake of a given percentage of ai in the honey stomach by the forager bees

			ForagerContactLD50:  0.6,  // Lethal dose for 50% of foragers via contact exposure [µg/bee]
			ForagerContactSlope: 1.08, // Slope of the dose-response relationship (forager, contact) [ ]

			LarvaeOralLD50:  0.0014, // Lethal oral dose for 50% mortality of larvae [µg/larvae]
			LarvaeOralSlope: 1.6,    // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented

			NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
			NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]
		},
		WaterParams: WaterParams{
			WaterForaging:             false,       // Determines whether water foraging takes place or not.
			ETOX_cropvolume_water:     44. / 1000., // [g]: 44 mg water per forager Visscher et al. 1996
			ETOX_Watertripsperh:       7.,          // Number of trips per h for water foragers
			ETOX_Durationofwatertrips: 1.,          // How long do the bees go out in the morning exclusively searching for water 1 h according to Lindauer 1954
		},
		WaterForagingPeriod: WaterForagingPeriod{
			Files:       []string{"ETOX_waterforcooling_daily/waterlistExample.txt"},
			Builtin:     true,
			RandomYears: false,
		},
	}
}

// FromJSONFile fills the parameter set with values from a JSON file.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParams_etox) FromJSONFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return p.FromJSON(content)
}

// FromJSON fills the parameter set with values from JSON.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParams_etox) FromJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(p)
}

// Apply the parameters to a world by adding them as resources.
func (p *DefaultParams_etox) Apply(world *ecs.World) {
	pCopy := *p

	// Resources
	ecs.AddResource(world, &pCopy.WaterForagingPeriod)
	ecs.AddResource(world, &pCopy.ETOXparams)
	ecs.AddResource(world, &pCopy.Toxicityparams)
	ecs.AddResource(world, &pCopy.WaterParams)
}
