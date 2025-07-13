package comp

import (
	"github.com/fzeitner/beecs_ecotox/enum/activity"
	"github.com/mlange-42/ark/ecs"
)

// Milage component for forager squadrons.
type Milage struct {
	Today float32 // Milage today [km].
	Total float32 // Milage over total lifetime [km].
}

// Age component for forager squadrons.
type Age struct {
	DayOfBirth int // Date of birth for calculating the age from the current model tick.
}

// Activity component for forager squadrons.
type Activity struct {
	Current       activity.ForagerActivity // Current activity.
	PollenForager bool                     // Whether it is currently foraging for pollen.

	WaterForager bool // Whether it is currently foraging for water.
}

// KnownPatch component for forager squadrons.
type KnownPatch struct {
	Nectar ecs.Entity // Known nectar patch.
	Pollen ecs.Entity // Known pollen patch.

	Water ecs.Entity // Known water patch.

	VisitedthisDay bool
}

// NectarLoad component for forager squadrons.
type NectarLoad struct {
	Energy float64 // Current nectar energy load per individual [kJ]
	Pollen float64 // Current amount of pollen per individual [g]

	PPPLoad float64 // Current amount of PPP in the load [µg]
}

// PPP exposure for forager squadrons.
type PPPExpo struct {
	OralDose    float64 // Current oral dose of this squadron to PPP (foraging) [µg]
	ContactDose float64 // Current contact dose of this squadron to PPP [µg]

	RdmSurvivalContact float64 // Survival chance or "resilience" of the squadron to PPP contact exposure
	RdmSurvivalOral    float64 // Survival chance or "resilience" of the squadron to PPP oral exposure
}
