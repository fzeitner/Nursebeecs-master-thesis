package comp_etox

// PatchProperties component for flower patches.
type PatchProperties_etox struct {
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kJ]
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/g]
	PPPcontactDose         float64 // PPP concentration for contact exposure on patch [mug] ??
}

// Resource component for flower patches.
//
// Holds information on available nectar and pollen resources.
type Resource_etox struct {
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kJ]
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/g]
	PPPcontactDose         float64 // PPP concentration for contact exposure on patch [mug] ??
}
