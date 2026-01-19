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

	p := params.Default()
	p.Termination.MaxTicks = 365

	start := time.Now()

	run_beecs := true // switch to run normal and/or nurse beecs
	if run_beecs {
		for i := 0; i < 100; i++ {
			run(app, i, &p)
		}
	}
	dur := time.Since(start)
	fmt.Println(dur)

	run_nbeecs := true // switch to run normal and/or nurse beecs
	if run_nbeecs {

		p.Nursing.WinterBees = true

		pn := params.DefaultNursebeecs()
		pn.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pn.NursingRework.NewBroodCare = true

		for i := 0; i < 100; i++ {
			run_nursebeecs(app, i, &p, &pn)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)

	run_nbeecs2 := true // switch to run normal and/or nurse beecs
	if run_nbeecs2 {

		p.Nursing.WinterBees = true

		pn := params.DefaultNursebeecs()
		pn.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pn.NursingRework.NewBroodCare = true
		pn.NursingRework.Nursebeecsv1 = true

		for i := 0; i < 100; i++ {
			run_nursebeecs2(app, i, &p, &pn)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params) {
	app = model.Default(params, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugPollenCons{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs(app *app.App, idx int, params params.Params, paramsNbeecs params.ParamsNursebeecs) {
	app = model.DefaultNbeecs(params, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugPollenCons{},
		File:     fmt.Sprintf("out/old-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, paramsNbeecs params.ParamsNursebeecs) {
	app = model.DefaultNbeecs(params, paramsNbeecs, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugPollenCons{},
		File:     fmt.Sprintf("out/new-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
