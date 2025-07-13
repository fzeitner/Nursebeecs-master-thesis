package obs

import (
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/mlange-42/ark/ecs"
)

// ForagingPeriod is a row observer for the foraging period of the current day, in hours.
//
// Has a single column "Foraging Period [h]", and reports one row/value per model tick.
type ForagingPeriod struct {
	period *globals.ForagingPeriod
	data   []float64
}

func (o *ForagingPeriod) Initialize(w *ecs.World) {
	o.period = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *ForagingPeriod) Update(w *ecs.World) {}
func (o *ForagingPeriod) Header() []string {
	return []string{"Foraging Period [h]"}
}
func (o *ForagingPeriod) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.period.SecondsToday) / 3600.0

	return o.data
}

// WaterForCooling is a row observer for the water [g] needed for cooling of the current day
type WaterForCooling struct {
	waterNeeds *globals.WaterNeeds
	data       []float64
}

func (o *WaterForCooling) Initialize(w *ecs.World) {
	o.waterNeeds = ecs.GetResource[globals.WaterNeeds](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *WaterForCooling) Update(w *ecs.World) {}
func (o *WaterForCooling) Header() []string {
	return []string{"Water for cooling [g]"}
}
func (o *WaterForCooling) Values(w *ecs.World) []float64 {
	o.data[0] = o.waterNeeds.ETOX_Waterneedforcooling

	return o.data
}
