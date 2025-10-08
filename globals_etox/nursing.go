package globals_etox

// this will keep all the necessary globals to calculate and track nursing dynamics

type Nursing_globals struct {
	Nurse_need_today     int // theoretical need of nurses of today; PROBABLY NOT USABLE ANYMORE BECAUSE OF LACK OF DATA
	Nurse_need_yesterday int // theoretical need of nurses of last day; PROBABLY NOT USABLE ANYMORE BECAUSE OF LACK OF DATA

	SuffNurses bool // were there sufficient nurses last consumption proc?

	NurseAgeMax   int     // current upper bound age of nurses
	NL_ratio      float64 // current nurse:larva ratio; probably only to be used as a analytical metric
	NurseWorkLoad float64 // ratio of protein nurses need to provide to the amount they on average can provide as a metric for their workload

	Total_honey  float64 // amount of total honey to be taken in by nurses on top of their own intake this timestep
	Total_pollen float64 // amount of total pollen to be taken in by nurses on top of their own intake this timestep
}

type Nursing_stats struct {
	MaxPollenIntake  float64 // maximum of pollen intake per nurse
	MeanPollenIntake float64 // mean pollen intake per nurse
	TotalNurses      int     // amount of nurses today
}
