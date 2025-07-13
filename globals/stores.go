package globals

// Stores of the hive.
type Stores struct {
	Honey  float64 // Stored honey [kJ].
	Pollen float64 // Stored pollen [g].

	DecentHoney float64 // Amount of honey currently considered decent [kJ].
	IdealPollen float64 // Amount of pollen currently considered ideal [g].

	ProteinFactorNurses float64 // Current protein store of nurse bees, as fraction of the maximum.

	PPPInHivePollenConc float64 // Concentration of PPP currently in stored pollen [mug/g].
	ETOX_EnergyThermo   float64 // Energy needed for Thermoregulation of hive/brood made global for Etox_consumption purposes.

	ETOX_HES_E_Capped float64 // Energy in the capped honey cells                                                 [kJ]
	ETOX_HES_C_Capped float64 // Average concentration of pesticide in the capped honey cells                     [µg/kJ]
	ETOX_HES_E_D0     float64 // Energy in the uncapped honey cells of today                                      [kJ]
	ETOX_HES_C_D0     float64 // Average concentration of pesticide in the uncapped honey cells of today          [µg/kJ]
	ETOX_HES_E_D1     float64 // Energy in the uncapped honey cells of yesterday                                  [kJ]
	ETOX_HES_C_D1     float64 // Average concentration of pesticide in the uncapped honey cells of yesterday      [µg/kJ]
	ETOX_HES_E_D2     float64 // Energy in the uncapped honey cells of two days ago                               [kJ]
	ETOX_HES_C_D2     float64 // Average concentration of pesticide in the uncapped honey cells of two days ago   [µg/kJ]
	ETOX_HES_E_D3     float64 // Energy in the uncapped honey cells of three days ago                             [kJ]
	ETOX_HES_C_D3     float64 // Average concentration of pesticide in the uncapped honey cells of three days ago [µg/kJ]
	ETOX_HES_E_D4     float64 // Energy in the uncapped honey cells of four days ago                              [kJ]
	ETOX_HES_C_D4     float64 // Average concentration of pesticide in the uncapped honey cells of four days ago  [µg/kJ]

	ETOX_Waterneedfordilution float64 // The amount of water needed for diluation of honey yesterday

	Pollenconcbeforeeating float64 // added for bugfixing
	Nectarconcbeforeeating float64 // added for bugfixing
}
