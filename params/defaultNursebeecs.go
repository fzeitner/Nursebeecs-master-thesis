package params

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/mlange-42/ark/ecs"
)

// ParamsNursebeecs is an interface for nursebeecs parameter sets.
type ParamsNursebeecs interface {
	// Apply the parameters to a world.
	Apply(world *ecs.World)
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSONFile(path string) error
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSON(data []byte) error
}

// DefaultParamsNursebeecs contains all default parameters of nursebeecs.
//
// DefaultParamsNursebeecs implements [ParamsNursebeecs].
type DefaultParamsNursebeecs struct {
	ConsumptionRework ConsumptionRework
	NursingRework     NursingRework
}

// DefaultNursebeecs returns the complete default parameter set of nursebeecs.
func DefaultNursebeecs() DefaultParamsNursebeecs {
	return DefaultParamsNursebeecs{
		ConsumptionRework: ConsumptionRework{
			HoneyAdultWorker:  11., // mg/day <- may need changing; Brodschneider&Crailsheim 2010 quote Barker & Lehner 1974 for 4mg of sugar per day for survival = ca. 5.1mg honey; old BEEHAVE val is 11 (Rortais Winterbees)
			PollenAdultWorker: 1.5, // mg/day <- old value for 14 day old bees from Rortais et al. 2005; should fit as a baseline for now; maybe adjust down the line

			MaxPollenNurse: 6.5, // + 1.5 per adult = 8 mg/day; this should be a field realistic total for a normal peak; Crailsheim reported up to 8 as a max, 12 as the highes statistical 95% bound under controlled conditions. 12 is cited as a maximum in BeeREX model and comes from Rortais et al. 2005 citing Crailsheim et al. 1992;
			MaxHoneyNurse:  60., // mg/day; estimate based on BeeREX citing Rortais et al. 2005; this is however most likely not accurately interpreted by BeeREX authors. This value is probably unknown and 60 is probably too high. This should not really matter though, as pollen intake will regulate amount of nurses and honey intake will be an emergent property

			HoneyAdultDrone:  10., // mg/day; taken from BEEHAVE, though the origin of this value is very unclear; this might be another placeholder
			PollenAdultDrone: 2.,  // mg/day; taken from BEEHAVE, already just a rough estimate; there appears to be no clear value anywhere, though it is known that they have an increased need for the first 9 days to reach maturity

			HoneyWorkerLarva:  make([]float64, 6), // gets initialized in sys.initEtox for now because I do not know how else to do this
			PollenWorkerLarva: make([]float64, 6), // gets initialized in sys.initEtox for now because I do not know how else to do this
			HWLtotal:          65.4,               // mg over a total of 6 days --> old BEEHAVE value; an increase to 75.5 might make sense, value taken from Rortais et al. 2005
			PWLtotal:          100.,               // mg over a total of 6 days; this is a lowered estimate as opposed to original BEEHAVE (used 142 mg), because some of the budged gets shifted to the first few days of adult development
			PFPworker:         42,                 // mg over the first 4 days of life; this gets taken in by nurses if possible, if not the bees eat it themselves. 42 mg makes the pollen budget turn out exactly the same as before
			HoneyDirect:       0.05,               // unknown, therefore same estimate as pollen value below for now
			PollenDirect:      0.05,               // 5% of pollen get taken in directly, estimated by Hrassnigg & Crailsheim (2005); for now assumes the same value for worker and drone larvae from day 3 onwards and thus has a bigger effect on drones that take 1 day longer to pupation
			// this should be okay though as many studies found that drone food has higher residues than worker food, which in turn is higher than royal jelly of queen larvae (example: Wueppenhorst et al. 2024). This should be a somewhat realistic and defensible assumption.

			HoneyDroneLarva:  make([]float64, 7), // gets initialized in sys.initEtox for now because I do not know how else to do this
			PollenDroneLarva: make([]float64, 7), // gets initialized in sys.initEtox for now because I do not know how else to do this
			HDLtotal:         124.9,              // mg over a total of 7 days; the old BEEHAVE value, taken from Rortais et al. 2005
			PDLtotal:         250.,               // mg over a total of 7 days; there is no proper estimate, this is lowered as opposed to original BEEHAVE (used 350 mg), because that is most likely too high. It is estimated that drones larvae weigh 1.8 - 2.6 more than workers (Hrassnigg and Crailsheim 2005)
			PFPdrone:         100.,               // mg over the first 9 days of adult life; this gets taken in by nurses, as drones do not really eat any pollen by themselves. 100 mg makes the pollen budget turn out exactly the same as before
			// Hrassnigg and Crailsheim (2005) use the same values for carbohydrates as Rortais for both larvae, but use a higher pollen budged than I estimated here. I could also simply adopt their budgets, but that would not be completely biologically accurate for modeling dynamics,
			// because both worker and drones have an increased need of pollen after emerging (worker for 3-5 days, drones for ca. 8-10 days) to reach complete maturity. The authors mention this as well, it is hard to estimate how much of the budget is allocated to priming as adults.

			DynamicProteinNursing: false,               // determines if nursing capability shall be dynamically adjusted based on worker age (there is evidence of a peaking in capability by age 6-10 roughly)
			Nursingcapabiliies:    make([]float64, 51), // array to save the dynamic capabilities of nurse bees depending on their age
		},
		NursingRework: NursingRework{
			//MinWLRatio:            2,                                            // might not be used for now; see Eischen et al. 1982, 1983, 1984; placeholder for now but it seems 2:1 Worker:Larva gives a good efficiency baseline for rearing, where adult longevity is somewhat as expected
			NurseAgeCeiling:        13,                                           // default age at which nurses stop working as nurses, unless model dynamics increase this
			BroodCannibalismChance: []float64{0.23, 0.3, 0.58, 0.06, 0., 0., 0.}, // based on data from Schmickl&Crailsheim (2001, 2002) and HoPoMo model quoting these studies
			NurseWorkLoadTH:        1.5,                                          // equals 11.25 mg of pollen per day, this should be a reasonable maximum intake for nurse bees (Rortais et al. 2005, Crailsheim et al. 1992)
			MinimumTH:              1.0,                                          // 1.0 equals per calculation of NurseWorkload a reasonable mean intake of nurse bees, because NurseWorkload is designed to represent exactly this

			NewBroodCare:           true,
			Nursebeecsv0:           false,
			Nursebeecsv1:           true,
			ForesightedCannibalism: false,

			HGEffects:    false,
			HGFoodIntake: false,
		},
	}
}

// FromJSONFile fills the parameter set with values from a JSON file.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParamsNursebeecs) FromJSONFile(path string) error {
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
func (p *DefaultParamsNursebeecs) FromJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(p)
}

// Apply the parameters to a world by adding them as resources.
func (p *DefaultParamsNursebeecs) Apply(world *ecs.World) {
	pCopy := *p

	// Resources
	ecs.AddResource(world, &pCopy.ConsumptionRework)
	ecs.AddResource(world, &pCopy.NursingRework)
}
