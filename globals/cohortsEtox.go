package globals

// Larvae contains worker and drone larvae age cohorts.
type LarvaeEtox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}

// InHive contains in-hive worker and drone age cohorts.
type InHiveEtox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}
