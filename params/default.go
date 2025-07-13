package params

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"os"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Params is an interface for parameter sets.
type Params interface {
	// Apply the parameters to a world.
	Apply(world *ecs.World)
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSONFile(path string) error
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSON(data []byte) error
}

// DefaultParams contains all default parameters of BEEHAVE.
//
// DefaultParams implements [Params].
type DefaultParams struct {
	WorkingDirectory  WorkingDirectory
	Termination       Termination
	InitialPatches    InitialPatches
	Nursing           Nursing
	Foraging          Foraging
	Foragers          Foragers
	ForagingPeriod    ForagingPeriod
	Dance             Dance
	HandlingTime      HandlingTime
	WorkerMortality   WorkerMortality
	DroneMortality    DroneMortality
	HoneyNeeds        HoneyNeeds
	PollenNeeds       PollenNeeds
	Stores            Stores
	WorkerDevelopment WorkerDevelopment
	DroneDevelopment  DroneDevelopment
	InitialPopulation InitialPopulation
	AgeFirstForaging  AgeFirstForaging
	EnergyContent     EnergyContent
	InitialStores     InitialStores
	RandomSeed        RandomSeed

	ETOXparams          ETOXparams
	WaterParams         WaterParams
	WaterForagingPeriod WaterForagingPeriod
}

// Default returns the complete default parameter set of BEEHAVE.
func Default() DefaultParams {
	return DefaultParams{
		WorkingDirectory: WorkingDirectory{
			Path: ".",
		},
		Termination: Termination{
			MaxTicks:     365,
			OnExtinction: false,
		},
		RandomSeed: RandomSeed{
			Seed: 0,
		},
		WorkerDevelopment: WorkerDevelopment{
			EggTime:     3,
			LarvaeTime:  6,
			PupaeTime:   12,
			MaxLifespan: 290,
		},
		DroneDevelopment: DroneDevelopment{
			EggTime:     3,
			LarvaeTime:  7,
			PupaeTime:   14,
			MaxLifespan: 37,
		},
		WorkerMortality: WorkerMortality{
			Eggs:      0.03,
			Larvae:    0.01,
			Pupae:     0.001,
			InHive:    0.004,
			MaxMilage: 800,
		},
		DroneMortality: DroneMortality{
			Eggs:   0.064,
			Larvae: 0.044,
			Pupae:  0.005,
			InHive: 0.05,
		},
		AgeFirstForaging: AgeFirstForaging{
			Base: 21,
			Min:  7,
			Max:  50,
		},
		Foragers: Foragers{
			FlightVelocity: 6.5,      // [m/s]
			FlightCostPerM: 0.000006, // [kJ/m]
			NectarLoad:     50,       // [muL]
			PollenLoad:     0.015,    // [g]
			MaxKmPerDay:    7299,     // ???
			SquadronSize:   100,
		},
		Foraging: Foraging{
			ProbBase:      0.01,
			ProbHigh:      0.05,
			ProbEmergency: 0.2,

			SearchLength: 6.5 * 60 * 17, // [m] (6630m, 17 min)

			EnergyOnFlower:  0.2,
			MortalityPerSec: 0.00001,

			StopProbability:     0.3,
			AbandonPollenPerSec: 0.00002,
		},
		ForagingPeriod: ForagingPeriod{
			Files:       []string{"foraging-period/berlin2000.txt"},
			Builtin:     true,
			RandomYears: false,
		},
		WaterForagingPeriod: WaterForagingPeriod{
			Files:       []string{"ETOX_waterforcooling_daily/waterlistExample.txt"},
			Builtin:     true,
			RandomYears: false,
		},
		HandlingTime: HandlingTime{
			NectarGathering:      1200,
			PollenGathering:      600,
			NectarUnloading:      116,
			PollenUnloading:      210,
			ConstantHandlingTime: false,

			ETOX_handlingTimeWater: 64,
		},
		Dance: Dance{
			Slope:                       1.16,
			Intercept:                   0.0,
			MaxCircuits:                 117,
			FindProbability:             0.5,
			PollenDanceFollowers:        2,
			MaxProportionPollenForagers: 0.8,
		},
		EnergyContent: EnergyContent{
			Honey:   12.78,
			Sucrose: 0.00582,
		},
		Stores: Stores{
			IdealPollenStoreDays: 7,
			MinIdealPollenStore:  250.0,
			MaxHoneyStoreKg:      50.0, // [kg]
			DecentHoneyPerWorker: 1.5,  // [g]
			ProteinStoreNurse:    7,    // [d]
			ETOXDensityOfHoney:   1.4,  // [kg/l]
		},
		HoneyNeeds: HoneyNeeds{
			WorkerResting:    11.0,  // [mg/d]
			WorkerNurse:      53.42, // [mg/d]
			WorkerLarvaTotal: 65.4,  // [mg]
			DroneLarva:       19.2,  // [mg/d]
			Drone:            10.0,  // [mg/d]
		},
		PollenNeeds: PollenNeeds{
			WorkerLarvaTotal: 142.0, // [mg]
			DroneLarva:       50.0,  // [mg/d]
			Worker:           1.5,   // [mg/d]
			Drone:            2.0,   // [mg/d]
		},
		Nursing: Nursing{
			MaxBroodNurseRatio:         3.0,
			ForagerNursingContribution: 0.2,
			MaxEggsPerDay:              1600,
			DroneEggsProportion:        0.04,
			EggNursingLimit:            true,
			MaxBroodCells:              200_000,
			DroneEggLayingSeasonStart:  115,
			DroneEggLayingSeasonEnd:    240,
		},
		InitialPopulation: InitialPopulation{
			Count:     10_000,
			MinAge:    100,
			MaxAge:    160,
			MinMilage: 0,
			MaxMilage: 200,
		},
		InitialStores: InitialStores{
			Honey:  25,  // [kg]
			Pollen: 100, // [g]
		},
		InitialPatches: InitialPatches{
			Patches: []comp.PatchConfig{
				{
					DistToColony: 1500,
					ConstantPatch: &comp.ConstantPatch{
						Nectar:               20,
						NectarConcentration:  1.5,
						Pollen:               1,
						DetectionProbability: 0.2,
					},
				},
				{
					DistToColony: 500,
					ConstantPatch: &comp.ConstantPatch{
						Nectar:               20,
						NectarConcentration:  1.5,
						Pollen:               1,
						DetectionProbability: 0.2,
					},
				},
			},
		},
		ETOXparams: ETOXparams{
			Application: false,
			PPPname:     "No applications", // Identifier for the PPP used.
			//ApplicationRate:        0.,                // Application rate of PPP [kg/ha].                  // probably take this out, not used atm and probably wont be
			PPPconcentrationNectar: 990,
			PPPconcentrationPollen: 26631,
			PPPcontactExposure:     0.3,   // actually dont have a value for this
			DT50:                   1000., // Whole plant DT50 from residue studies [d].
			AppDay:                 189,   // Day of the year in which application starts [d].
			ExposurePeriod:         8,     // Duration of exposure happening (irrespective of DT50) [d].
			SpinupPhase:            0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
			ExposurePhase:          3,     // Number of years in which exposure takes place [y].

			ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
			DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
			ContactSum:                false, // Determines whether contact exposure should be summed up per visit to a patch (true) or if the mean should be calculated whenever a new patch is visited (false)
			ContactExposureOneDay:     false, // Determines whether contact exposure should only be possible on the day of application

			ForagerOralLD50:  1000., // Lethal oral dose for 50% mortality of foragers [µg/bee].
			ForagerOralSlope: 100.,  // Slope of the dose-response relationship (forager, oral) [ ].
			HSuptake:         0.1,   // Uptake of a given percentage of ai in the honey stomach by the forager bees

			ForagerContactLD50:  0.6,  // Lethal dose for 50% of foragers via contact exposure [µg/bee]
			ForagerContactSlope: 1.08, // Slope of the dose-response relationship (forager, contact) [ ]

			LarvaeOralLD50:  0.0014, // Lethal oral dose for 50% mortality of larvae [µg/larvae]
			LarvaeOralSlope: 1.6,    // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented

			NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
			NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]
			DT50honey:       60.,  // Honey DT50 [d]
			RUD:             21.,  // Residue per Unit Dose  [(ha*mg)/(kg*kg)]
		},
		WaterParams: WaterParams{
			WaterForaging:             false,       // Determines whether water foraging takes place or not.
			ETOX_cropvolume_water:     44. / 1000., // [g]: 44 mg water per forager Visscher et al. 1996
			ETOX_Watertripsperh:       7.,          // Number of trips per h for water foragers
			ETOX_Durationofwatertrips: 1.,          // How long do the bees go out in the morning exclusively searching for water 1 h according to Lindauer 1954

		},
	}
}

// FromJSONFile fills the parameter set with values from a JSON file.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParams) FromJSONFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return p.FromJSON(content)
}

// FromJSON fills the parameter set with values from JSON.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParams) FromJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(p)
}

// Apply the parameters to a world by adding them as resources.
func (p *DefaultParams) Apply(world *ecs.World) {
	// Random seed
	seed := p.RandomSeed
	if seed.Seed <= 0 {
		seed.Seed = int(rand.Int32())
	}
	rng := ecs.GetResource[resource.Rand](world)
	rng.Source = rand.NewPCG(0, uint64(seed.Seed))

	pCopy := *p

	// Resources
	ecs.AddResource(world, &seed)
	ecs.AddResource(world, &pCopy.WorkingDirectory)
	ecs.AddResource(world, &pCopy.Termination)
	ecs.AddResource(world, &pCopy.WorkerDevelopment)
	ecs.AddResource(world, &pCopy.DroneDevelopment)
	ecs.AddResource(world, &pCopy.WorkerMortality)
	ecs.AddResource(world, &pCopy.DroneMortality)
	ecs.AddResource(world, &pCopy.AgeFirstForaging)
	ecs.AddResource(world, &pCopy.Foragers)
	ecs.AddResource(world, &pCopy.Foraging)
	ecs.AddResource(world, &pCopy.ForagingPeriod)
	ecs.AddResource(world, &pCopy.WaterForagingPeriod)
	ecs.AddResource(world, &pCopy.HandlingTime)
	ecs.AddResource(world, &pCopy.Dance)
	ecs.AddResource(world, &pCopy.EnergyContent)
	ecs.AddResource(world, &pCopy.Stores)
	ecs.AddResource(world, &pCopy.HoneyNeeds)
	ecs.AddResource(world, &pCopy.PollenNeeds)
	ecs.AddResource(world, &pCopy.Nursing)
	ecs.AddResource(world, &pCopy.InitialPopulation)
	ecs.AddResource(world, &pCopy.InitialStores)
	ecs.AddResource(world, &pCopy.InitialPatches)
	ecs.AddResource(world, &pCopy.ETOXparams)
	ecs.AddResource(world, &pCopy.WaterParams)
}
