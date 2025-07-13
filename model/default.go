package model

import (
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/fzeitner/beecs_ecotox/sys"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default beecs model with the standard sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func Default(p params.Params, app *app.App) *app.App {

	// Add parameters and other resources

	app = initializeModel(p, app)

	// Initialization
	//app.AddSystem(&sys.TimeControl{})         // well I tried I guess; lets ask Martin if there is a way to do this that does not involve to just let every subsystem pass if time.Tick == 0
	app.AddSystem(&sys.InitStore{})
	app.AddSystem(&sys.InitCohorts{})
	app.AddSystem(&sys.InitPopulation{})
	app.AddSystem(&sys.InitPatchesList{})
	app.AddSystem(&sys.InitForagingPeriod{})

	// Sub-models

	app.AddSystem(&sys.CalcAff{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	app.AddSystem(&sys.ReplenishPatches{}) // introduced PPP exposure at patches

	app.AddSystem(&sys.MortalityCohorts{}) // introduced ETOXMortality as an additional process for all cohorts
	app.AddSystem(&sys.AgeCohorts{})
	app.AddSystem(&sys.EggLaying{})       // no counting before this one because this happens before counting in the timestep in orig. model and has to happen after ageing in beecs
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys.BroodCare{})       // Moved after the first countingproc to resemble the original model further, as counting twice is inevitable because of ETOXmortality processes.

	app.AddSystem(&sys.TransitionForagers{})
	app.AddSystem(&sys.Foraging{})          // introduced the uptake of PPP into foragers and the hive through contaminated honey/pollen
	app.AddSystem(&sys.MortalityForagers{}) // introduced ETOXMortality as an additional process for foragers and put after Foraging, because same in BEEHAVE

	app.AddSystem(&sys.CountPopulation{}) // necessary here because of food comsumption in the next steps
	app.AddSystem(&sys.PollenConsumption{})
	app.AddSystem(&sys.HoneyConsumption{})
	app.AddSystem(&sys.EtoxStorages{}) // regulates in-hive exposition and fate of PPP and the newly introduced honey compartiments

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a beecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystems(p params.Params, sys []app.System, app *app.App) *app.App {

	app = initializeModel(p, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModel(p params.Params, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(&a.World)

	factory := globals.NewForagerFactory(&a.World)
	ecs.AddResource(&a.World, &factory)

	stats := globals.PopulationStats{}
	ecs.AddResource(&a.World, &stats)

	consumptionStats := globals.ConsumptionStats{}
	ecs.AddResource(&a.World, &consumptionStats)

	foragingStats := globals.ForagingStats{}
	ecs.AddResource(&a.World, &foragingStats)

	return a
}
