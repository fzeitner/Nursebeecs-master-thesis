// trying to create a way to let sim start only in tick 1
// meaning I habe to pause systems on step 0 and unpause them on 1 I guess
package sys

import (
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

type TimeControl struct {
	time *resource.Tick
	app  *app.App
}

func (s *TimeControl) Initialize(w *ecs.World) {
	s.time = ecs.GetResource[resource.Tick](w)
	s.app = ecs.GetResource[app.App](w)
	if s.time.Tick == 0 { // can this ever work here?
		s.app.Systems.Paused = true
	} else {
		s.app.Systems.Paused = false
	}
}

func (s *TimeControl) Update(w *ecs.World) {

	if s.time.Tick == 0 { // can this ever work here?
		s.app.Systems.Paused = true
	} else {
		s.app.Systems.Paused = false
	}
}

func (s *TimeControl) Finalize(w *ecs.World) {}
