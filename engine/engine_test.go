package engine

import "testing"

func ptrF(f float64) *float64 { return &f }
func ptrI(i int) *int         { return &i }
func ptrB(b bool) *bool       { return &b }

// Priya (29, Bengaluru): salaried MNC, net ₹1.1L/mo, car EMI ₹14k, CIBIL
// 780, rent ₹28k, wants ₹8L personal loan. Tenure isn't part of the
// original narrative -- 36 months is assumed as a typical personal-loan
// tenure (documented in WALKTHROUGH.md).
func priya() Profile {
	return Profile{
		Purpose:           "personal",
		AmountWanted:      800000,
		LoanType:          LoanPersonal,
		TenureMonths:      36,
		NetMonthlyIncome:  110000,
		Employment:        EmploymentSalaried,
		ExistingEMIs:      14000,
		HouseholdExpenses: 28000,
		Age:               29,
		CreditScore:       ptrI(780),
	}
}

func TestPriya_ComfortableSalariedProfile(t *testing.T) {
	a := Assess(priya())

	if a.Verdict.Verdict != VerdictBorrow {
		t.Errorf("expected Borrow, got %s (%s)", a.Verdict.Verdict, a.Verdict.Justification)
	}
	if a.MaxAmount.SafeCapacityLimit < 800000 {
		t.Errorf("expected safe capacity to cover the ₹8L ask, got %.0f", a.MaxAmount.SafeCapacityLimit)
	}
	if a.InterestRate.Wide {
		t.Error("expected a tight (non-widened) band: credit score and history are known")
	}
	if a.InterestRate.HighPercent > 13 {
		t.Errorf("expected excellent-credit salaried rate near the low end of the band, got high=%.1f", a.InterestRate.HighPercent)
	}
	if a.Routed {
		t.Error("Priya should not be routed to a secured product")
	}
	if a.Confidence != ConfidenceHigh {
		t.Errorf("expected high confidence (credit score answered), got %s", a.Confidence)
	}
	if a.EMICeiling.Ceiling <= 0 {
		t.Errorf("expected a positive EMI ceiling, got %.0f", a.EMICeiling.Ceiling)
	}
}

// Ravi (42, Mysuru): self-employed kirana owner. Cash income ₹40k-80k/mo
// (midpoint ₹60k used as declared income), ITR ₹4.2L/yr implies roughly
// 42% of that is unverifiable cash on top of a provable base, modelled as
// VariableIncomeShare=0.4. Shop unencumbered worth ₹45L. No credit score.
// Wife earns ₹18k, modelled as a co-applicant. Wants ₹15L for
// business/vehicle. Household expenses (₹25k) and existing EMIs (₹0) are
// not given in the narrative and are estimation gaps documented in
// WALKTHROUGH.md.
func ravi() Profile {
	return Profile{
		Purpose:             "business",
		AmountWanted:        1500000,
		LoanType:            LoanBusiness,
		TenureMonths:        36,
		NetMonthlyIncome:    60000,
		Employment:          EmploymentSelfEmployed,
		ExistingEMIs:        0,
		HouseholdExpenses:   25000,
		Age:                 42,
		VariableIncomeShare: ptrF(0.4),
		CollateralValue:     ptrF(4500000),
		HasCoApplicant:      ptrB(true),
		CoApplicantIncome:   ptrF(18000),
	}
}

func TestRavi_SelfEmployedRoutedToSecured(t *testing.T) {
	a := Assess(ravi())

	if !a.Routed {
		t.Error("expected Ravi to be routed to a secured product given his unencumbered shop and no credit score")
	}
	if a.RecommendedProduct != LoanSecured {
		t.Errorf("expected recommended product to be secured, got %s", a.RecommendedProduct)
	}
	if a.InterestRate.HighPercent > 11 {
		t.Errorf("expected secured-band pricing (<=11%%), got high=%.1f", a.InterestRate.HighPercent)
	}
	if a.Verdict.Verdict == VerdictBorrow {
		t.Errorf("expected Ravi's ₹15L ask to exceed his safe capacity (Borrow less), got plain Borrow: safe=%.0f", a.MaxAmount.SafeCapacityLimit)
	}
	if a.MaxAmount.SafeCapacityLimit <= 0 {
		t.Error("expected some positive safe capacity given his collateral and pooled income")
	}
}

// Anita (35, Hubballi): informal delivery/tailoring income ₹26k-30k/mo
// (midpoint ₹28k). 3 outstanding app loans at 30%+ are modelled as a
// combined existing EMI of ₹9k/month; 1 bounced EMI last month. Household
// expenses for her and 2 kids assumed at ₹16k/month. Wants ₹1.5L for an EV
// scooter. Existing-EMI and household-expense figures are estimation gaps
// documented in WALKTHROUGH.md since the narrative gives ranges, not
// single numbers.
func anita() Profile {
	return Profile{
		Purpose:           "vehicle",
		AmountWanted:      150000,
		LoanType:          LoanVehicle,
		TenureMonths:      24,
		NetMonthlyIncome:  28000,
		Employment:        EmploymentInformal,
		ExistingEMIs:      9000,
		HouseholdExpenses: 16000,
		Age:               35,
		PastBounces:       ptrI(1),
	}
}

func TestAnita_HighRiskInformalTriggersDontBorrow(t *testing.T) {
	a := Assess(anita())

	if a.Verdict.Verdict != VerdictDontBorrow {
		t.Errorf("expected Don't borrow given a recent bounce and near-zero safe room, got %s (%s)", a.Verdict.Verdict, a.Verdict.Justification)
	}
	if a.EMICeiling.Ceiling != 0 {
		t.Errorf("expected EMI ceiling to floor at 0, got %.0f", a.EMICeiling.Ceiling)
	}
	if a.MaxAmount.SafeCapacityLimit != 0 {
		t.Errorf("expected zero safe capacity, got %.0f", a.MaxAmount.SafeCapacityLimit)
	}
	if a.Confidence != ConfidenceHigh {
		t.Error("expected high confidence: disclosing even one high-impact additional answer (past bounces) escapes the all-core-only low-confidence fallback")
	}
}

func TestAPRIncludesProcessingFee(t *testing.T) {
	// The RBI-style formula annualises total interest over the tenure, so
	// it is not guaranteed to exceed the nominal rate for a long,
	// reducing-balance tenure -- but adding a processing fee must always
	// push the all-in APR up relative to the same loan with no fee.
	principal := 100000.0
	rate := 12.0
	tenure := 12
	withFee := APR(principal, rate, 2.0, tenure)
	noFee := APR(principal, rate, 0.0, tenure)
	if withFee <= noFee {
		t.Errorf("APR with a processing fee (%.2f) should exceed APR without one (%.2f)", withFee, noFee)
	}
}

func TestEMIAndMaxPrincipalAreInverses(t *testing.T) {
	principal := 500000.0
	rate := 12.0
	tenure := 36
	emi := EMI(principal, rate, tenure)
	back := MaxPrincipal(emi, rate, tenure)
	diff := back - principal
	if diff < -1 || diff > 1 {
		t.Errorf("MaxPrincipal(EMI(P)) should round-trip to P, got %.2f vs %.2f", back, principal)
	}
}

func TestUnknownCreditScoreWidensNotAverages(t *testing.T) {
	known := priya()
	unknown := priya()
	unknown.CreditScore = nil

	aKnown := Assess(known)
	aUnknown := Assess(unknown)

	if !aUnknown.InterestRate.Wide {
		t.Error("expected an unknown credit score to widen the rate band")
	}
	if aUnknown.InterestRate.HighPercent <= aKnown.InterestRate.HighPercent {
		t.Error("expected the unknown-credit band to sit at or above the known-good-credit band, never narrower")
	}
	if aUnknown.Confidence != ConfidenceLow {
		t.Error("expected low confidence when all additional questions are unanswered")
	}
}
