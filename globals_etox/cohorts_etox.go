package globals_etox

// Larvae contains worker and drone larvae age cohorts.
type Larvae_etox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}

// InHive contains in-hive worker and drone age cohorts.
type InHive_etox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort. Only counts the actual exposure through eating; no decay processes involved. Gets reset every timestep unless GUTS is turned on.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.

	WorkerCohortITthreshold []float64 // IT threshold per Cohort. May need to be reworked as it might make sense to have multiple IT thresholds per cohort as cohorts can become quite big
	WorkerCohortC_i         []float64 // mean internal concentration of PPP per bee per cohort; different to WokerCohortDose, as this one gets updated through GUTS
}
