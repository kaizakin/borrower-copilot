package engine

// QuestionKind is the HTML input shape a Question renders as.
type QuestionKind string

const (
	KindNumber QuestionKind = "number"
	KindSelect QuestionKind = "select"
	KindBool   QuestionKind = "bool"
)

type Option struct {
	Value, Label string
}

// Question describes one form field in the questionnaire, core or
// additional. ID matches the form field name the web layer parses.
type Question struct {
	ID      string
	Label   string
	Help    string
	Kind    QuestionKind
	Options []Option
	Unit    string
}

// CoreQuestions is the fixed, always-asked must-answer sequence (tier 1).
func CoreQuestions() []Question {
	return []Question{
		{ID: "purpose", Label: "What is this loan for?", Kind: KindSelect, Options: []Option{
			{"personal", "Personal need (wedding, medical, travel...)"},
			{"business", "Business / working capital"},
			{"vehicle", "Vehicle purchase"},
			{"home", "Home purchase or renovation"},
		}},
		{ID: "loan_type", Label: "What kind of loan are you looking at?", Kind: KindSelect, Options: []Option{
			{"personal", "Unsecured personal loan"},
			{"business", "Unsecured business loan"},
			{"vehicle", "Vehicle loan"},
			{"home", "Home loan"},
			{"secured", "Secured loan (against property/gold/vehicle)"},
		}},
		{ID: "amount_wanted", Label: "How much do you want to borrow?", Kind: KindNumber, Unit: "₹"},
		{ID: "tenure_months", Label: "Over how many months do you want to repay it?", Kind: KindNumber, Unit: "months"},
		{ID: "net_monthly_income", Label: "What is your net (take-home) monthly income?", Kind: KindNumber, Unit: "₹/month"},
		{ID: "employment", Label: "How do you earn your income?", Kind: KindSelect, Options: []Option{
			{"salaried", "Salaried"},
			{"self_employed", "Self-employed (business/professional with ITR)"},
			{"informal", "Informal / cash income (no ITR)"},
		}},
		{ID: "existing_emis", Label: "What do your existing EMIs add up to per month?", Kind: KindNumber, Unit: "₹/month"},
		{ID: "household_expenses", Label: "What are your household expenses per month (rent, food, school fees...)?", Kind: KindNumber, Unit: "₹/month"},
		{ID: "age", Label: "What is your age?", Kind: KindNumber, Unit: "years"},
	}
}

// AdditionalQuestions returns, in order, the additional questions still
// relevant given the core answers so far. Every question returned here
// must move at least one final output for this borrower -- if it wouldn't,
// it is skipped rather than asked pointlessly.
func AdditionalQuestions(p Profile) []Question {
	var qs []Question

	if p.Employment != EmploymentSalaried {
		// A salaried income is definitionally stable; this only moves the
		// number for self-employed/informal borrowers.
		qs = append(qs, Question{
			ID: "variable_income_share", Kind: KindNumber, Unit: "%",
			Label: "Roughly what share of your income is irregular/cash rather than fixed?",
			Help:  "We count irregular income at half value when working out what you can safely repay.",
		})
	}

	qs = append(qs, Question{
		ID: "credit_score", Kind: KindNumber,
		Label: "What is your CIBIL/credit score, if you know it?",
		Help:  "Leave blank if you don't know -- we'll price the unknown as higher risk rather than guessing an average.",
	})

	qs = append(qs, Question{
		ID: "past_bounces", Kind: KindNumber,
		Label: "Have any EMIs or payments bounced in the last 3 months? How many?",
	})

	qs = append(qs, Question{
		ID: "emergency_savings_months", Kind: KindNumber, Unit: "months",
		Label: "How many months of expenses do you have in savings?",
	})

	// A kirana-store owner doesn't need to be asked about collateral for a
	// small salaried-scale request; but any non-salaried borrower, or
	// anyone asking for a large amount, might have an asset worth offering.
	if p.Employment != EmploymentSalaried || p.AmountWanted >= 500000 {
		qs = append(qs, Question{
			ID: "collateral_value", Kind: KindNumber, Unit: "₹",
			Label: "Do you have property, gold, or a vehicle you could offer as security? What's it worth?",
			Help:  "Leave blank if none -- this can unlock a much lower secured-loan rate.",
		})
	}

	qs = append(qs, Question{
		ID: "co_applicant_income", Kind: KindNumber, Unit: "₹/month",
		Label: "If a co-applicant (spouse/parent) would join the loan, what is their monthly income?",
		Help:  "Leave blank if there's no co-applicant. Joining pools your incomes but makes you both liable for repayment.",
	})

	qs = append(qs, Question{
		ID: "offered_apr", Kind: KindNumber, Unit: "%",
		Label: "Has a lender already quoted you a rate? What was it?",
	})

	return qs
}
