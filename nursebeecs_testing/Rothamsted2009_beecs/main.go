package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/beecs_masterthesis/model"
	"github.com/fzeitner/beecs_masterthesis/obs"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	p.Nursing.WinterBees = false // just to debug/test this once, turn off again later
	p.ForagingPeriod = params.ForagingPeriod{
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}
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
		pe := params_etox.Default_etox()
		pe.Nursing.NewConsumption = true
		pe.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pe.Nursing.NewBroodCare = false

		for i := 0; i < 100; i++ {
			run_nursebeecs(app, i, &p, &pe)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)

	run_nbeecs2 := true // switch to run normal and/or nurse beecs
	if run_nbeecs2 {
		pe := params_etox.Default_etox()
		pe.Nursing.NewConsumption = true
		pe.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pe.Nursing.NewBroodCare = true

		for i := 0; i < 100; i++ {
			run_nursebeecs2(app, i, &p, &pe)
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

func run_nursebeecs(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model.Default_nbeecs(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugPollenCons{},
		File:     fmt.Sprintf("out/oldbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model.Default_nbeecs(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugPollenCons{},
		File:     fmt.Sprintf("out/newbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
