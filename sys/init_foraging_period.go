package sys

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/fzeitner/beecs_ecotox/data"
	"github.com/fzeitner/beecs_ecotox/globals"
	"github.com/fzeitner/beecs_ecotox/params"
	"github.com/fzeitner/beecs_ecotox/util"
	"github.com/mlange-42/ark/ecs"
)

// WATERFORAGING IMPLEMENTATION HERE: EToX_WaterforcoolingREP <- Reporter for Water need per day
type InitForagingPeriod struct {
	periodData      globals.ForagingPeriodData
	waterperiodData globals.WaterForagingPeriodData
}

func (s *InitForagingPeriod) Initialize(w *ecs.World) {
	s.periodData = globals.ForagingPeriodData{}
	ecs.AddResource(w, &s.periodData)

	periodParams := ecs.GetResource[params.ForagingPeriod](w)
	var fileSys fs.FS = data.ForagingPeriod
	if !periodParams.Builtin {
		wd := ecs.GetResource[params.WorkingDirectory](w).Path
		fileSys = os.DirFS(wd)
	}

	// fill from data provided directly
	for _, arr := range periodParams.Years {
		if len(arr)%365 != 0 {
			log.Fatal(fmt.Errorf("foraging period entries requires multiple of 365 values, parameters have %d", len(arr)))
		}
		years := len(arr) / 365
		for year := 0; year < years; year++ {
			s.periodData.Years = append(s.periodData.Years, arr[year*365:(year+1)*365])
		}
	}

	// fill from files
	for _, f := range periodParams.Files {
		arr, err := util.FloatArrayFromFile(fileSys, f)
		if err != nil {
			log.Fatal(fmt.Errorf("error reading foraging period file '%s': %s", f, err.Error()))
		}
		if len(arr)%365 != 0 {
			log.Fatal(fmt.Errorf("foraging period file requires multiple of 365 values, '%s' has %d", f, len(arr)))
		}
		years := len(arr) / 365
		for year := 0; year < years; year++ {
			s.periodData.Years = append(s.periodData.Years, arr[year*365:(year+1)*365])
		}
	}

	// water foraging from here on, does work with embedded files, but water foraging as a whole is not tested
	// and should not be turned on as of yet
	waterParams := ecs.GetResource[params.WaterParams](w)
	s.waterperiodData = globals.WaterForagingPeriodData{}
	ecs.AddResource(w, &s.waterperiodData)
	if waterParams.WaterForaging {
		waterPeriodParams := ecs.GetResource[params.WaterForagingPeriod](w)
		var WaterfileSys fs.FS = data.WaterNeedsDaily
		if !periodParams.Builtin {
			wd := ecs.GetResource[params.WorkingDirectory](w).Path
			WaterfileSys = os.DirFS(wd)
		}

		// fill from data provided directly
		for _, arr := range waterPeriodParams.Years {
			if len(arr)%365 != 0 {
				log.Fatal(fmt.Errorf("foraging period entries requires multiple of 365 values, parameters have %d", len(arr)))
			}
			years := len(arr) / 365
			for year := 0; year < years; year++ {
				s.waterperiodData.Years = append(s.waterperiodData.Years, arr[year*365:(year+1)*365])
			}
		}

		// fill from files
		for _, f := range waterPeriodParams.Files {
			arr, err := util.FloatArrayFromFile(WaterfileSys, f)
			if err != nil {
				log.Fatal(fmt.Errorf("error reading foraging period file '%s': %s", f, err.Error()))
			}
			if len(arr)%365 != 0 {
				log.Fatal(fmt.Errorf("foraging period file requires multiple of 365 values, '%s' has %d", f, len(arr)))
			}
			years := len(arr) / 365
			for year := 0; year < years; year++ {
				s.waterperiodData.Years = append(s.waterperiodData.Years, arr[year*365:(year+1)*365])
			}
		}
	}
}

func (s *InitForagingPeriod) Update(w *ecs.World) {}

func (s *InitForagingPeriod) Finalize(w *ecs.World) {}
