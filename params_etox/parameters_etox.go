package params_etox

// ETOX parameters regarding application of the pesticide.
type ETOXparams struct {
	Application               bool // Determines if there is an application at all at any point in the model and if the _ecotox-module should be turned on for all purposes
	ForagerImmediateMortality bool // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
	DegradationHoney          bool // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
	ContactSum                bool // Determines whether contact exposures of different flower visits shall be summed up.
	ContactExposureOneDay     bool // Determines whether contact exposure shall only be relevant on the one day of application

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
