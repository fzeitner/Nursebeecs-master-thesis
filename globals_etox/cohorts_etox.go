package globals_etox

// Larvae contains worker and drone larvae age cohorts.
type Larvae_etox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}

// InHive contains in-hive worker and drone age cohorts.
type InHive_etox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}
