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

	pe.ETOXparams.HSUfix = false
	pe.ETOXparams.Nursebeefix = false
	pe.ETOXparams.ReworkedThermoETOX = false

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p, &pe)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, params_etox params_etox.Params_etox) {
	app = model_etox.Default(params, params_etox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugDrones{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
