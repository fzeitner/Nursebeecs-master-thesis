package globals

// globals necessary for the _ecotox additions of water foraging, ecotoxicological variables and honey storage compartimentation
// none of this is used with the default ecotox settings and should not, because of a lack of testing atm

type WaterNeeds struct {
	ETOX_Waterneedforcooling float64 // The amount of water needed for cooling today
}

// WaterForagingPeriodData contains data on daily foraging hours.
type WaterForagingPeriodData struct {
	// Foraging period per day [h].
	// First index: year, second index: day of year.
	Years [][]float64
	// The currently selected year.
	CurrentYear int
}
