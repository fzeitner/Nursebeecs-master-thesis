// Demonstrates how to read parameters from JSON.
package main

import (
	"fmt"
	"log"

	"github.com/fzeitner/Nursebeecs-master-thesis/model"
	"github.com/fzeitner/Nursebeecs-master-thesis/obs"
	"github.com/fzeitner/Nursebeecs-master-thesis/params"
	"github.com/mlange-42/ark-tools/reporter"
)

func main() {
	// Get the default parameters.
	p := params.Default()
	// Read JSON to modify some parameters.
	err := p.FromJSONFile("_examples/json_parameters/params.json")
	if err != nil {
		log.Fatal(err)
	}
	// Print one of the modified sections of the parameters.
	fmt.Printf("%+v\n", p.Foragers)

	// Create a model with the default sub-models.
	m := model.Default(&p, nil)

	// Add a CSV output system using observer [obs.WorkerCohorts].
	m.AddSystem(&reporter.CSV{
		Observer: &obs.WorkerCohorts{},
		File:     "out/worker-cohorts.csv",
	})

	// Run the model.
	m.Run()
}
