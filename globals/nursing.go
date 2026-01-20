package globals

import "github.com/mlange-42/ark/ecs"

// introduce all the necessary globals to calculate and track nursing dynamics
type NursingGlobals struct {
	SuffNurses        bool // were there sufficient nurses during the last day?
	Reductionpossible bool // is a reduction in the nursing force possible?
	AbortNursing      bool // can only be turned true if there is a severe lack of nurses and makes all brood starve next day
	KillDrones        bool // can only be turned on once the amount of available nurses gets reduced below the minimum threshold (200 atm, based on Kama and Shpigler 2025); kills all adult drones that are still dependent on being fed

	LastPollenInflux   int     // days/ticks since the lase pollen influx --> regulates cannibalism in times of reudced pollen income
	PollenStoreLastDay float64 // amount of pollen in store yesterday; used to calculate if any fresh pollen were added without changing foraging submodule

	NurseAgeMax   int     // current upper bound age of nurses
	NurseWorkLoad float64 // ratio of protein nurses need to provide to the amount they on average can provide as a metric for their workload

	Total_honey           float64 // amount of total honey to be taken in by nurses on top of their own intake this timestep [mg/d]
	Total_pollen          float64 // amount of total pollen to be taken in by nurses on top of their own intake this timestep [mg/d]
	CurrentMaxPollenNurse float64 // maximum intake of pollen at the current point in time; this can might be reduced during PPP exposure because of HG effects

	WLHoney  float64 // amount of direct honey worker larvae take in this timestep [mg/d]
	WLPollen float64 // amount of direct pollen worker larvae take in this timestep [mg/d]
	DLHoney  float64 // amount of direct honey drone larvae take in this timestep [mg/d]
	DLPollen float64 // amount of direct pollen drone larvae take in this timestep [mg/d]

	WorkerPriming float64 // amount of pollen needed for young workers to prime their HPG this timestep [mg/d]

	WinterBees []ecs.Entity // saves all the current winterbees
	Reverted   []ecs.Entity // saves all the current reverted foragers

	SquadstoReduce int // amount of reverted forager squadrons that will be returned to normal forager squadrons next timestep
}

// track nursing related consumption and population stats; the ints below are also used for regulations in nursing subsystems
type NursingStats struct {
	MaxPollenIntake  float64 // current maximum of pollen intake per nurse
	MeanPollenIntake float64 // current mean pollen intake per nurse
	MaxHoneyIntake   float64 // current maximum of pollen intake per nurse
	MeanHoneyIntake  float64 // current mean pollen intake per nurse

	NL_ratio float64 // current nurse:larva ratio; probably only to be used as a analytical metric

	NonNurseIHbees int     // current amount of non-nursing IHbees --> very young or middle aged bees, that neither forage nor nurse
	IHbeeNurses    int     // current amount of IHbee nurses
	TotalNurses    int     // total amount of nurses today; including reverted foragers and winterbees
	NurseFraction  float64 // current fraction of nurses from total adult worker population (IHbees + foragers)

	RevertedForagers int // current amount of reverted foragers
	WinterBees       int // current amount of winterbees
}
