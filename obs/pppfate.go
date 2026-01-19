package obs

import (
	"github.com/fzeitner/Nursebeecs-master-thesis/globals"
	"github.com/mlange-42/ark/ecs"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of Nursebeecs-master-thesis against BEEHAVE_ecotox.
type PPPFateObs struct {
	pppfate *globals.PPPFate
	stores  *globals.StoragesEtox
	data    []float64
}

func (o *PPPFateObs) Initialize(w *ecs.World) {
	o.pppfate = ecs.GetResource[globals.PPPFate](w)
	o.stores = ecs.GetResource[globals.StoragesEtox](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *PPPFateObs) Update(w *ecs.World) {}
func (o *PPPFateObs) Header() []string {
	return []string{"TotalPPForaged", "PPPhoneyStores", "PPPpollenStores", "PPPforagersImmediate", "PPPforagerDiedinFlight", "PPPforagersinHive", "PPPforagersTotal", "PPPIHbees", "PPPnurses", "PPPlarvae", "PPPdrones", "PPPdlarvae", "PPPpollenTotal", "PPPhoneyTotal", "PPPIHTotal"}
}
func (o *PPPFateObs) Values(w *ecs.World) []float64 {
	o.data[0] = o.pppfate.TotalPPPforaged
	o.data[1] = o.pppfate.PPPhoneyStores
	o.data[2] = o.pppfate.PPPpollenStores
	o.data[3] = o.pppfate.PPPforagersImmediate
	o.data[4] = o.pppfate.ForagerDiedInFlight

	o.data[5] = o.pppfate.PPPforagersinHive
	o.data[6] = o.pppfate.PPPforagersTotal
	o.data[7] = o.pppfate.PPPIHbees
	o.data[8] = o.pppfate.PPPNurses
	o.data[9] = o.pppfate.PPPlarvae
	o.data[10] = o.pppfate.PPPdrones
	o.data[11] = o.pppfate.PPPdlarvae

	o.data[12] = o.stores.PPPpollenTotal
	o.data[13] = o.stores.PPPhoneyTotal
	o.data[14] = o.stores.PPPTotal

	return o.data
}
