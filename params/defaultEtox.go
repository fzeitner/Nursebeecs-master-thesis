package params

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/mlange-42/ark/ecs"
)

// ParamsEtox is an interface for the beecs_ecotox parameter sets.
type ParamsEtox interface {
	// Apply the parameters to a world.
	Apply(world *ecs.World)
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSONFile(path string) error
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSON(data []byte) error
}

// DefaultParamsEtox contains all default parameters of BEEHAVE_ecotox.
//
// DefaultParamsEtox implements [ParamsEtox].
type DefaultParamsEtox struct {
	PPPApplication      PPPApplication
	WaterForaging       WaterForaging
	WaterForagingPeriod WaterForagingPeriod
	PPPToxicity         PPPToxicity
}

// DefaultEtox returns the complete default parameter set for beecs_ecotox. ReworkedThermoEtox, RealisticStoch and the two fixes are additions created by me.
func DefaultEtox() DefaultParamsEtox {
	return DefaultParamsEtox{
		PPPApplication: PPPApplication{
			Application:               false, // Determines whether there is an application at all (and turns on/off the necessary code)
			ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
			DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
			ContactSum:                false, // Determines whether contact exposure should be summed up per visit to a patch (true) or if the mean should be calculated whenever a new patch is visited (false)
			ContactExposureOneDay:     false, // Determines whether contact exposure should only be possible on the day of application

			RealisticStoch:     false, // Determines whether stochstic death for low numbers of IHbees in one cohort shall be made more realistic by calculating a chance for each bee
			ReworkedThermoETOX: false, // Determines whether thermoregulation energy shall be taken in equally by all adult bees (True, new version) or if one cohort/squad shall take it all (false; Netlogo version)
			Nursebeefix:        true,  // Determines whether the nurse bee intake from BEEHAVE_ecotox's nursebeefactors shall be added to IHbees instead of dissipating
			HSUfix:             true,  // Determines if the PPP lost to the second call of HSuptake when unloading nectar shall be redirected to IHbees (true) insted of dissipating

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
		PPPToxicity: PPPToxicity{
			ForagerOralLD50:  1000., // Lethal oral dose for 50% mortality of foragers [µg/bee].
			ForagerOralSlope: 100.,  // Slope of the dose-response relationship (forager, oral) [ ].
			HSuptake:         0.1,   // Uptake of a given percentage of ai in the honey stomach by the forager bees

			ForagerContactLD50:  0.6,  // Lethal dose for 50% of foragers via contact exposure [µg/bee]
			ForagerContactSlope: 1.08, // Slope of the dose-response relationship (forager, contact) [ ]

			LarvaeOralLD50:  0.0014, // Lethal oral dose for 50% mortality of larvae [µg/larvae]
			LarvaeOralSlope: 1.6,    // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented

			NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
			NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]

			HGthreshold: []float64{0.000727 / 12.78, 0.000727 / 12.78 * 10, 0.000727 / 12.78 * 100}, // authors used 1 mug/L Clothianidin in 400 ml of Apiinvert; Apiinvert has 1 kg of sugar per Liter and a concentration of 72.7% (https://www.beefeed.com/en/apiinvert/) --> 1 l of Apiinvert should be equivalkent to 1/0.727 = 1.376 kg
			// therefore 1 mug/L (w/v) should be equivalent to 0.727 mug/kg (w/w) in apiinvert; 0.727 mug/kg = 0.727 ng/g = 0.000727 mug/g; BEEHAVE needs values in weight per kJ, therefore we need to consider honey energy content of 12.78 kJ/g
			ProteinFactorNurseExposed: []float64{0.82, 0.77, 0.}, // very much experimental; straight up taken from Schott et al. 2021
			MaxPollenRed:              []float64{0.9, 0.8, 0.25}, // very much experimental and just a guess for now, needs to be calibrated probperly for each PPP
		},
		WaterForaging: WaterForaging{
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
func (p *DefaultParamsEtox) FromJSONFile(path string) error {
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
func (p *DefaultParamsEtox) FromJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(p)
}

// Apply the parameters to a world by adding them as resources.
func (p *DefaultParamsEtox) Apply(world *ecs.World) {
	pCopy := *p

	// Resources
	ecs.AddResource(world, &pCopy.WaterForagingPeriod)
	ecs.AddResource(world, &pCopy.PPPApplication)
	ecs.AddResource(world, &pCopy.PPPToxicity)
	ecs.AddResource(world, &pCopy.WaterForaging)
}
