package params_etox

// ETOX parameters regarding application of the pesticide.
type ETOXparams struct {
	Application               bool // Determines if there is an application at all at any point in the model and if the _ecotox-module should be turned on for all purposes
	GUTS                      bool // Determines whether BeeGUTS shall be used (true) or the old dose-response relationship
	ReworkedThermoETOX        bool // Determines whether to use the reworked splitting of thermoregulatory energy cost amongst all IHbees and foragers (true) or to use the classic BEEHAVE_ecotox method to give all of this energy to one squadron/cohort (false)
	ForagerImmediateMortality bool // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
	DegradationHoney          bool // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
	ContactSum                bool // Determines whether contact exposures of different flower visits shall be summed up.
	ContactExposureOneDay     bool // Determines whether contact exposure shall only be relevant on the one day of application. NEED TO TEST IF THIS MAKES SENSE WITH GUTS VERSION

	PPPname                string  // Identifier for the PPP used.
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kg]
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/kg]
	PPPcontactExposure     float64 // PPP concentration for contact exposure on patch [kg/ha]

	AppDay         int     // Day of the year in which application starts [d].
	ExposurePeriod int     // Duration of exposure happening (irrespective of DT50) [d].
	SpinupPhase    int     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
	ExposurePhase  int     // Number of years in which exposure takes place [y].
	DT50           float64 // Whole plant DT50 from residue studies [d].
	RUD            float64 // Residue per Unit Dose  [(ha*mg)/(kg*kg)].
	DT50honey      float64 // Honey DT50 [d].

	ETOXDensityOfHoney float64 // The density of honey is 1.4 [kg/l].
}

// parameters regarding the toxicity of the pesticide and the tolerance/uptake of the foragers and cohorts.
type Toxicityparams struct {
	ForagerOralLD50  float64 // Lethal oral dose for 50% mortality of foragers [µg/bee].
	ForagerOralSlope float64 // Slope of the dose-response relationship (forager, oral) [ ].
	HSuptake         float64 // Uptake of a given percentage of ai in the honey stomach by the forager bees

	ForagerContactLD50  float64 // Lethal dose for 50% of foragers via contact exposure [µg/bee].
	ForagerContactSlope float64 // Slope of the dose-response relationship (forager, contact) [ ].

	LarvaeOralLD50  float64 // Lethal oral dose for 50% mortality of larvae [µg/larvae].
	LarvaeOralSlope float64 // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented.

	NursebeesNectar float64 // Factor describing the filter effect of nurse bees for nectar [ ].
	NursebeesPollen float64 // Factor describing the filter effect of nurse bees for pollen [ ].
}

// WaterForaging parameters. Not used in the current state of the model.
type WaterParams struct {
	WaterForaging             bool    // Determines whether water foraging takes place or not for the ecotox processes involved.
	ETOX_cropvolume_water     float64 // The amount of water in [g] which is carried by one forager, calculated from 44 mg (Visscher et al. 1996)
	ETOX_Watertripsperh       float64 // The number of trips per h for a water forager: 7 Robinson et al. 1984
	ETOX_Durationofwatertrips float64 // The number of h in the morning they are only looking for water 2 h Kuhlholtz & Seeley 1997
	ETOX_PROBWatercollection  float64 // The probability that a forager gets a water forager
	ETOX_handlingTimeWater    float64
}

// WaterForagingPeriod parameters.
//
// Data read from files (field Files) is appended to data provided directly (field Years).
type WaterForagingPeriod struct {
	Years       [][]float64 // Foraging period per day [h] as raw data. Each row must have a whole-numbered multiple of 365 entries.
	Files       []string    // Files with daily foraging period data to use.
	Builtin     bool        // Whether the used files are built-in. Use local files otherwise.
	RandomYears bool        // Whether to randomize years.
}

// GUTS-RED parameters to implement BeeGUTS SD and IT into the mortality functions
type GUTSParams struct {
	Type string  // activate SD/IT here
	K_SR float64 // honey stomach release rate [d**-1]
	K_CA float64 // contact availability rate [d**-1]
	T    int     // amount of timesteps per day for numeric GUTS approximation

	// IT params
	Kd_IT float64 // Dominant rate constant for the reduced-IT-model
	MW_IT float64 // Median of the distribution of thresholds for calculating IT threshold distribution
	F_S   float64 // Fraction spread in distribution of thresholds; used to calculate beta for the threshold distribution calc

	// SD params
	Kd_SD float64 // Dominant rate constant for the reduced-SD-model
	MW_SD float64 // Median of the distribution of thresholds for calculating h in the red-SD-model
	BW_SD float64 // Killing rate for calculating h in the red-SD-model
}

// this contains all the adjusted new estimates for consumption that became necessary in the nurse rework
type ConsumptionRework struct {
	Nursebeecs bool // switch to turn on this rework of consumption behavior via nursebees

	HoneyAdultWorker  float64 // honey intake that each adult worker bee takes in (also the baseline for nurses)
	PollenAdultWorker float64 // pollen intake that each adult worker bee takes in (also the baseline for nurses)

	MaxPollenNurse float64 // maximum of pollen that nurses can theoretically take in
	MaxHoneyNurse  float64 // maximum of honey that nurses can theoretically take in

	HoneyAdultDrone  float64 // daily intake needs for drones
	PollenAdultDrone float64 // daily intake needs for drones

	HoneyWorkerLarva  []float64 // reworked honey needs per larva per day; now differentiates between each larval stage
	PollenWorkerLarva []float64 // reworked pollen needs per larva per day; now differentiates between each larval stage
	HWLtotal          float64   // HoneyWorkerLarva_total --> total amount of honey necessary to rear one worker larva
	PWLtotal          float64   // PollenWorkerLarva_total --> total amount of pollen necessary to rear one worker larva
	PFPworker         float64   // PollenForPriming of HG (hypopharyngeal glands) of workers, added on consumption over the first 4 days of adult life
	HoneyDirect       float64   // fraction of direct honey intake per larva from age 3 onwards
	PollenDirect      float64   // fraction of direct pollen intake per larvae from age 3 onwards

	HoneyDroneLarva  []float64 // reworked honey needs per larva per day; now differentiates between each larval stage
	PollenDroneLarva []float64 // reworked pollen needs per larva per day; now differentiates between each larval stage
	HDLtotal         float64   // HoneyDroneLarva_total --> total amount of honey necessary to rear one drone larva
	PDLtotal         float64   // PollenDroneLarva_total --> total amount of pollen necessary to rear one drone larva
	PFPdrone         float64   // PollenForPriming sexual maturity in drones, added on baseline consumption over the first 9 days of adult life

	DynamicProteinNursing bool      // switch to turn on dynamic nursing capability
	Nursingcapabiliies    []float64 // this is an array full of factors defining efficiency/capabiliy of the nurse cohort depending on age
}

type Nursing struct {
	MinWL_ratio            float64   // target of minimum necessary worker:larva ratio that the colony tries to go back to if possible; goes back to Eischen et al. 1982, 1983, 1984
	NurseAgeCeiling        int       // baseline age until model assumes that workers will act as nurses
	BroodCannibalismChance []float64 // chance of brood to be cannibalized based on age (Schmickl & Crailsheim 2001,2002)
	NurseWorkLoadTH        float64   // threshold of nurse workload above which ProteinFactorNurses gets reduced

	StartWinterBees bool // switch to turn starting foragers into winter bees --> necessary if we start simulating at the beginning of the year
	NewBroodCare    bool // switch to turn on new nurse based brood care mechanism (i.e. killing of brood based on nursing capacitys)
	ScrambleComp    bool // switch to turn on scramble competition mechanism within new brood care which kicks in if nurse workload is too high

	HPGeffects                bool      // switch to turn on reduced brood care capabilies from PPP induced reduced HPG activity
	HPGthreshold              []float64 // threshold value of PPP necessary in honey to activate the HPG effects
	ProteinFactorNurseExposed []float64 // new ceiling for ProteinFactorNurses in case HPG effects are turned on and the threshold is exceeded
}
