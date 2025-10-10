package globals_etox

// this will keep all the necessary globals to calculate and track nursing dynamics

type Nursing_globals struct {
	SuffNurses bool // were there sufficient nurses last consumption proc?

	NurseAgeMax   int     // current upper bound age of nurses
	NurseWorkLoad float64 // ratio of protein nurses need to provide to the amount they on average can provide as a metric for their workload

	Total_honey  float64 // amount of total honey to be taken in by nurses on top of their own intake this timestep
	Total_pollen float64 // amount of total pollen to be taken in by nurses on top of their own intake this timestep

	WLHoney  float64 // amount of direct honey worker larvae take in this timestep
	WLPollen float64 // amount of direct pollen worker larvae take in this timestep
	DLHoney  float64 // amount of direct honey drone larvae take in this timestep
	DLPollen float64 // amount of direct pollen drone larvae take in this timestep

	WorkerPriming float64 // amount of pollen needed for young workers to prime their HPG this timestep
}

type Nursing_stats struct {
	MaxPollenIntake  float64 // maximum of pollen intake per nurse
	MeanPollenIntake float64 // mean pollen intake per nurse
	NL_ratio         float64 // current nurse:larva ratio; probably only to be used as a analytical metric
	TotalNurses      int     // amount of nurses today
}
