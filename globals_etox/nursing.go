package globals_etox

import "github.com/mlange-42/ark/ecs"

// this will keep all the necessary globals to calculate and track nursing dynamics

type Nursing_globals struct {
	SuffNurses        bool // were there sufficient nurses last consumption proc?
	Reductionpossible bool // is a reduction in the nursing force possible?
	AbortNursing      bool // can only be turned true if there is a severe lack of nurses and makes all brood starve next day

	LastPollenInflux   int     // days/ticks since the lase pollen influx --> regulates cannibalism in times of reudced pollen income
	PollenStoreLastDay float64 // amount of pollen in store yesterday to calculate if any fresh pollen were added without changing foraging submodule

	NurseAgeMax   int     // current upper bound age of nurses
	NurseWorkLoad float64 // ratio of protein nurses need to provide to the amount they on average can provide as a metric for their workload

	Total_honey  float64 // amount of total honey to be taken in by nurses on top of their own intake this timestep
	Total_pollen float64 // amount of total pollen to be taken in by nurses on top of their own intake this timestep

	WLHoney  float64 // amount of direct honey worker larvae take in this timestep
	WLPollen float64 // amount of direct pollen worker larvae take in this timestep
	DLHoney  float64 // amount of direct honey drone larvae take in this timestep
	DLPollen float64 // amount of direct pollen drone larvae take in this timestep

	WorkerPriming float64 // amount of pollen needed for young workers to prime their HPG this timestep

	WinterBees []ecs.Entity // saves all the current winterbees and could be used to specify their consumption/jobs further
	Reverted   []ecs.Entity // saves all the current reverted foragers

	SquadstoReduce int // amount of squads that can be reduced again --> amount of reverted forager squadrons that will be returned to normal forager squadrons next timestep
}

type Nursing_stats struct {
	MaxPollenIntake  float64 // maximum of pollen intake per nurse
	MeanPollenIntake float64 // mean pollen intake per nurse

	NL_ratio float64 // current nurse:larva ratio; probably only to be used as a analytical metric

	NonNurseIHbees int     // amount of not nursing IHbees --> MAB, that do nor forage and do not nurse+
	IHbeeNurses    int     // amount of IHbee nurses
	TotalNurses    int     // total amount of nurses today; including reverted foragers and winterbees
	NurseFraction  float64 // fraction of nurses from total adult population

	RevertedForagers int // amount of reverted foragers at this current tick
	WinterBees       int // amount of winterbees at this current tick
}
