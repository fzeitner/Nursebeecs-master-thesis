package model_etox

import (
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/sys"
	"github.com/fzeitner/beecs_masterthesis/sys_etox"
	"github.com/mlange-42/ark-tools/app"
)

// Default sets up the default beecs model with the standard sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func Default_nbeecs(p params.Params, pe params_etox.Params_etox, app *app.App) *app.App {

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
	app.AddSystem(&sys_etox.CalcAffNbeecs{})
	app.AddSystem(&sys.CalcForagingPeriod{})
	//app.AddSystem(&sys_etox.CalcWaterForagingPeriod{}) // might as well be disabled atm because no function actually uses water data as of yet, because water foraging seems irrelevant/untested in netlogo as well
	app.AddSystem(&sys.ReplenishPatches{})    // same old function as in beecs
	app.AddSystem(&sys_etox.PPPApplication{}) // introduced PPP exposure at patches

	app.AddSystem(&sys.MortalityCohorts{})           // same old mortality function now again
	app.AddSystem(&sys_etox.MortalityCohorts_etox{}) // introduced ETOXMortality as an additional process for all cohorts
	app.AddSystem(&sys.AgeCohorts{})                 // introduced takeover of dosage/Ci values between cohorts for GUTS only
	app.AddSystem(&sys.EggLaying{})                  // no counting before EggLaying, therefore we can just let it run here after ageing in beecs. Necessary to first age to free up space for new eggs. Therefore has to happen after Mortaliy procs too which have to happen before ageing
	app.AddSystem(&sys.TransitionForagers{})         // now only counts how many foragers are going to be transitioned and empties the IHbeecohort but does not initialize anything to resemble original BEEHAVE more closely

	app.AddSystem(&sys.CountPopulation{})   // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys_etox.NursingNeeds{}) // calculates need of nurses based on population dynamics and nursing metrics from the last step; determines available nurses for broodcare and consumption
	app.AddSystem(&sys_etox.Nbroodcare{})   // new nurse based brood care process (if turned on)

	app.AddSystem(&sys.NewCohorts{})
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys_etox.Foraging_etox{})          // introduced the uptake of PPP into foragers and the hive through contaminated honey/pollen, very tedious to decouple from normal foraging process
	app.AddSystem(&sys.MortalityForagers{})           // now once again exactly the same as in baseline BEEHAVE, decoupled from etox mortality
	app.AddSystem(&sys_etox.MortalityForagers_etox{}) // introduced ETOXMortality as an additional process for foragers and put after Foraging, because same in BEEHAVE

	app.AddSystem(&sys.CountPopulation{})       // necessary here because of food comsumption in the next steps
	app.AddSystem(&sys_etox.NurseConsumption{}) //
	app.AddSystem(&sys_etox.EtoxStorages{})     // regulates in-hive exposition and fate of PPP and the newly introduced honey compartiments

	app.AddSystem(&sys.FixedTermination{})

	return app
}
