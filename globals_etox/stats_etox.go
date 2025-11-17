package globals_etox

// PopulationStats contains summarized population numbers for development stages.
//
// PopulationStats is updated at the end of each simulation step.
// Thus, it contains stats of the previous step.
type PopulationStats_etox struct {
	NumberIHbeeCohorts int // only for debugging, will probably remove this later

	MeanDoseIHBees      float64 // doesn´t do anything, just there for debugging
	MeanDoseLarvae      float64 // doesn´t do anything, just there for debugging
	MeanDoseDrones      float64 // doesn´t do anything, just there for debugging
	MeanDoseDroneLarvae float64 // doesn´t do anything, just there for debugging
	MeanDoseForager     float64 // doesn´t do anything, just there for debugging

	CumDoseIHBees      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseLarvae      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseForagers    float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDrones      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDroneLarvae float64 // cumulative dose before calculating a mean, used for debugging

	PPPNursebees float64 // variable for debugging and finding out how much PPP is "lost" to nursebees, who are not explicitely modeled

	CumDoseNurses  float64
	MeanDoseNurses float64
}

// Reset all stats to zero.
func (s *PopulationStats_etox) Reset() {
	s.MeanDoseIHBees = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseLarvae = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDrones = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDroneLarvae = 0 // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses

	s.CumDoseIHBees = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseLarvae = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseDrones = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseDroneLarvae = 0 // cumulative dose before calculating a mean, used for debugging

	s.PPPNursebees = 0

	s.CumDoseNurses = 0
	s.MeanDoseNurses = 0

}

// ForagingRound contains statistics for a single foraging round.
// Not used as an ECS resource directly!
type ForagingRound_etox struct {
	Lazy      int
	Resting   int
	Searching int
	Recruited int
	Nectar    int
	Pollen    int

	Water int
}

// ForagingStats contains statistics on foraging per foraging round.
type ForagingStats_etox struct {
	Rounds []ForagingRound_etox

	ContactExp_once        int     // for debugging how often foragers get contact exp per foraging day
	ContactExp_repeat      int     // for debugging how often foragers get contact exp per foraging day
	Prob                   float64 // debugging global
	SumDur                 float64 // debugging global
	Pollensuccess          float64 // debugging global
	Collectionflightstotal int     // debugging global
	Foragerdied            int     // debugging global
	TotalSearches          int     //debugging global
	ForagerDiedLifespan    int     //debugging global

	MaxEnergyExpenditure  float64 // for tracking the energy use of foragers per foraging day
	MeanEnergyExpenditure float64 // for tracking the energy use of foragers per foraging day
}

// Reset all stats.
func (s *ForagingStats_etox) Reset() {
	s.Rounds = s.Rounds[:0]

	s.ContactExp_once = 0
	s.ContactExp_repeat = 0
	s.Prob = 0.
	s.SumDur = 0.
	s.Pollensuccess = 0.
	s.Collectionflightstotal = 0
	s.Foragerdied = 0
	s.TotalSearches = 0
	s.ForagerDiedLifespan = 0

	s.MaxEnergyExpenditure = 0.
	s.MeanEnergyExpenditure = 0.
}
