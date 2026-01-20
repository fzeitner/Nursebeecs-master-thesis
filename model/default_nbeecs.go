package model

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/sys"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// DefaultNbeecs sets up the default nursebeecs model with the standard beecs and nursing related sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func DefaultNbeecs(p params.Params, pn params.ParamsNursebeecs, app *app.App) *app.App {

	// Add parameters and other resources

	app = initializeModelNbeecs(p, pn, app)

	// Initialization
	app.AddSystem(&sys.InitStore{})
	app.AddSystem(&sys.InitCohorts{})
	app.AddSystem(&sys.InitPopulation{})
	app.AddSystem(&sys.InitPatchesList{})
	app.AddSystem(&sys.InitForagingPeriod{})
	app.AddSystem(&sys.InitNursebeecs{}) // inits all the changes necessary for the etox and nursing module

	// Sub-models
	app.AddSystem(&sys.CalcAffNbeecs{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	app.AddSystem(&sys.ReplenishPatches{})

	app.AddSystem(&sys.MortalityCohorts{})
	app.AddSystem(&sys.AgeCohorts{})
	app.AddSystem(&sys.EggLaying{})
	app.AddSystem(&sys.TransitionForagers{})

	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys.NursingNeeds{})    // calculates need of nurses based on population dynamics and nursing metrics from the last step; determines available nurses for broodcare and consumption
	app.AddSystem(&sys.Nbroodcare{})      // new nurse based brood care process (if turned on)

	app.AddSystem(&sys.NewCohorts{})      // here the new cohorts get initialized now
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys.Foraging{})
	app.AddSystem(&sys.MortalityForagers{})

	app.AddSystem(&sys.CountPopulation{})
	app.AddSystem(&sys.NurseConsumption{}) // regulates all consumption of honey and pollen with consideration of nurse bee feeding

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a nbeecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystemsNbeecs(p params.Params, pn params.ParamsNursebeecs, sys []app.System, app *app.App) *app.App {

	app = initializeModelNbeecs(p, pn, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModelNbeecs(p params.Params, pn params.ParamsNursebeecs, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(&a.World)
	pn.Apply(&a.World)

	factory := globals.NewForagerFactory(&a.World)
	ecs.AddResource(&a.World, &factory)

	stats := globals.PopulationStats{}
	ecs.AddResource(&a.World, &stats)

	consumptionStats := globals.ConsumptionStats{}
	ecs.AddResource(&a.World, &consumptionStats)

	foragingStats := globals.ForagingStats{}
	ecs.AddResource(&a.World, &foragingStats)

	nurseStats := globals.NursingStats{}
	ecs.AddResource(&a.World, &nurseStats)

	nurseGlobals := globals.NursingGlobals{}
	ecs.AddResource(&a.World, &nurseGlobals)

	return a
}
