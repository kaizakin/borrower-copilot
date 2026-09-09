package engine

import "fmt"

// tryProfile recomputes core numbers and the verdict for a hypothetical
// profile -- the shape every lever below uses to test one change without
// touching the borrower's real profile, and without duplicating Assess's
// arithmetic or its risk gates.
func tryProfile(p Profile, mutate func(*Profile)) (coreCalc, VerdictResult) {
	mutate(&p)
	c := computeCore(p)
	return c, computeVerdict(p, c.income, c.emiCeiling, c.safeCapacity)
}

// findLevers searches a fixed set of independent single-variable changes
// -- one at a time, everything else held at the borrower's actual answers
// -- for what it would take to flip the verdict to Borrow. Assess calls
// this only when the plain verdict isn't already Borrow.
func findLevers(p Profile) []LeverResult {
	var levers []LeverResult
	if l, ok := reduceEMILever(p); ok {
		levers = append(levers, l)
	}
	if l, ok := creditScoreLever(p); ok {
		levers = append(levers, l)
	}
	if l, ok := savingsLever(p); ok {
		levers = append(levers, l)
	}
	if l, ok := securedLever(p); ok {
		levers = append(levers, l)
	}
	return levers
}

// reduceEMILever binary-searches the smallest cut to ExistingEMIs that
// flips the verdict to Borrow. Cutting existing EMIs only ever raises the
// EMI ceiling and safe capacity (and never worsens either risk gate in
// computeVerdict), so the verdict is monotonic in this one variable and a
// binary search finds the boundary.
func reduceEMILever(p Profile) (LeverResult, bool) {
	if p.ExistingEMIs <= 0 {
		return LeverResult{}, false
	}
	zeroCore, zeroV := tryProfile(p, func(m *Profile) { m.ExistingEMIs = 0 })
	if zeroV.Verdict != VerdictBorrow {
		return LeverResult{
			Lever: "Reduce existing EMIs",
			Note: fmt.Sprintf(
				"even paying off all ₹%s/month of your existing EMIs only raises safe capacity to ₹%s -- not enough alone, combine it with another change here",
				formatINR(p.ExistingEMIs), formatINR(zeroCore.safeCapacity),
			),
			Achievable: false,
		}, true
	}
	lo, hi := 0.0, p.ExistingEMIs
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2
		if _, v := tryProfile(p, func(m *Profile) { m.ExistingEMIs = mid }); v.Verdict == VerdictBorrow {
			lo = mid
		} else {
			hi = mid
		}
	}
	return LeverResult{
		Lever: "Reduce existing EMIs",
		Note: fmt.Sprintf(
			"cut your existing EMIs by about ₹%s/month, to ₹%s, and you can safely borrow the full ₹%s you want",
			formatINR(p.ExistingEMIs-lo), formatINR(lo), formatINR(p.AmountWanted),
		),
		Achievable: true,
	}, true
}

// creditScoreLever tests raising CreditScore to the good-credit threshold.
// Skipped once the borrower is already there -- it has nothing left to
// offer.
func creditScoreLever(p Profile) (LeverResult, bool) {
	if p.CreditScore != nil && *p.CreditScore >= goodCreditScore {
		return LeverResult{}, false
	}
	target := goodCreditScore
	c, v := tryProfile(p, func(m *Profile) { s := target; m.CreditScore = &s })
	note := fmt.Sprintf("improve your CIBIL score to %d+ and safe capacity moves to ₹%s", target, formatINR(c.safeCapacity))
	achievable := v.Verdict == VerdictBorrow
	if achievable {
		note += fmt.Sprintf(" -- enough to cover the ₹%s you want", formatINR(p.AmountWanted))
	} else {
		note += " -- not enough alone, combine it with another change here"
	}
	return LeverResult{Lever: "Improve credit score", Note: note, Achievable: achievable}, true
}

// savingsLever tests building EmergencySavingsMonths up to the threshold
// that earns the FOIR bonus. Skipped once the borrower is already there.
func savingsLever(p Profile) (LeverResult, bool) {
	if p.EmergencySavingsMonths != nil && *p.EmergencySavingsMonths >= emergencySavingsMonthsThreshold {
		return LeverResult{}, false
	}
	target := emergencySavingsMonthsThreshold
	c, v := tryProfile(p, func(m *Profile) { s := target; m.EmergencySavingsMonths = &s })
	note := fmt.Sprintf("build savings up to %.0f months of expenses and safe capacity moves to ₹%s", target, formatINR(c.safeCapacity))
	achievable := v.Verdict == VerdictBorrow
	if achievable {
		note += fmt.Sprintf(" -- enough to cover the ₹%s you want", formatINR(p.AmountWanted))
	} else {
		note += " -- not enough alone, combine it with another change here"
	}
	return LeverResult{Lever: "Build emergency savings", Note: note, Achievable: achievable}, true
}

// securedLever tests pricing the same collateral as a secured loan
// instead of the product the borrower originally asked for. Only offered
// when collateral is disclosed and the profile isn't already priced as
// secured (the auto-routing case in Assess already covers that).
func securedLever(p Profile) (LeverResult, bool) {
	hasCollateral := p.CollateralValue != nil && *p.CollateralValue > 0
	if !hasCollateral {
		return LeverResult{}, false
	}
	baseline := computeCore(p)
	if baseline.secured {
		return LeverResult{}, false
	}
	c, v := tryProfile(p, func(m *Profile) { m.LoanType = LoanSecured })
	note := fmt.Sprintf(
		"switch to a secured loan against your ₹%s collateral: your rate drops from %.1f%%-%.1f%% to %.1f%%-%.1f%% and safe capacity moves to ₹%s",
		formatINR(*p.CollateralValue), baseline.low, baseline.high, c.low, c.high, formatINR(c.safeCapacity),
	)
	achievable := v.Verdict == VerdictBorrow
	if achievable {
		note += fmt.Sprintf(" -- enough to cover the ₹%s you want", formatINR(p.AmountWanted))
	} else {
		note += " -- not enough alone, combine it with another change here"
	}
	return LeverResult{Lever: "Switch to secured lending", Note: note, Achievable: achievable}, true
}
