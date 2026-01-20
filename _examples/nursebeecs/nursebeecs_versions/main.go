package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/Nursebeecs-master-thesis/model"
	"github.com/fzeitner/Nursebeecs-master-thesis/obs"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	// define default parameters; uses Berlin2000 as the default weather scenario. To change see beecs_examples.
	p := params.Default()
	start := time.Now()

	// define DefaultNursebeecs parameters
	pn := params.DefaultNursebeecs()

	// these parameters can be adjusted and look like this:
	pn.ConsumptionRework = params.ConsumptionRework{
		HoneyAdultWorker:  11., // mg/day <- may need changing; Brodschneider&Crailsheim 2010 quote Barker & Lehner 1974 for 4mg of sugar per day for survival = ca. 5.1mg honey; old BEEHAVE val is 11 (Rortais Winterbees)
		PollenAdultWorker: 1.5, // mg/day <- old value for 14 day old bees from Rortais et al. 2005; should fit as a baseline for now; maybe adjust down the line

		MaxPollenNurse: 6.5, // + 1.5 per adult = 8 mg/day; this should be a field realistic total for a normal peak; Crailsheim reported up to 8 as a max, 12 as the highes statistical 95% bound under controlled conditions. 12 is cited as a maximum in BeeREX model and comes from Rortais et al. 2005 citing Crailsheim et al. 1992;
		MaxHoneyNurse:  60., // mg/day; estimate based on BeeREX citing Rortais et al. 2005; this is however most likely not accurately interpreted by BeeREX authors. This value is probably unknown and 60 is probably too high. This should not really matter though, as pollen intake will regulate amount of nurses and honey intake will be an emergent property

		HoneyAdultDrone:  10., // mg/day; taken from BEEHAVE, though the origin of this value is very unclear; this might be another placeholder
		PollenAdultDrone: 2.,  // mg/day; taken from BEEHAVE, already just a rough estimate; there appears to be no clear value anywhere, though it is known that they have an increased need for the first 9 days to reach maturity

		HoneyWorkerLarva:  make([]float64, 6), // gets initialized in sys.initEtox for now because I do not know how else to do this
		PollenWorkerLarva: make([]float64, 6), // gets initialized in sys.initEtox for now because I do not know how else to do this
		HWLtotal:          65.4,               // mg over a total of 6 days --> old BEEHAVE value; an increase to 75.5 might make sense, value taken from Rortais et al. 2005
		PWLtotal:          100.,               // mg over a total of 6 days; this is a lowered estimate as opposed to original BEEHAVE (used 142 mg), because some of the budged gets shifted to the first few days of adult development
		PFPworker:         42,                 // mg over the first 4 days of life; this gets taken in by nurses if possible, if not the bees eat it themselves. 42 mg makes the pollen budget turn out exactly the same as before
		HoneyDirect:       0.05,               // unknown, therefore same estimate as pollen value below for now
		PollenDirect:      0.05,               // 5% of pollen get taken in directly, estimated by Hrassnigg & Crailsheim (2005); for now assumes the same value for worker and drone larvae from day 3 onwards and thus has a bigger effect on drones that take 1 day longer to pupation
		// this should be okay though as many studies found that drone food has higher residues than worker food, which in turn is higher than royal jelly of queen larvae (example: Wueppenhorst et al. 2024). This should be a somewhat realistic and defensible assumption.

		HoneyDroneLarva:  make([]float64, 7), // gets initialized in sys.initEtox for now because I do not know how else to do this
		PollenDroneLarva: make([]float64, 7), // gets initialized in sys.initEtox for now because I do not know how else to do this
		HDLtotal:         124.9,              // mg over a total of 7 days; the old BEEHAVE value, taken from Rortais et al. 2005
		PDLtotal:         250.,               // mg over a total of 7 days; there is no proper estimate, this is lowered as opposed to original BEEHAVE (used 350 mg), because that is most likely too high. It is estimated that drones larvae weigh 1.8 - 2.6 more than workers (Hrassnigg and Crailsheim 2005)
		PFPdrone:         100.,               // mg over the first 9 days of adult life; this gets taken in by nurses, as drones do not really eat any pollen by themselves. 100 mg makes the pollen budget turn out exactly the same as before
		// Hrassnigg and Crailsheim (2005) use the same values for carbohydrates as Rortais for both larvae, but use a higher pollen budged than I estimated here. I could also simply adopt their budgets, but that would not be completely biologically accurate for modeling dynamics,
		// because both worker and drones have an increased need of pollen after emerging (worker for 3-5 days, drones for ca. 8-10 days) to reach complete maturity. The authors mention this as well, it is hard to estimate how much of the budget is allocated to priming as adults.

		DynamicProteinNursing: false,               // determines if nursing capability shall be dynamically adjusted based on worker age (there is evidence of a peaking in capability by age 6-10 roughly)
		Nursingcapabiliies:    make([]float64, 51), // array to save the dynamic capabilities of nurse bees depending on their age
	}
	// the ConsumptionRework parameters should not be adjusted on a whim; the parameter values used here are a product of my masther thesis and
	// explained in detail there. The parametersNursebeecs.go file in the params-folder gives some more explanation for these parameter values.

	pn.NursingRework = params.NursingRework{
		//MinWLRatio:            2,                                            // might not be used for now; see Eischen et al. 1982, 1983, 1984; placeholder for now but it seems 2:1 Worker:Larva gives a good efficiency baseline for rearing, where adult longevity is somewhat as expected
		NurseAgeCeiling:        13,                                           // default age at which nurses stop working as nurses, unless model dynamics increase this
		BroodCannibalismChance: []float64{0.23, 0.3, 0.58, 0.06, 0., 0., 0.}, // based on data from Schmickl&Crailsheim (2001, 2002) and HoPoMo model quoting these studies
		NurseWorkLoadTH:        1.5,                                          // equals 11.25 mg of pollen per day, this should be a reasonable maximum intake for nurse bees (Rortais et al. 2005, Crailsheim et al. 1992)
		MinimumTH:              1.0,                                          // 1.0 equals per calculation of NurseWorkload a reasonable mean intake of nurse bees, because NurseWorkload is designed to represent exactly this

		NewBroodCare:           true,
		Nursebeecsv0:           false,
		Nursebeecsv1:           true,
		ForesightedCannibalism: false,

		HGEffects:    false,
		HGFoodIntake: false,
	}
	// the NursingRework parameter subset can be adjusted slightly; specifically the boolean switches beginning with NewBroodCare above can be switched to true to enable
	// specific simulation. At default everything but NewBroodCare and Nursebeecsv1 is turned off; NewBroodCare enables any new brood care mechanisms that are not the BEEHAVE/beecs default brood care
	// and Nursebeecsv1 activates the final model version that was created and discussed in depth in my master thesis. Note that NewBroodCare needs to be turned on to access any of the other switches below it.

	// first, lets deactivate all booleans in NursingRework that switch on the various versions:
	pn.NursingRework.NewBroodCare = false
	pn.NursingRework.Nursebeecsv1 = false

	// running this model version does not make a lot of sense, as this activates the new consumption of nursebeecs but does not consider any compatibility with the other subsystems.
	// the only reason to run this is to debug some specific subsystems. Lets run it nonetheless:
	// run nursebeecs model; note that baseline- and nursebeecs-parameters are needed here, because the model versions require them to function at all.
	filename := "oldBC"
	for i := 0; i < 100; i++ {
		runNursebeecs(app, i, filename, &p, &pn)
	}
	dur := time.Since(start)
	fmt.Println(dur)

	// now, lets activate at least the NewBroodCare; this is equivalent to the ConsumptionRework model version that was created and explained in my master thesis
	pn.NursingRework.NewBroodCare = true
	filename = "ConsumptionRework"
	for i := 0; i < 100; i++ {
		runNursebeecs(app, i, filename, &p, &pn)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// now, we can acticate different model versions, if we wanted to. Nursebeecsv0 activates one of the first somewhat stable versions that were created during
	// coupling of ProteinFactorNurses to the population dynamics that resulted in rather unnatural oscillations
	pn.NursingRework.Nursebeecsv0 = true
	filename = "Nursebeecsv0"
	for i := 0; i < 100; i++ {
		runNursebeecs(app, i, filename, &p, &pn)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// we could activate the foresighted cannibalism model version now as well, that was created and discussed during the coupling process of ProteinFactorNurses to
	// broodcare by activating ForesightedCannibalism on top of Nursebeecsv0. But I think you get the idea by now, therefore lets go to the final Nursebeecs model version now:
	pn.NursingRework.Nursebeecsv0 = false
	pn.NursingRework.Nursebeecsv1 = true
	filename = "Nursebeecs"
	for i := 0; i < 100; i++ {
		runNursebeecs(app, i, filename, &p, &pn)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// the last two switches activate the simulated sublethal effects on brood care that were one of the main topics of my master thesis. To consider simulations with these the
	// nursebeecs_ecotox model needs to be run. Therefore, lets define the ecotox parameters and run the model:
	pe := params.DefaultEtox()
	pn.NursingRework.HGEffects = true     // this activates HGEffects by lowering the ceiling of ProteinFactorNurses and forcing a fraction of brood to be killed every timestep when the PPP threshold gets exceeded
	pn.NursingRework.HGFoodIntake = false // activating this as well would deactivate the ceiling effect and activate the reduction in jelly provision by reducing the pollen intake of nurse bees
	filename = "sublethal1"
	for i := 0; i < 100; i++ {
		runNursebeecsEtox(app, i, filename, &p, &pe, &pn)
	}
	dur = time.Since(start)
	fmt.Println(dur)

	// note that the effect thresholds for PPP as well as the effects on the respective parameters are defined in the PPPToxicity parameter set of the etox parameters:
	pe.PPPToxicity.HGthreshold = []float64{0.000727 / 12.78, 0.000727 / 12.78 * 10, 0.000727 / 12.78 * 100}
	pe.PPPToxicity.ProteinFactorNurseExposed = []float64{0.65, 0.65, 0.65}
	pe.PPPToxicity.MaxPollenRed = []float64{0.3, 0.3, 0.3}
	// these slices leave place for three effect thresholds in the HGthreshold variable that connects to three effect levels in the PFNexposed or MaxPollenRed slices
	// all of this is experimental and based on Schott et al. (2021). It should be reevaluated and implemented properly; this was only a first try at implementing sublethal effects at all.
}

// lastly, define the functions that run the model versions here. This can be copy-pasted easily to run various model versions
// note that it might be necessary to include more parameters for some model versions depending on your needs.
// also, you may want to change observer or create a new one depending on your needs. The observer defines which and how metrics are reported in the CSV-files.
// The observer also defines how the CSV-files will be names, see "File:" declaration below.

func runNursebeecs(app *app.App, idx int, filename string, params params.Params, paramsNbeecs params.ParamsNursebeecs) {
	// define beecs_ecotox here by choosing model.DefaultEtox
	app = model.DefaultNbeecs(params, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.Debug{},
		File:     fmt.Sprintf("out/"+filename+"-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func runNursebeecsEtox(app *app.App, idx int, filename string, params params.Params, paramsEtox params.ParamsEtox, paramsNbeecs params.ParamsNursebeecs) {
	// define beecs_ecotox here by choosing model.DefaultEtox
	app = model.DefaultNbeecsEtox(params, paramsEtox, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.Debug{},
		File:     fmt.Sprintf("out/"+filename+"-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

// the created CSV-files can be aggregated and visualized depending on your preferences and needs; rudimentary visualizations using Python's Matplotlib can be found
// in the py-folders of etox_validation_testing or nursebeecs_testing.
