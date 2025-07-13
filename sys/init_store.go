package sys

import (
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/mlange-42/ark/ecs"
)

// InitStore initializes and adds [globals.Stores]
// according to the settings in [params.InitialStores].
type InitStore struct{}

func (s *InitStore) Initialize(w *ecs.World) {
	init := ecs.GetResource[params.InitialStores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)
	stores := globals.Stores{
		Honey:               init.Honey * 1000.0 * energyParams.Honey,
		Pollen:              init.Pollen,
		ProteinFactorNurses: 1.0,

		ETOX_HES_E_Capped: init.Honey * 1000.0 * energyParams.Honey,
		ETOX_HES_C_Capped: 0,
		ETOX_HES_E_D0:     0,
		ETOX_HES_C_D0:     0,
		ETOX_HES_E_D1:     0,
		ETOX_HES_C_D1:     0,
		ETOX_HES_E_D2:     0,
		ETOX_HES_C_D2:     0,
		ETOX_HES_E_D3:     0,
		ETOX_HES_C_D3:     0,
		ETOX_HES_E_D4:     0,
		ETOX_HES_C_D4:     0,
	}

	ecs.AddResource(w, &stores)

}

func (s *InitStore) Update(w *ecs.World) {}

func (s *InitStore) Finalize(w *ecs.World) {}
