package params

// New estimates for consumption that became necessary to create Nursebeecs
type ConsumptionRework struct {
	HoneyAdultWorker  float64 // honey intake that each adult worker bee takes in (also the baseline for nurses) [mg/d]
	PollenAdultWorker float64 // pollen intake that each adult worker bee takes in (also the baseline for nurses) [mg/d]

	MaxPollenNurse float64 // maximum of pollen that nurses can theoretically take in [mg/d]
	MaxHoneyNurse  float64 // maximum of honey that nurses can theoretically take in [mg/d]

	HoneyAdultDrone  float64 // daily intake needs for drones [mg/d]
	PollenAdultDrone float64 // daily intake needs for drones [mg/d]

	HoneyWorkerLarva  []float64 // reworked honey needs per larva per day; now differentiates between each larval stage [mg/d]
	PollenWorkerLarva []float64 // reworked pollen needs per larva per day; now differentiates between each larval stage [mg/d]
	HWLtotal          float64   // HoneyWorkerLarva_total --> total amount of honey necessary to rear one worker larva [mg/d]
	PWLtotal          float64   // PollenWorkerLarva_total --> total amount of pollen necessary to rear one worker larva [mg]
	PFPworker         float64   // PollenForPriming of HG (hypopharyngeal glands) of workers, added on consumption over the first 4 days of adult life [mg/d]
	HoneyDirect       float64   // fraction of direct honey intake per larva from age 3 onwards [-]
	PollenDirect      float64   // fraction of direct pollen intake per larvae from age 3 onwards [-]

	HoneyDroneLarva  []float64 // reworked honey needs per larva per day; now differentiates between each larval stage [%/d]
	PollenDroneLarva []float64 // reworked pollen needs per larva per day; now differentiates between each larval stage [%/d]
	HDLtotal         float64   // HoneyDroneLarva_total --> total amount of honey necessary to rear one drone larva [mg]
	PDLtotal         float64   // PollenDroneLarva_total --> total amount of pollen necessary to rear one drone larva [mg]
	PFPdrone         float64   // PollenForPriming sexual maturity in drones, added on baseline consumption over the first 9 days of adult life [mg]

	DynamicProteinNursing bool      // switch to turn on dynamic nursing capability
	Nursingcapabiliies    []float64 // this is an array full of factors defining efficiency/capabiliy of the nurse cohort depending on age [-]
}

// Switches and key parameters for Nursebeecs that are not strictly consumption related
type NursingRework struct {
	NurseAgeCeiling        int       // baseline age until model assumes that workers will act as nurses
	BroodCannibalismChance []float64 // chance of brood to be cannibalized based on age (Schmickl & Crailsheim 2001,2002)
	NurseWorkLoadTH        float64   // threshold of nurse workload above which ProteinFactorNurses gets reduced
	MinimumTH              float64   // threshold of nurse workload below which ProteinFactorNurses is allowed to recover

	NewBroodCare           bool // switch to turn on new nurse based brood care mechanism (i.e. killing of brood based on nursing capacitys)
	Nursebeecsv0           bool // switch to turn on Nbeecs v0.5 --> first attempt at coupling of nurseworkload ending up with large fluctuations
	Nursebeecsv1           bool // switch to turn on Nbeecs v.1 --> coupling of nurseworkload to ProteinFactorNurses
	ForesightedCannibalism bool // switch to turn on ForesightedCannibalism based on Schmickl&Crailsheim 2001&2002--> cannibalization depending on time passed since last pollen influx

	HGEffects    bool // switch to turn on reduced brood care capabilies from PPP induced reduced HPG activity
	HGFoodIntake bool // switch to turn on reduced maximum food intake capability as a PPP induced sublethal effect
}
