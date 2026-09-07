package engine

import (
	"fmt"
	"math"
	"strings"
)

// Assess is the engine's single entry point: a pure, deterministic
// function from a Profile to an Assessment.
func Assess(p Profile) Assessment {
	income, incomeNotes := pooledIncome(p)
	foir := borrowerFOIR(p)

	hasCollateral := p.CollateralValue != nil && *p.CollateralValue > 0
	requestedSecured := p.LoanType == LoanSecured
	secured := requestedSecured || (hasCollateral && p.Employment != EmploymentSalaried)

	routed := false
	routingReason := ""
	if !requestedSecured && hasCollateral && p.Employment != EmploymentSalaried {
		routed = true
		routingReason = fmt.Sprintf(
			"as a %s applicant without an established credit score, an unsecured %s loan is unlikely to be sanctioned at a fair rate; securing it against your ₹%s asset (loan-against-property/gold) instead unlocks a far lower, safer rate",
			employmentLabel(p.Employment), string(p.LoanType), formatINR(*p.CollateralValue),
		)
	}

	low, high, wide, rateNotes := rateBand(p, secured)
	midRate := (low + high) / 2
	// Quote APR at the conservative (upper) edge of the band -- the rate a
	// borrower should plan around until a lender proves otherwise.
	apr := APR(p.AmountWanted, high, processingFeePercent, p.TenureMonths)

	foirEMI := income*foir.Percent/100 - p.ExistingEMIs
	residual := income - p.HouseholdExpenses - p.ExistingEMIs
	residualEMI := residual * residualSafetyFactor
	emiCeiling := math.Min(foirEMI, residualEMI)
	ceilingDriver := "your FOIR limit"
	if residualEMI < foirEMI {
		ceilingDriver = "what's left after your household expenses"
	}
	if emiCeiling < 0 {
		emiCeiling = 0
	}
	emiJustification := fmt.Sprintf(
		"your EMI ceiling is ₹%s/month because %s caps it there (income ₹%s, FOIR %.0f%%, existing EMIs ₹%s, household expenses ₹%s)",
		formatINR(emiCeiling), ceilingDriver, formatINR(income), foir.Percent, formatINR(p.ExistingEMIs), formatINR(p.HouseholdExpenses),
	)

	stressIncome := stressScenario(income*0.8, p.HouseholdExpenses, p.ExistingEMIs, foir.Percent, midRate, p.TenureMonths, p.AmountWanted, "your income drops 20%")
	stressRate := stressScenario(income, p.HouseholdExpenses, p.ExistingEMIs, foir.Percent, midRate+2, p.TenureMonths, p.AmountWanted, "rates rise 2 points")

	safeCapacity := MaxPrincipal(emiCeiling, high, p.TenureMonths)
	safeJustification := fmt.Sprintf(
		"this is the most you can safely repay: it converts your ₹%s/month EMI ceiling into a loan amount at the conservative %.1f%% edge of your rate band over %d months",
		formatINR(emiCeiling), high, p.TenureMonths,
	)

	bf := bankFOIR(p)
	bankEMI := income*bf/100 - p.ExistingEMIs
	if bankEMI < 0 {
		bankEMI = 0
	}
	sanctionLimit := MaxPrincipal(bankEMI, midRate, p.TenureMonths)
	sanctionJustification := fmt.Sprintf(
		"a bank applying its typical %.0f%% FOIR for %s applicants to your ₹%s pooled income would sanction up to this, before it ever checks your household expenses",
		bf, employmentLabel(p.Employment), formatINR(income),
	)
	if hasCollateral {
		ltvLimit := *p.CollateralValue * collateralLTV
		if ltvLimit < sanctionLimit {
			sanctionLimit = ltvLimit
			sanctionJustification = fmt.Sprintf(
				"capped at %.0f%% loan-to-value of your ₹%s collateral, which is tighter than your income-based sanction limit",
				collateralLTV*100, formatINR(*p.CollateralValue),
			)
		}
	}

	verdict := computeVerdict(p, income, emiCeiling, safeCapacity)

	confidence := ConfidenceHigh
	confidenceNote := "core and additional questions answered; ranges reflect your specific risk profile"
	if countAnsweredAdditional(p) == 0 {
		confidence = ConfidenceLow
		confidenceNote = "only the core questions were answered -- ranges are wide because credit score, income stability, and repayment history are all unknown; answer the additional questions to narrow them"
	}

	allNotes := append(append([]string{}, incomeNotes...), foir.Notes...)
	negotiationPoints := buildNegotiationPoints(p, low, high, apr, safeCapacity, emiCeiling, rateNotes, allNotes)

	return Assessment{
		Verdict: verdict,
		MaxAmount: MaxAmountResult{
			SanctionLimit:             sanctionLimit,
			SanctionJustification:    sanctionJustification,
			SafeCapacityLimit:        safeCapacity,
			SafeCapacityJustification: safeJustification,
			Recommended:              "safe capacity",
			RecommendationReason:     "always anchor to the Safe Capacity Limit -- the Sanction Limit only reflects what a bank's formula allows, not what your actual cash flow (after rent, groceries, and existing EMIs) can sustain",
		},
		InterestRate: InterestRateResult{
			LowPercent:        low,
			HighPercent:       high,
			BandJustification: strings.Join(rateNotes, "; "),
			AllInAPR:          apr,
			APRJustification: fmt.Sprintf(
				"at the %.1f%% conservative edge of your band plus a standard %.0f%% processing fee, the true annualised cost (APR) on a ₹%s loan over %d months is %.1f%%, not %.1f%%",
				high, processingFeePercent, formatINR(p.AmountWanted), p.TenureMonths, apr, high,
			),
			Wide: wide,
		},
		EMICeiling: EMICeilingResult{
			Ceiling:            emiCeiling,
			Justification:      emiJustification,
			StressIncomeDrop20: stressIncome,
			StressRateUp2:      stressRate,
		},
		Confidence:         confidence,
		ConfidenceNote:     confidenceNote,
		Routed:             routed,
		RecommendedProduct: recommendedProductFor(p, routed),
		RoutingReason:      routingReason,
		NegotiationPoints:  negotiationPoints,
	}
}

func recommendedProductFor(p Profile, routed bool) LoanType {
	if routed {
		return LoanSecured
	}
	return p.LoanType
}

func countAnsweredAdditional(p Profile) int {
	n := 0
	if p.CreditScore != nil {
		n++
	}
	if p.IncomeStability != nil {
		n++
	}
	if p.VariableIncomeShare != nil {
		n++
	}
	if p.PastBounces != nil {
		n++
	}
	if p.EmergencySavingsMonths != nil {
		n++
	}
	if p.CollateralValue != nil {
		n++
	}
	if p.HasCoApplicant != nil {
		n++
	}
	if p.OfferedAPR != nil {
		n++
	}
	return n
}

func buildNegotiationPoints(p Profile, low, high, apr, safeCapacity, emiCeiling float64, rateNotes, otherNotes []string) []string {
	points := []string{
		fmt.Sprintf("Fair rate for my profile: %.1f%%-%.1f%% (all-in APR %.1f%% including the standard processing fee)", low, high, apr),
		fmt.Sprintf("My safe repayment capacity: ₹%s at a ₹%s/month EMI", formatINR(safeCapacity), formatINR(emiCeiling)),
	}
	points = append(points, rateNotes...)
	points = append(points, otherNotes...)
	if p.OfferedAPR != nil {
		gap := *p.OfferedAPR - high
		if gap > 0 {
			points = append(points, fmt.Sprintf("Your quoted %.1f%% is %.1f points above the top of my justified band -- I'm asking for %.1f%%", *p.OfferedAPR, gap, high))
		} else {
			points = append(points, fmt.Sprintf("Your quoted %.1f%% is already within my justified band -- I'll take it", *p.OfferedAPR))
		}
	}
	return points
}
