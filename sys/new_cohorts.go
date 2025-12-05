package sys

import (
	"github.com/fzeitner/beecs_masterthesis/globals"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Here new IHbees and drone cohorts get initialized each tick.
// The process had to be decoupled this from the other systems to get closer to original BEEHAVE,
// because the devs decided to have Broodcare happen after ageing of cohorts and thus calculating new Ihbees/Drones,
// but before actually initializing them as new cohorts and having them counted to the total (for broodcare or other processes).
// Overall somewhat of a small difference, but it is noticable in quantitative model behaviour, especially for brood.

type NewCohorts struct {
	inHive     *globals.InHive
	newCohorts *globals.NewCohorts

	time *resource.Tick
}

func (s *NewCohorts) Initialize(w *ecs.World) {
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)

	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *NewCohorts) Update(w *ecs.World) {
	s.inHive.Drones[0] = s.newCohorts.Drones
	s.inHive.Workers[0] = s.newCohorts.IHbees

	s.newCohorts.Drones = 0
	s.newCohorts.IHbees = 0
}

func (s *NewCohorts) Finalize(w *ecs.World) {}
