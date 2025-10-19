package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/beecs_masterthesis/model_etox"
	"github.com/fzeitner/beecs_masterthesis/obs"
	"github.com/fzeitner/beecs_masterthesis/params"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	p.Termination.MaxTicks = 365

	pe := params_etox.Default_etox()

	start := time.Now()

	run_beecs := false // switch to run normal and/or nurse beecs
	if run_beecs {
		for i := 0; i < 100; i++ {
			run(app, i, &p, &pe)
		}
	}
	dur := time.Since(start)
	fmt.Println(dur)

	run_nbeecs := true // switch to run normal and/or nurse beecs
	if run_nbeecs {
		pe.ConsumptionRework.Nursebeecs = true
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
		pe.ConsumptionRework.Nursebeecs = true
		pe.ConsumptionRework.HoneyAdultWorker = 11. // old BEEHAVE val
		pe.Nursing.NewBroodCare = true

		for i := 0; i < 100; i++ {
			run_nursebeecs2(app, i, &p, &pe)
		}
	}
	dur = time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.NurseBeecsDefault(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/oldbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}

func run_nursebeecs2(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.NurseBeecsDefault(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugNursing{},
		File:     fmt.Sprintf("out/newbc-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
