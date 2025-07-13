package main

import (
	"fmt"
	"time"

	"github.com/fzeitner/beecs_ecotox/model"
	"github.com/fzeitner/beecs_ecotox/obs"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	app := app.New()

	p := params.Default()
	p.Termination.MaxTicks = 365

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params) {
	app = model.Default(params, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugEcotox{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
