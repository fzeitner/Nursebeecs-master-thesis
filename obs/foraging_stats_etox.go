package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// ForagingStats is a table observer for foraging activity over the day.
type ForagingStatsEtox struct {
	Relative bool
	stats    *globals.ForagingStatsEtox
	pop      *globals.PopulationStats
	data     []float64
}

func (o *ForagingStatsEtox) Initialize(w *ecs.World) {
	o.stats = ecs.GetResource[globals.ForagingStatsEtox](w)
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *ForagingStatsEtox) Update(w *ecs.World) {}
func (o *ForagingStatsEtox) Header() []string {
	return []string{"Round", "Lazy", "Resting", "Searching", "Recruited", "Nectar", "Pollen", "Water"}
}
func (o *ForagingStatsEtox) Values(w *ecs.World) []float64 {
	for i, round := range o.stats.Rounds {
		o.data[0] = float64(i)
		o.data[1] += float64(round.Lazy)
		o.data[2] += float64(round.Resting)
		o.data[3] += float64(round.Searching)
		o.data[4] += float64(round.Recruited)
		o.data[5] += float64(round.Nectar)
		o.data[6] += float64(round.Pollen)
		o.data[7] += float64(round.Water)
	}
	for i := 1; i <= len(o.data)-1; i++ {
		if o.data[0] == 0 {
			continue
		}
		o.data[i] = o.data[i] / o.data[0]
	}
	return o.data
}
