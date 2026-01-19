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

	pe := params.DefaultEtox()

	pe.PPPApplication.HSUfix = false
	pe.PPPApplication.Nursebeefix = false
	pe.PPPApplication.ReworkedThermoETOX = false

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p, &pe)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugDrones{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
