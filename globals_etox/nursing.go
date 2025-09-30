package globals_etox

// this will keep all the necessary globals to calculate and track nursing dynamics

type nurse_need struct {
	Nurse_need_today     int // theoretical need of nurses of today
	Nurse_need_yesterday int // theoretical need of nurses of last day

	Nurse_age_max int     // current maximum age of nurses
	N_L_ratio     float64 // current nurse:larva ratio

	Total_honey  float64 // amount of total honey to be taken in by nurses on top of their own intake
	Total_pollen float64 // amount of total pollen to be taken in by nurses on top of their own intake
}
