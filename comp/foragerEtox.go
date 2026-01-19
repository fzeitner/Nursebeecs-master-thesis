package comp

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/enum/activity"
	"github.com/mlange-42/ark/ecs"
)

// NectarLoad component for forager squadrons.
type EtoxLoad struct {
	PPPLoad float64 // Current amount of PPP in the load [µg]

	EnergyUsed float64 // amount of energy used this day (from foraging/scouting only); only used for debugging
}

// PPP exposure for forager squadrons.
type PPPExpo struct {
	OralDose    float64 // Current daily oral dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]
	ContactDose float64 // Current daily contact dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]

	RdmSurvivalContact float64 // Survival chance or "resilience" of the squadron to PPP contact exposure
	RdmSurvivalOral    float64 // Survival chance or "resilience" of the squadron to PPP oral exposure
}

type KnownPatchEtox struct {
	Nectar ecs.Entity // Known nectar patch.
	Pollen ecs.Entity // Known pollen patch.

	Water ecs.Entity // Known water patch.

	VisitedthisDay bool // addition for debugging patch visits
}

// Activity component for forager squadrons.
type ActivityEtox struct {
	Current       activity.ForagerActivity // Current activity.
	PollenForager bool                     // Whether it is currently foraging for pollen.

	Winterbee bool // switches to have foragers counted as winterbees --> they can forage and nurse at the same time. This will need to be adressed in the future. for now necessary for nursing
	Reverted  bool // switches to have foragers counted as reverted foragers that can nurse again but do not forage. for now necessary for nursing

	WaterForager bool // Whether it is currently foraging for water.
}
