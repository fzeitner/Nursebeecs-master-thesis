package model_etox

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/globals_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/params_etox"
	"github.com/fzeitner/Nursebeecs-master-thesis/sys"
	"github.com/fzeitner/Nursebeecs-master-thesis/sys_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default beecs_ecotox model with the standard sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func Default(p params.Params, pe params_etox.Params_etox, app *app.App) *app.App {

	// Add parameters and other resources

	app = initializeModel(p, pe, app)

	// Initialization

	app.AddSystem(&sys.InitStore{})
	app.AddSystem(&sys.InitCohorts{})
	app.AddSystem(&sys.InitPopulation{})
	app.AddSystem(&sys.InitPatchesList{})
	app.AddSystem(&sys.InitForagingPeriod{})
	app.AddSystem(&sys_etox.Init_etox{}) // inits all the changes necessary for the etox module

	// Sub-models
	app.AddSystem(&sys.CalcAff{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	//app.AddSystem(&sys_etox.CalcWaterForagingPeriod{}) // might as well be disabled atm because no function actually uses water data as of yet, because water foraging seems irrelevant/untested in netlogo as well
	app.AddSystem(&sys.ReplenishPatches{})    // same old function as in beecs
	app.AddSystem(&sys_etox.PPPApplication{}) // introduced PPP exposure at patches

	app.AddSystem(&sys.MortalityCohorts{})           // same old mortality function now again
	app.AddSystem(&sys_etox.MortalityCohorts_etox{}) // introduced ETOXMortality as an additional process for all cohorts
	app.AddSystem(&sys.AgeCohorts{})
	app.AddSystem(&sys.EggLaying{})          // no counting before EggLaying, therefore we can just let it run here after ageing in beecs. Necessary to first age to free up space for new eggs. Therefore has to happen after Mortaliy procs too which have to happen before ageing
	app.AddSystem(&sys.TransitionForagers{}) // now only counts how many foragers are going to be transitioned and empties the IHbeecohort but does not initialize anything to resemble original BEEHAVE more closely

	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys.BroodCare{})       // Moved after the first countingproc to resemble the original model further, as counting twice is inevitable because of ETOXmortality processes.

	app.AddSystem(&sys.NewCohorts{})
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys_etox.Foraging_etox{})          // introduced the uptake of PPP into foragers and the hive through contaminated honey/pollen, very tedious to decouple from normal foraging process
	app.AddSystem(&sys.MortalityForagers{})           // now once again exactly the same as in baseline BEEHAVE, decoupled from etox mortality
	app.AddSystem(&sys_etox.MortalityForagers_etox{}) // introduced ETOXMortality as an additional process for foragers and put after Foraging, because same in BEEHAVE

	app.AddSystem(&sys.CountPopulation{}) // necessary here because of food comsumption in the next steps
	app.AddSystem(&sys.PollenConsumption{})
	app.AddSystem(&sys.HoneyConsumption{})
	app.AddSystem(&sys_etox.EtoxStorages{}) // regulates in-hive exposition and fate of PPP and the newly introduced honey compartiments

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a beecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystems(p params.Params, pe params_etox.Params_etox, sys []app.System, app *app.App) *app.App {

	app = initializeModel(p, pe, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModel(p params.Params, pe params_etox.Params_etox, a *app.App) *app.App {
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
