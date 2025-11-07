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
	GUTSParams          GUTSParams
	ConsumptionRework   ConsumptionRework
	Nursing             Nursing
}

// Default returns the complete default parameter set of BEEHAVE.
func Default_etox() DefaultParams_etox {
	return DefaultParams_etox{
		ETOXparams: ETOXparams{
			Application:               false, // Determines if there is an application at all (and turns on/off the necessary code)
			GUTS:                      false, // Determines whether BeeGUTS or dose-response shall be used for effect calculation
			ReworkedThermoETOX:        true,  // Determines whether to use the reworked splitting of thermoregulatory energy cost amongst all IHbees and foragers (true) or to use the classic BEEHAVE_ecotox method to give all of this energy to one squadron/cohort (false)
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
		GUTSParams: GUTSParams{ // default values are taken from supplementary information of Baas et al. 2022; params are specified for dimethoate
			Type: "SD",  // GUTS mode
			K_SR: 0.625, //default values taken from Baas et al. 2022
			K_CA: 0.4,   //default values taken from Baas et al. 2022
			T:    24,    // amount of timesteps per day for numeric GUTS approximation

			// IT params
			Kd_IT: 0.012,  // Dominant rate constant for the reduced-IT-model
			MW_IT: 0.0024, // Median of the distribution of thresholds for calculating IT threshold distribution; adjusted from ng/bee (Baas et al. 2022) to mug/bee as this model uses mug as primary unit
			F_S:   3,      // Fraction spread in distribution of thresholds; used to calculate beta for the threshold distribution calc

			// SD params
			Kd_SD: 0.36,  // Dominant rate constant for the reduced-SD-model
			MW_SD: 0.014, // Median of the distribution of thresholds for calculating h in the red-SD-model; adjusted from ng/bee (Baas et al. 2022) to mug/bee as this model uses mug as primary unit
			BW_SD: 21.,   // Killing rate for calculating h in the red-SD-model; adjusted from 1/(ng/bee d) (Baas et al. 2022) to 1/(mug/bee d) as this model uses mug as primary unit
		},
		ConsumptionRework: ConsumptionRework{
			Nursebeecs: false, // turned off to keep basic model as default

			HoneyAdultWorker:  11., // mg/day <- may need changing; Brodschneider&Crailsheim 2010 quote Barker & Lehner 1974 for 4mg of sugar per day for survival = ca. 5.1mg honey; old BEEHAVE val is 11 (Rortais Winterbees)
			PollenAdultWorker: 1.5, // mg/day <- old value for 14 day old bees from Rortais et al. 2005; should fit as a baseline for now; maybe adjust down the line

			MaxPollenNurse: 6.5, // + 1.5 per adult = 8 mg/day; this should be a field realistic total for a normal peak; Crailsheim reported up to 8 as a max, 12 as the highes statistical 95% bound under controlled conditions. 12 is cited as a maximum in BeeREX model and comes from Rortais et al. 2005 citing Crailsheim et al. 1992;
			MaxHoneyNurse:  60., // mg/day; estimate based on BeeREX citing Rortais et al. 2005; this is however most likely not accurately interpreted by BeeREX authors. This value is probably unknown and 60 is probably far too high. It should not really matter though, as pollen intake will regulate amount of nurses and honey intake will be emergent property

			HoneyAdultDrone:  10., // mg/day; taken from BEEHAVE, though the origin of this value is very unclear; this might be another placeholder
			PollenAdultDrone: 2.,  // mg/day; taken from BEEHAVE, already just a rough estimate; there appears to be no clear value anywhere, though it is known that they have an increased need for the first 9 days to reach maturity

			HoneyWorkerLarva:  make([]float64, 6), // gets initialized in sys.init_etox for now because I do not know how else to do this
			PollenWorkerLarva: make([]float64, 6), // gets initialized in sys.init_etox for now because I do not know how else to do this
			HWLtotal:          65.4,               // mg over a total of 6 days --> old BEEHAVE value; an increase to 75.5 might make sense, value taken from Rortais et al. 2005
			PWLtotal:          100.,               // mg over a total of 6 days; this is a lowered estimate as opposed to original BEEHAVE (used 142 mg), because some of the budged gets shifted to the first few days of adult development
			PFPworker:         42,                 // mg over the first 4 days of life; this gets taken in by nurses if possible, if not the bees eat it themselves. 42 mg makes the pollen budget turn out exactly the same as before
			HoneyDirect:       0.05,               // unknown, therefore same estimate as pollen value below for now
			PollenDirect:      0.05,               // 5% of pollen get taken in directly, estimated by Hrassnigg & Crailsheim (2005); for now assumes the same value for worker and drone larvae from day 3 onwards and thus has a bigger effect on drones that take 1 day longer to pupation
			// this should be okay though as many studies found that drone food has higher residues than worker food, which in turn is higher than royal jelly of queen larvae (example: Wueppenhorst et al. 2024). This should be a somewhat realistic and defensible assumption.

			HoneyDroneLarva:  make([]float64, 7), // gets initialized in sys.init_etox for now because I do not know how else to do this
			PollenDroneLarva: make([]float64, 7), // gets initialized in sys.init_etox for now because I do not know how else to do this
			HDLtotal:         124.9,              // mg over a total of 7 days; the old BEEHAVE value, taken from Rortais et al. 2005
			PDLtotal:         250.,               // mg over a total of 7 days; there is no proper estimate, this is lowered as opposed to original BEEHAVE (used 350 mg), because that is most likely too high. It is estimated that drones larvae weigh 1.8 - 2.6 more than workers (Hrassnigg and Crailsheim 2005)
			PFPdrone:         100.,               // mg over the first 9 days of adult life; this gets taken in by nurses, as drones do not really eat any pollen by themselves. 100 mg makes the pollen budget turn out exactly the same as before
			// Hrassnigg and Crailsheim (2005) use the same values for carbohydrates as Rortais for both larvae, but use a higher pollen budged than I estimated here. I could also simply adopt their budgets, but that would not be completely biologically accurate for modeling dynamics,
			// because both worker and drones have an increased need of pollen after emerging (worker for 3-5 days, drones for ca. 8-10 days) to reach complete maturity. The authors mention this as well, it is hard to estimate how much of the budget is allocated to priming as adults.

			DynamicProteinNursing: false, // determines if nursing capability shall be dynamically adjusted based on worker age (there is evidence of a peaking in capability by age 6-10 roughly)
			Nursingcapabiliies:    make([]float64, 51),
		},
		Nursing: Nursing{
			MinWL_ratio:            2,                                           // might not be used for now; see Eischen et al. 1982, 1983, 1984; placeholder for now but it seems 2:1 W:L gives a good efficiency baseline for rearing, where adult longevity is somewhat as expected
			NurseAgeCeiling:        13,                                          // default age at which nurses stop working as nurses, unless model dynamics increase this
			BroodCannibalismChance: []float64{0.1, 0.3, 0.5, 0.075, 0., 0., 0.}, // just an assumption to start with, based on data from Schmickl&Crailsheim (2001, 2002)

			StartWinterBees: true,
			NewBroodCare:    false,
			ScrambleComp:    false,

			HPGeffects:   false,
			HPGthreshold: []float64{0.000727 / 12.78, 0.000727 / 12.78 * 10, 0.000727 / 12.78 * 100}, // authors used 1 mug/L Clothianidin in 400 ml of Apiinvert; Apiinvert has 1 kg of sugar per Liter and a concentration of 72.7% (https://www.beefeed.com/en/apiinvert/) --> 1 l of Apiinvert should be equivalkent to 1/0.727 = 1.376 kg
			// therefore 1 mug/L (w/v) should be equivalent to 0.727 mug/kg (w/w) in apiinvert; 0.727 mug/kg = 0.727 ng/g = 0.000727 mug/g; BEEHAVE needs values in weight per kJ, therefore we need to consider honey energy content of 12.78 kJ/g
			ProteinFactorNurseExposed: []float64{0.82, 0.77, 0.}, // very much experimental; straight up taken from Schott et al. 2021
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
	ecs.AddResource(world, &pCopy.GUTSParams)
	ecs.AddResource(world, &pCopy.ConsumptionRework)
	ecs.AddResource(world, &pCopy.Nursing)
}
