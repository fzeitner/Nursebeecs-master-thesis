package globals

import "github.com/mlange-42/ark/ecs"

// introduce all the necessary globals to calculate and track nursing dynamics
type NursingGlobals struct {
	SuffNurses        bool // were there sufficient nurses last consumption proc?
	Reductionpossible bool // is a reduction in the nursing force possible?
	AbortNursing      bool // can only be turned true if there is a severe lack of nurses and makes all brood starve next day
	KillDrones        bool // can only be turned on once the amount of available nurses gets reduced below the minimum threshold (200 atm, based on Kama and Shpigler 2025); kills all adult drones that are still dependent on being fed

	LastPollenInflux   int     // days/ticks since the lase pollen influx --> regulates cannibalism in times of reudced pollen income
	PollenStoreLastDay float64 // amount of pollen in store yesterday to calculate if any fresh pollen were added without changing foraging submodule

	NurseAgeMax   int     // current upper bound age of nurses
	NurseWorkLoad float64 // ratio of protein nurses need to provide to the amount they on average can provide as a metric for their workload

	Total_honey           float64 // amount of total honey to be taken in by nurses on top of their own intake this timestep [mg/d]
	Total_pollen          float64 // amount of total pollen to be taken in by nurses on top of their own intake this timestep [mg/d]
	CurrentMaxPollenNurse float64 // maximum intake of pollen at the current point in time; this might be reduced during PPP exposure because of HG effects

	WLHoney  float64 // amount of direct honey worker larvae take in this timestep [mg/d]
	WLPollen float64 // amount of direct pollen worker larvae take in this timestep [mg/d]
	DLHoney  float64 // amount of direct honey drone larvae take in this timestep [mg/d]
	DLPollen float64 // amount of direct pollen drone larvae take in this timestep [mg/d]

	WorkerPriming float64 // amount of pollen needed for young workers to prime their HPG this timestep [mg/d]

	WinterBees []ecs.Entity // saves all the current winterbees and could be used to specify their consumption/jobs further
	Reverted   []ecs.Entity // saves all the current reverted foragers

	SquadstoReduce int // amount of squads that can be reduced again --> amount of reverted forager squadrons that will be returned to normal forager squadrons next timestep
}

// track nursing related consumption and population stats; the ints below are also used for regulations in nursing subsystems
type NursingStats struct {
	MaxPollenIntake  float64 // maximum of pollen intake per nurse
	MeanPollenIntake float64 // mean pollen intake per nurse
	MaxHoneyIntake   float64 // maximum of pollen intake per nurse
	MeanHoneyIntake  float64 // mean pollen intake per nurse

	NL_ratio float64 // current nurse:larva ratio; probably only to be used as a analytical metric

	NonNurseIHbees int     // amount of not nursing IHbees --> MAB, that do nor forage and do not nurse
	IHbeeNurses    int     // amount of IHbee nurses
	TotalNurses    int     // total amount of nurses today; including reverted foragers and winterbees
	NurseFraction  float64 // fraction of nurses from total adult population

	RevertedForagers int // amount of reverted foragers at this current tick
	WinterBees       int // amount of winterbees at this current tick
}
