package engine

import "math"

// EMI returns the standard reducing-balance equated monthly instalment.
func EMI(principal, annualRatePercent float64, tenureMonths int) float64 {
	if tenureMonths <= 0 || principal <= 0 {
		return 0
	}
	r := annualRatePercent / 12 / 100
	if r == 0 {
		return principal / float64(tenureMonths)
	}
	pow := math.Pow(1+r, float64(tenureMonths))
	return principal * r * pow / (pow - 1)
}

// MaxPrincipal inverts EMI: the largest loan a given monthly EMI budget can
// service over a tenure at a rate. Returns 0 for a non-positive budget.
func MaxPrincipal(emiBudget, annualRatePercent float64, tenureMonths int) float64 {
	if tenureMonths <= 0 || emiBudget <= 0 {
		return 0
	}
	r := annualRatePercent / 12 / 100
	if r == 0 {
		return emiBudget * float64(tenureMonths)
	}
	pow := math.Pow(1+r, float64(tenureMonths))
	return emiBudget * (pow - 1) / (r * pow)
}

// APR is the RBI-mandated all-in cost of credit: it folds the mandatory
// processing fee into the interest cost and annualises over the actual
// tenure:
//
//	APR = ((Total Interest + Processing Fees) / Principal) * (365 / Tenure in Days) * 100
func APR(principal, annualRatePercent, processingFeePercent float64, tenureMonths int) float64 {
	if tenureMonths <= 0 || principal <= 0 {
		return 0
	}
	emi := EMI(principal, annualRatePercent, tenureMonths)
	totalInterest := emi*float64(tenureMonths) - principal
	fee := principal * processingFeePercent / 100
	tenureDays := float64(tenureMonths) * 30
	return ((totalInterest + fee) / principal) * (365 / tenureDays) * 100
}
