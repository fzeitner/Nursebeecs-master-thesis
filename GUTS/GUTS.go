package GUTS

import (
	"math"
	"math/rand/v2"

	"github.com/fzeitner/beecs_masterthesis/globals_etox"
	"github.com/fzeitner/beecs_masterthesis/params_etox"
	"github.com/fzeitner/beecs_masterthesis/util"
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

	i := 0
	for i < gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_contact + C_oral

		Ci = math.Max(0, Ci+(gutsparams.Kd_SD*(current_dose-Ci))*1/float64(gutsparams.T))
		h := math.Max(0, gutsparams.BW_SD*(Ci-gutsparams.MW_SD)*1/float64(gutsparams.T))

		if rng.Float64() > math.Exp(-h) {
			lethaldose = true
			break
		}
		i++
	}

	return
}

func SD_IHbee(num int, Oraldose float64, C_i float64, rng *rand.Rand, w *ecs.World) (survivors int, C_oral float64, Ci float64) {
	// runs GUTS-RED-SD based on BeeGUTS (Baas et al. 2022) and the GUTS model framework created by Jager et al. 2011
	// uses... to return...
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	toDie := 0
	C_oral = Oraldose
	//C_contact = Contactdose
	Ci = C_i

	i := 0
	for i < gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		//C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_oral //+ C_contact

		Ci = math.Max(0, Ci+(gutsparams.Kd_SD*(current_dose-Ci))*1/float64(gutsparams.T))
		h := math.Max(0, gutsparams.BW_SD*(Ci-gutsparams.MW_SD)*1/float64(gutsparams.T))

		if num >= 100 {
			toDie += int(math.Round(float64(num) * (1 - math.Exp(-h))))
		} else {
			j := 0
			for j < num {
				if rng.Float64() > math.Exp(-h) {
					toDie++
				}
				j++
			}
		}
		num -= toDie
		toDie = 0

		i++
	}
	survivors = util.MaxInt(0, num)
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

	i := 0
	for i < gutsparams.T {
		C_oral *= math.Exp(-gutsparams.K_SR / float64(gutsparams.T))
		C_contact *= math.Exp(-gutsparams.K_CA / float64(gutsparams.T))
		current_dose := C_contact + C_oral

		Ci = math.Max(0, Ci+(gutsparams.Kd_IT*(current_dose-Ci))*1/float64(gutsparams.T))
		if Ci >= Fval {
			lethaldose = true
			break
		}
		i++
	}

	return
}

func Calc_F(randomval float64, w *ecs.World) (Fval float64) {
	// calculates a value of the log-logistic distribution depending on GUTS parameters and a random value between 0 and 1 given to this function when called (zval)
	// still needs to be debugged and assured, but is based on the "Modelling survival under chemical stress" (Jager&Ashauer 2018) and should be theoretically sound to work with
	dist := ecs.GetResource[globals_etox.GUTSDistribution](w)
	num := int(math.Floor(randomval * 1000))
	Fval = dist.Dist[num]
	return
}

func Calc_F_old(zval float64, w *ecs.World) (Fval float64) {
	// calculates a value of the log-logistic distribution depending on GUTS parameters and a random value between 0 and 1 given to this function when called (zval)
	// still needs to be debugged and assured, but is based on the "Modelling survival under chemical stress" (Jager&Ashauer 2018) and should be theoretically sound to work with
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)

	beta := math.Log10(39) / math.Log10(gutsparams.F_S)
	Fval = 1 / (1 + (math.Pow((zval / gutsparams.MW_IT), -1*beta)))
	return
}

func Calc_distribution(w *ecs.World) (dist []float64) {
	gutsparams := ecs.GetResource[params_etox.GUTSParams](w)
	beta := math.Log10(39) / math.Log10(gutsparams.F_S)
	Fval := 0.
	prob := 0.001
	conc := 0.

	i := 0
	for i < 1000 {
		j := 0
		for j < 99999 {
			Fval = 1 / (1 + (math.Pow((conc / gutsparams.MW_IT), -1*beta)))
			conc += 0.001
			if Fval+0.0000001 > prob {
				break
			}
			j++
		}
		dist = append(dist, Fval)
		prob += 0.001
		i++
	}
	return
}
