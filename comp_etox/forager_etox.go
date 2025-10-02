package comp_etox

import (
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
	"github.com/mlange-42/ark/ecs"
)

// NectarLoad component for forager squadrons.
type EtoxLoad struct {
	PPPLoad float64 // Current amount of PPP in the load [µg]

	EnergyUsed float64 // amount of energy used this day (from foraging/scouting only)
}

// PPP exposure for forager squadrons.
type PPPExpo struct {
	OralDose    float64 // Current daily oral dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]
	ContactDose float64 // Current daily contact dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]

	RdmSurvivalContact float64 // Survival chance or "resilience" of the squadron to PPP contact exposure
	RdmSurvivalOral    float64 // Survival chance or "resilience" of the squadron to PPP oral exposure

	C_i           float64 // Current effective exposure per forager squadron to PPP; uses BeeGUTS rate constants to calculate an equilibrium concentration based on honey stomach
	RmdSurvivalIT float64 // log-logistically drawn survival threshold for ITmodel
}

type KnownPatch_etox struct {
	Nectar ecs.Entity // Known nectar patch.
	Pollen ecs.Entity // Known pollen patch.

	Water ecs.Entity // Known water patch.

	VisitedthisDay bool // addition for debugging patch visits
}

// Activity component for forager squadrons.
type Activity_etox struct {
	Current       activity.ForagerActivity // Current activity.
	PollenForager bool                     // Whether it is currently foraging for pollen.

	WaterForager bool // Whether it is currently foraging for water.
}
