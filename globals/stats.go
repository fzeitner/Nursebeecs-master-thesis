package globals

// PopulationStats contains summarized population numbers for development stages.
//
// PopulationStats is updated at the end of each simulation step.
// Thus, it contains stats of the previous step.
type PopulationStats struct {
	WorkerEggs      int
	WorkerLarvae    int
	WorkerPupae     int
	WorkersInHive   int
	WorkersForagers int

	DroneEggs    int
	DroneLarvae  int
	DronePupae   int
	DronesInHive int

	TotalBrood      int
	TotalAdults     int
	TotalPopulation int

	NumberIHbeeCohorts int // only for debugging, will probably remove this later

	MeanDoseIHBees      float64 // to avoid creating any more ECS entities only for PPP dosage/vulnerability for now, may need to change
	MeanDoseLarvae      float64 // to avoid creating any more ECS entities only for PPP dosage/vulnerability for now, may need to change
	MeanDoseDrones      float64 // to avoid creating any more ECS entities only for PPP dosage/vulnerability for now, may need to change
	MeanDoseDroneLarvae float64 // to avoid creating any more ECS entities only for PPP dosage/vulnerability for now, may need to change
	MeanDoseForager     float64 // doesn´t do anything, just there for debugging

	CumDoseIHBees      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseLarvae      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseForagers    float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDrones      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDroneLarvae float64 // cumulative dose before calculating a mean, used for debugging

	PPPNursebees float64 // variable for debugging and finding out how much PPP is "lost" to nursebees, who are not explicitely modeled
}

// Reset all stats to zero.
func (s *PopulationStats) Reset() {
	s.WorkerEggs = 0
	s.WorkerLarvae = 0
	s.WorkerPupae = 0
	s.WorkersInHive = 0
	s.WorkersForagers = 0

	s.DroneEggs = 0
	s.DroneLarvae = 0
	s.DronePupae = 0
	s.DronesInHive = 0

	s.TotalBrood = 0
	s.TotalAdults = 0
	s.TotalPopulation = 0

	s.MeanDoseIHBees = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseLarvae = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDrones = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDroneLarvae = 0 // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
}

// ConsumptionStats contains statistics on daily consumption.
type ConsumptionStats struct {
	HoneyDaily float64 // Today's honey consumption [mg].
}

// Reset all stats to zero.
func (s *ConsumptionStats) Reset() {
	s.HoneyDaily = 0
}

// ForagingStats contains statistics on foraging per foraging round.
type ForagingStats struct {
	Rounds []ForagingRound

	ContactExp_once   int // for debugging how often foragers get contact exp per foraging day
	ContactExp_repeat int // for debugging how often foragers get contact exp per foraging day
}

// Reset all stats.
func (s *ForagingStats) Reset() {
	s.Rounds = s.Rounds[:0]

	s.ContactExp_once = 0
	s.ContactExp_repeat = 0
}

// ForagingRound contains statistics for a single foraging round.
// Not used as an ECS resource directly!
type ForagingRound struct {
	Lazy      int
	Resting   int
	Searching int
	Recruited int
	Nectar    int
	Pollen    int

	Water int
}
