package comp_etox

import (
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
	"github.com/mlange-42/ark/ecs"
)

// NectarLoad component for forager squadrons.
type PPPLoad struct {
	PPPLoad float64 // Current amount of PPP in the load [µg]
}

// PPP exposure for forager squadrons.
type PPPExpo struct {
	OralDose    float64 // Current oral dose of this squadron to PPP (foraging) [µg]
	ContactDose float64 // Current contact dose of this squadron to PPP [µg]

	RdmSurvivalContact float64 // Survival chance or "resilience" of the squadron to PPP contact exposure
	RdmSurvivalOral    float64 // Survival chance or "resilience" of the squadron to PPP oral exposure
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
