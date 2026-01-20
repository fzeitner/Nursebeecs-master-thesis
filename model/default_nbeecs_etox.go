package model

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/fzeitner/Nursebeecs-master-thesis/sys"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// DefaultNbeecsEtox sets up the default nursebeecs_ecotox model with the standard sub-models.
// This includes all nursebeecs submodels as well as all _ecotox submodels.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func DefaultNbeecsEtox(p params.Params, pe params.ParamsEtox, pn params.ParamsNursebeecs, app *app.App) *app.App {

	// Add parameters and other resources
	app = initializeModelNbeecsEtox(p, pe, pn, app)

	// Initialization
	app.AddSystem(&sys.InitStore{})
	app.AddSystem(&sys.InitCohorts{})
	app.AddSystem(&sys.InitPopulation{})
	app.AddSystem(&sys.InitPatchesList{})
	app.AddSystem(&sys.InitForagingPeriod{})
	app.AddSystem(&sys.InitEtoxNursebeecs{}) // inits all the changes necessary for the nursebeecs_ecotox submodels

	// Sub-models
	app.AddSystem(&sys.CalcAffNbeecs{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	app.AddSystem(&sys.ReplenishPatches{}) // unchanged to beecs
	app.AddSystem(&sys.PPPApplication{})   // introduced PPP exposure at patches

	app.AddSystem(&sys.MortalityCohorts{})     // unchanged to beecs
	app.AddSystem(&sys.MortalityCohortsEtox{}) // introduced ETOXMortality as an additional process for all cohorts
	app.AddSystem(&sys.AgeCohorts{})           // unchanged to beecs
	app.AddSystem(&sys.EggLaying{})            // unchanged to beecs
	app.AddSystem(&sys.TransitionForagers{})   // unchanged to beecs

	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys.NursingNeeds{})    // calculates need of nurses based on population dynamics and nursing metrics from the last step; determines available nurses for broodcare and consumption
	app.AddSystem(&sys.Nbroodcare{})      // new nurse based brood care process (if turned on)

	app.AddSystem(&sys.NewCohorts{})      // unchanged to beecs
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys.ForagingEtox{})          // introduced the uptake of PPP into foragers and the hive through contaminated honey/pollen, very tedious to decouple from normal foraging process
	app.AddSystem(&sys.MortalityForagers{})     // unchanged to beecs
	app.AddSystem(&sys.MortalityForagersEtox{}) // introduced ETOXMortality as an additional process for foragers after normal foraging mortality, analogous to BEEHAVE_ecotox

	app.AddSystem(&sys.CountPopulation{})      // necessary here because of food comsumption in the next steps
	app.AddSystem(&sys.NurseConsumptionEtox{}) // regulates all consumption of honey and pollen with consideration of nurse bee feeding
	app.AddSystem(&sys.EtoxStoragesNbeecs{})   // regulates in-hive exposure and fate of PPP for nursebeecs

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a beecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystemsNbeecsEtox(p params.Params, pe params.ParamsEtox, pn params.ParamsNursebeecs, sys []app.System, app *app.App) *app.App {

	app = initializeModelNbeecsEtox(p, pe, pn, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModelNbeecsEtox(p params.Params, pe params.ParamsEtox, pn params.ParamsNursebeecs, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(&a.World)
	pe.Apply(&a.World)
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
