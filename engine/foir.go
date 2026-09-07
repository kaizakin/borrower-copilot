package engine

import "fmt"

// FOIR and rate-band tuning constants. Every value here is a modelling
// judgement call, not a regulatory constant (RBI does not mandate a single
// FOIR number) -- each one is documented with its rationale in RULES.md.
const (
	foirBaseSalaried     = 45.0
	foirBaseSelfEmployed = 40.0
	foirBaseInformal     = 30.0

	foirMin = 20.0
	foirMax = 60.0

	highIncomeThreshold = 200000.0
	highIncomeBonus     = 15.0

	goodCreditScore  = 750
	goodCreditBonus  = 5.0
	badCreditScore   = 650
	badCreditPenalty = 10.0

	perBouncePenalty = 10.0

	emergencySavingsMonthsThreshold = 3.0
	emergencySavingsBonus           = 5.0

	variableIncomeDiscount = 0.5 // variable/cash income counted at half value for affordability

	residualSafetyFactor = 0.8 // keep a 20% buffer below pure residual income

	bankFOIRSalaried     = 50.0
	bankFOIRSelfEmployed = 45.0
	bankFOIRInformal     = 35.0

	collateralLTV = 0.5 // typical loan-to-value for LAP/gold/vehicle-lien

	processingFeePercent = 2.0
)

// pooledIncome folds in variable-income discounting and an optional
// co-applicant's income, returning the affordability-relevant income and
// the reasoning for each adjustment applied.
func pooledIncome(p Profile) (income float64, notes []string) {
	income = p.NetMonthlyIncome
	if p.VariableIncomeShare != nil && *p.VariableIncomeShare > 0 {
		share := *p.VariableIncomeShare
		if share > 1 {
			share = 1
		}
		discounted := income * share * variableIncomeDiscount
		stable := income * (1 - share)
		notes = append(notes, fmt.Sprintf(
			"%.0f%% of your income is variable/cash, so it's counted at half value for affordability",
			share*100,
		))
		income = stable + discounted
	}
	if p.HasCoApplicant != nil && *p.HasCoApplicant && p.CoApplicantIncome != nil && *p.CoApplicantIncome > 0 {
		notes = append(notes, fmt.Sprintf(
			"a co-applicant's ₹%s/month income is pooled in, though you both become jointly liable for repayment",
			formatINR(*p.CoApplicantIncome),
		))
		income += *p.CoApplicantIncome
	}
	return income, notes
}

type foirResult struct {
	Percent float64
	Notes   []string
}

// borrowerFOIR computes the realistic FOIR ceiling for this specific
// borrower -- the number the borrower's own budget should be judged
// against, as opposed to a bank's looser sanctioning formula (bankFOIR).
func borrowerFOIR(p Profile) foirResult {
	var base float64
	switch p.Employment {
	case EmploymentSalaried:
		base = foirBaseSalaried
	case EmploymentSelfEmployed:
		base = foirBaseSelfEmployed
	default:
		base = foirBaseInformal
	}
	notes := []string{fmt.Sprintf("base FOIR for %s income is %.0f%%", employmentLabel(p.Employment), base)}
	foir := base

	if p.NetMonthlyIncome > highIncomeThreshold && p.Employment != EmploymentInformal {
		foir += highIncomeBonus
		notes = append(notes, fmt.Sprintf(
			"+%.0f points because net income above ₹%s/month is treated as a high earner who can safely carry more fixed obligation",
			highIncomeBonus, formatINR(highIncomeThreshold),
		))
	}
	if p.CreditScore != nil {
		if *p.CreditScore >= goodCreditScore {
			foir += goodCreditBonus
			notes = append(notes, fmt.Sprintf("+%.0f points for a CIBIL score of %d (>= %d)", goodCreditBonus, *p.CreditScore, goodCreditScore))
		} else if *p.CreditScore < badCreditScore {
			foir -= badCreditPenalty
			notes = append(notes, fmt.Sprintf("-%.0f points for a CIBIL score of %d (< %d)", badCreditPenalty, *p.CreditScore, badCreditScore))
		}
	}
	if p.PastBounces != nil && *p.PastBounces > 0 {
		penalty := float64(*p.PastBounces) * perBouncePenalty
		foir -= penalty
		notes = append(notes, fmt.Sprintf("-%.0f points for %d bounced payment(s) in the last 3 months", penalty, *p.PastBounces))
	}
	if p.EmergencySavingsMonths != nil && *p.EmergencySavingsMonths >= emergencySavingsMonthsThreshold {
		foir += emergencySavingsBonus
		notes = append(notes, fmt.Sprintf("+%.0f points for holding at least %.0f months of expenses in savings", emergencySavingsBonus, emergencySavingsMonthsThreshold))
	}

	if foir > foirMax {
		foir = foirMax
	}
	if foir < foirMin {
		foir = foirMin
	}
	return foirResult{Percent: foir, Notes: notes}
}

// bankFOIR is the looser, income-only ceiling a bank's own underwriting
// formula typically applies -- it does not see household expenses,
// bounce history discounts, or savings, so it is always used for the
// Sanction Limit, never the Safe Capacity Limit.
func bankFOIR(p Profile) float64 {
	switch p.Employment {
	case EmploymentSalaried:
		return bankFOIRSalaried
	case EmploymentSelfEmployed:
		return bankFOIRSelfEmployed
	default:
		return bankFOIRInformal
	}
}
