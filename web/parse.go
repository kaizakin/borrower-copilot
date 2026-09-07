package web

import (
	"net/http"
	"strconv"

	"github.com/kaizakin/borrower-copilot/engine"
)

func formFloat(r *http.Request, key string) float64 {
	v, _ := strconv.ParseFloat(r.FormValue(key), 64)
	return v
}

func formInt(r *http.Request, key string) int {
	v, _ := strconv.Atoi(r.FormValue(key))
	return v
}

// optFloat returns nil for a blank field -- a skipped question is unknown,
// never zero.
func optFloat(r *http.Request, key string) *float64 {
	raw := r.FormValue(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

func optInt(r *http.Request, key string) *int {
	raw := r.FormValue(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &v
}

// parseCore builds a Profile from the tier-1 (must-answer) fields only.
func parseCore(r *http.Request) engine.Profile {
	return engine.Profile{
		Purpose:           r.FormValue("purpose"),
		AmountWanted:      formFloat(r, "amount_wanted"),
		LoanType:          engine.LoanType(r.FormValue("loan_type")),
		TenureMonths:      formInt(r, "tenure_months"),
		NetMonthlyIncome:  formFloat(r, "net_monthly_income"),
		Employment:        engine.Employment(r.FormValue("employment")),
		ExistingEMIs:      formFloat(r, "existing_emis"),
		HouseholdExpenses: formFloat(r, "household_expenses"),
		Age:               formInt(r, "age"),
	}
}

// applyAdditional layers whichever additional (tier-2) fields were
// submitted onto an already-parsed core Profile. A blank field stays nil.
func applyAdditional(p *engine.Profile, r *http.Request) {
	p.CreditScore = optInt(r, "credit_score")
	p.PastBounces = optInt(r, "past_bounces")
	p.EmergencySavingsMonths = optFloat(r, "emergency_savings_months")
	p.CollateralValue = optFloat(r, "collateral_value")
	p.OfferedAPR = optFloat(r, "offered_apr")

	if share := optFloat(r, "variable_income_share"); share != nil {
		s := *share / 100
		p.VariableIncomeShare = &s
		stability := engine.IncomeVariable
		p.IncomeStability = &stability
	}

	if coIncome := optFloat(r, "co_applicant_income"); coIncome != nil && *coIncome > 0 {
		p.CoApplicantIncome = coIncome
		yes := true
		p.HasCoApplicant = &yes
	}
}
