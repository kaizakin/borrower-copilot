package engine

import "fmt"

// rateBand returns the fair annual interest-rate band [low, high] for a
// profile, whether the band was widened for missing information, and the
// reasoning behind each adjustment. An unknown input always pushes the band
// up and wider -- it is never treated as average risk.
func rateBand(p Profile, secured bool) (low, high float64, wide bool, notes []string) {
	if secured {
		low, high = 9.0, 11.0
		notes = append(notes, "secured against collateral, so priced near secured-loan rates (9%-11%) regardless of employment type or credit history")
		return
	}

	switch p.Employment {
	case EmploymentSalaried:
		low, high = 11.0, 13.0
		notes = append(notes, "salaried unsecured base band is 11%-13%")
	case EmploymentSelfEmployed:
		low, high = 14.0, 18.0
		notes = append(notes, "self-employed unsecured base band is 14%-18%, loaded for harder-to-verify income")
	default:
		low, high = 24.0, 36.0
		notes = append(notes, "informal-income unsecured lending sits in the 24%-36% band -- the same territory as high-cost app loans")
	}

	if p.CreditScore == nil {
		wide = true
		low -= 2
		high += 8
		notes = append(notes, "widened and pushed up because no credit score was provided -- unknown risk is priced as high risk, never as average risk")
	} else if *p.CreditScore >= goodCreditScore {
		low -= 1.5
		high -= 2
		notes = append(notes, fmt.Sprintf("tightened toward the low end for a CIBIL score of %d", *p.CreditScore))
	} else if *p.CreditScore < badCreditScore {
		low += 3
		high += 6
		wide = true
		notes = append(notes, fmt.Sprintf("pushed up and widened for a sub-%d CIBIL score of %d", badCreditScore, *p.CreditScore))
	}

	if p.PastBounces != nil && *p.PastBounces > 0 {
		high += float64(*p.PastBounces) * 2
		wide = true
		notes = append(notes, fmt.Sprintf("widened further for %d recent bounced payment(s)", *p.PastBounces))
	}

	if low < 8 {
		low = 8
	}
	if high < low {
		high = low
	}
	return
}
