package engine

import "fmt"

// computeVerdict applies risk gates before ever comparing capacity to the
// amount wanted -- a profile that is already underwater, or that has both
// a recent bounce and almost no safe room, is told not to borrow
// regardless of how small the request is.
func computeVerdict(p Profile, income, emiCeiling, safeCapacity float64) VerdictResult {
	if income-p.HouseholdExpenses-p.ExistingEMIs <= 0 {
		return VerdictResult{
			Verdict: VerdictDontBorrow,
			Justification: fmt.Sprintf(
				"don't borrow: your existing EMIs (₹%s) and household expenses (₹%s) already consume all of your ₹%s income, leaving nothing for a new EMI",
				formatINR(p.ExistingEMIs), formatINR(p.HouseholdExpenses), formatINR(income),
			),
		}
	}
	if p.PastBounces != nil && *p.PastBounces > 0 && emiCeiling < income*0.05 {
		return VerdictResult{
			Verdict: VerdictDontBorrow,
			Justification: fmt.Sprintf(
				"don't borrow: you've missed a payment in the last 3 months and your safe EMI ceiling is only ₹%s -- taking on more debt now risks another bounce and further credit damage",
				formatINR(emiCeiling),
			),
		}
	}
	if emiCeiling <= 0 || safeCapacity <= 0 {
		return VerdictResult{
			Verdict:       VerdictDontBorrow,
			Justification: "don't borrow: after your existing EMIs and household expenses, there is no safe room left in your budget for a new EMI",
		}
	}
	if safeCapacity >= p.AmountWanted {
		return VerdictResult{
			Verdict: VerdictBorrow,
			Justification: fmt.Sprintf(
				"borrow: your safe capacity of ₹%s comfortably covers the ₹%s you want",
				formatINR(safeCapacity), formatINR(p.AmountWanted),
			),
		}
	}
	return VerdictResult{
		Verdict: VerdictBorrowLess,
		Justification: fmt.Sprintf(
			"borrow less: your safe capacity is ₹%s, below the ₹%s you asked for -- borrowing the full amount would push your EMI past what your budget can sustain",
			formatINR(safeCapacity), formatINR(p.AmountWanted),
		),
	}
}

// stressScenario recomputes the EMI ceiling under a shocked income or rate
// and reports whether the requested loan's own EMI would still survive it.
func stressScenario(income, householdExpenses, existingEMIs, foirPercent, rate float64, tenureMonths int, amountWanted float64, label string) StressScenario {
	foirEMI := income*foirPercent/100 - existingEMIs
	residual := income - householdExpenses - existingEMIs
	residualEMI := residual * residualSafetyFactor
	ceiling := foirEMI
	if residualEMI < ceiling {
		ceiling = residualEMI
	}
	if ceiling < 0 {
		ceiling = 0
	}
	requiredEMI := EMI(amountWanted, rate, tenureMonths)
	stillOK := ceiling >= requiredEMI
	note := fmt.Sprintf("if %s, your safe EMI ceiling becomes ₹%s", label, formatINR(ceiling))
	if stillOK {
		note += fmt.Sprintf(" -- still enough to cover the ₹%s EMI this loan needs", formatINR(requiredEMI))
	} else {
		note += fmt.Sprintf(", below the ₹%s EMI this loan would need -- it does not survive this stress test", formatINR(requiredEMI))
	}
	return StressScenario{Description: label, NewCeiling: ceiling, StillAffordable: stillOK, Note: note}
}
