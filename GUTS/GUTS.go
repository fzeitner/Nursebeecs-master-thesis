package GUTS

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/mlange-42/ark/ecs"
)

// this package includes the calculations of the reduced GUTS models as per Baas et al. 2022 and Jager et al. 2011

func SD_for(Oraldose float64, Contactdose float64, C_i float64, rng *rand.Rand, w *ecs.World) (lethaldose bool, C_oral float64, C_contact float64, Ci float64) {
	// runs GUTS-RED-SD based on BeeGUTS (Baas et al. 2022) and the GUTS model framework created by Jager et al. 2011
	// uses... to return...
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	lethaldose = false
	C_oral = Oraldose
	C_contact = Contactdose
	Ci = C_i

	for i := range gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_contact + C_oral

		Ci = math.Max(0, Ci+(gutsparams.Kd_SD*(current_dose-Ci))*1/float64(gutsparams.T))
		h := math.Max(0, gutsparams.BW_SD*(Ci-gutsparams.MW_SD)*1/float64(gutsparams.T))

		if rng.Float64() > math.Exp(-h) {
			lethaldose = true
			break
		}
		i += 1
	}

	return
}

func SD_IHbee(Oraldose float64, Contactdose float64, C_i float64, rng *rand.Rand, w *ecs.World) (lethaldose bool, C_oral float64, C_contact float64, Ci float64) {
	// runs GUTS-RED-SD based on BeeGUTS (Baas et al. 2022) and the GUTS model framework created by Jager et al. 2011
	// uses... to return...
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	lethaldose = false
	C_oral = Oraldose
	C_contact = Contactdose
	Ci = C_i

	for i := range gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_contact + C_oral

		Ci = math.Max(0, Ci+(gutsparams.Kd_SD*(current_dose-Ci))*1/float64(gutsparams.T))
		h := math.Max(0, gutsparams.BW_SD*(Ci-gutsparams.MW_SD)*1/float64(gutsparams.T))

		if rng.Float64() > math.Exp(-h) {
			lethaldose = true
		}
		i += 1
	}

	return
}

func IT(Fval float64, Oraldose float64, Contactdose float64, C_i float64, w *ecs.World) (lethaldose bool, C_oral float64, C_contact float64, Ci float64) {
	// runs GUTS-RED-IT based on BeeGUTS (Baas et al. 2022) and the GUTS model framework created by Jager et al. 2011
	// uses... to return...
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	lethaldose = false
	C_oral = Oraldose
	C_contact = Contactdose
	Ci = C_i

	for i := range gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_contact + C_oral

		Ci = math.Max(0, Ci+(gutsparams.Kd_IT*(current_dose-Ci))*1/float64(gutsparams.T))
		if Ci >= Fval {
			lethaldose = true
			break
		}

		i += 1
	}

	return
}

func Calc_F(zval float64, w *ecs.World) (Fval float64) {
	// calculates a value of the log-logistic distribution depending on GUTS parameters and a random value between 0 and 1 given to this function when called (zval)
	// still needs to be debugged and assured, but is based on the "Modelling survival under chemical stress" (Jager&Ashauer 2018) and should be theoretically sound to work with
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	beta := math.Log(39) / math.Log(gutsparams.F_S)
	Fval = 1 / (1 + (math.Pow((zval / gutsparams.MW_IT), -1*beta)))
	return
}
