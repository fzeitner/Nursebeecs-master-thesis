package model

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/sys"
	"github.com/fzeitner/beecs_masterthesis/sys_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default beecs model with the standard sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func Default_nbeecs(p params.Params, pe params_etox.Params_etox, app *app.App) *app.App {

	// Add parameters and other resources

	app = initializeModel_nbeecs(p, pe, app)

	// Initialization
	app.AddSystem(&sys.InitStore{})
	app.AddSystem(&sys.InitCohorts{})
	app.AddSystem(&sys.InitPopulation{})
	app.AddSystem(&sys.InitPatchesList{})
	app.AddSystem(&sys.InitForagingPeriod{})
	app.AddSystem(&sys_etox.Init_etox{}) // inits all the changes necessary for the etox and nursing module

	// Sub-models
	app.AddSystem(&sys.CalcAff{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	app.AddSystem(&sys.ReplenishPatches{})

	app.AddSystem(&sys.MortalityCohorts{})
	app.AddSystem(&sys.AgeCohorts{})
	app.AddSystem(&sys.EggLaying{})
	app.AddSystem(&sys.TransitionForagers{})

	app.AddSystem(&sys.CountPopulation{})   // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys_etox.NursingNeeds{}) // calculates need of nurses based on population dynamics and nursing metrics from the last step; determines available nurses for broodcare and consumption
	app.AddSystem(&sys_etox.Nbroodcare{})   // new nurse based brood care process (if turned on)

	app.AddSystem(&sys.NewCohorts{})      // here the new cohorts get initialized now
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys.Foraging{}) // also initializes foragers now
	app.AddSystem(&sys.MortalityForagers{})

	app.AddSystem(&sys.CountPopulation{})
	app.AddSystem(&sys_etox.NurseConsumption{})

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a nbeecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystems_nbeecs(p params.Params, pe params_etox.Params_etox, sys []app.System, app *app.App) *app.App {

	app = initializeModel_nbeecs(p, pe, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModel_nbeecs(p params.Params, pe params_etox.Params_etox, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(&a.World)
	pe.Apply(&a.World)

	factory := globals.NewForagerFactory(&a.World)
	ecs.AddResource(&a.World, &factory)

	stats := globals.PopulationStats{}
	ecs.AddResource(&a.World, &stats)

	consumptionStats := globals.ConsumptionStats{}
	ecs.AddResource(&a.World, &consumptionStats)

	foragingStats := globals.ForagingStats{}
	ecs.AddResource(&a.World, &foragingStats)

	nurseStats := globals_etox.Nursing_stats{}
	ecs.AddResource(&a.World, &nurseStats)

	nurseGlobals := globals_etox.Nursing_globals{}
	ecs.AddResource(&a.World, &nurseGlobals)

	return a
}
