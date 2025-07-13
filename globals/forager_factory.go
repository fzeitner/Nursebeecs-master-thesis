package globals

import (
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/enum/activity"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// ForagerFactory is a helper resource for creating forager squadron entities.
type ForagerFactory struct {
	builder *ecs.Map6[comp.Activity, comp.KnownPatch, comp.Age, comp.Milage, comp.NectarLoad, comp.PPPExpo]
	source  rand.Source
}

// NewForagerFactory creates a new ForagerFactory
func NewForagerFactory(world *ecs.World) ForagerFactory {
	return ForagerFactory{
		builder: ecs.NewMap6[comp.Activity, comp.KnownPatch, comp.Age, comp.Milage, comp.NectarLoad, comp.PPPExpo](world),
		source:  rand.New(ecs.GetResource[resource.Rand](world)),
	}
}

// CreateSquadrons creates the given number of squadrons with the given day of birth
// (usually the current model tick).
func (f *ForagerFactory) CreateSquadrons(count int, dayOfBirth int) {
	rng := rand.New(f.source)

	f.builder.NewBatchFn(count, func(entity ecs.Entity, act *comp.Activity, _ *comp.KnownPatch, age *comp.Age, _ *comp.Milage, _ *comp.NectarLoad, ppp *comp.PPPExpo) {
		age.DayOfBirth = dayOfBirth
		act.Current = activity.Resting

		ppp.RdmSurvivalContact = rng.Float64()
		ppp.RdmSurvivalOral = rng.Float64()
	})
}

// CreateInitialSquadrons creates the given number of squadrons with random day of birth and milage as given by the ranges.
//
// Used to create initial foragers.
func (f *ForagerFactory) CreateInitialSquadrons(count int, minDayOfBirth, maxDayOfBirth int, minMilage, maxMilage float32, rnd rand.Source) {
	rng := rand.New(rnd)
	f.builder.NewBatchFn(count, func(entity ecs.Entity, act *comp.Activity, _ *comp.KnownPatch, age *comp.Age, milage *comp.Milage, _ *comp.NectarLoad, ppp *comp.PPPExpo) {
		age.DayOfBirth = rng.IntN(maxDayOfBirth-minDayOfBirth) + minDayOfBirth
		milage.Total = rng.Float32()*(maxMilage-minMilage) + minMilage
		act.Current = activity.Resting

		ppp.RdmSurvivalContact = rng.Float64()
		ppp.RdmSurvivalOral = rng.Float64()
	})
}
