// Package engine is the Borrower Copilot's domain layer: a pure,
// deterministic function from a borrower's self-reported Profile to an
// Assessment. It has no knowledge of HTTP, HTML, or storage.
package engine

// Employment classifies how a borrower earns income. It drives the base
// FOIR (Fixed Obligation to Income Ratio) ceiling because income
// verifiability and stability differ sharply across these groups.
type Employment string

const (
	EmploymentSalaried     Employment = "salaried"
	EmploymentSelfEmployed Employment = "self_employed"
	EmploymentInformal     Employment = "informal"
)

// LoanType is what the borrower says they want. LoanSecured covers
// loan-against-property, gold, or vehicle-lien products; everything else is
// unsecured until the engine decides otherwise (see Assessment.Routed).
type LoanType string

const (
	LoanPersonal LoanType = "personal"
	LoanBusiness LoanType = "business"
	LoanVehicle  LoanType = "vehicle"
	LoanHome     LoanType = "home"
	LoanSecured  LoanType = "secured"
)

// IncomeStability describes how predictable the borrower's monthly cash
// flow is. Only asked of non-salaried borrowers -- a salaried income is
// definitionally stable already.
type IncomeStability string

const (
	IncomeStable   IncomeStability = "stable"
	IncomeSeasonal IncomeStability = "seasonal"
	IncomeVariable IncomeStability = "variable"
)

// Profile is the borrower's self-reported state.
//
// Core fields are plain values: they are always asked, so they are always
// known. Additional fields are pointers: nil means "unknown", and must
// never be treated as zero or as an average. A borrower who skips their
// credit score is not the same as one with a credit score of 0 -- the
// engine widens confidence bands for the former and must never narrow them
// by guessing the latter.
type Profile struct {
	// --- Must-answer core ---
	Purpose           string
	AmountWanted      float64
	LoanType          LoanType
	TenureMonths      int
	NetMonthlyIncome  float64
	Employment        Employment
	ExistingEMIs      float64
	HouseholdExpenses float64
	Age               int

	// --- Additional (nullable; each one, if answered, must move a number) ---
	CreditScore            *int
	IncomeStability        *IncomeStability
	VariableIncomeShare    *float64 // 0..1, share of income that is variable/cash
	PastBounces            *int     // bounced EMIs/payments in the last 3 months
	EmergencySavingsMonths *float64 // months of expenses held in savings
	CollateralValue        *float64 // market value of an asset offered as security
	HasCoApplicant         *bool
	CoApplicantIncome      *float64
	OfferedAPR             *float64 // the rate a lender has already quoted them
}

// Verdict is the engine's headline recommendation.
type Verdict string

const (
	VerdictBorrow     Verdict = "Borrow"
	VerdictBorrowLess Verdict = "Borrow less"
	VerdictDontBorrow Verdict = "Don't borrow"
)

// Confidence flags whether the assessment rests on core answers alone.
type Confidence string

const (
	ConfidenceLow  Confidence = "low"
	ConfidenceHigh Confidence = "high"
)

type VerdictResult struct {
	Verdict       Verdict
	Justification string
}

type MaxAmountResult struct {
	SanctionLimit             float64
	SanctionJustification     string
	SafeCapacityLimit         float64
	SafeCapacityJustification string
	Recommended               string // which limit to actually use
	RecommendationReason      string
}

type InterestRateResult struct {
	LowPercent        float64
	HighPercent       float64
	BandJustification string
	AllInAPR          float64
	APRJustification  string
	Wide              bool // widened because of unknown inputs (never narrowed by guessing)
}

// StressScenario recomputes the EMI ceiling under a shock and states
// whether the requested loan would still survive it.
type StressScenario struct {
	Description     string
	NewCeiling      float64
	StillAffordable bool
	Note            string
}

type EMICeilingResult struct {
	Ceiling             float64
	Justification       string
	StressIncomeDrop20  StressScenario
	StressRateUp2       StressScenario
}

// Assessment is the engine's complete, deterministic output for a Profile.
type Assessment struct {
	Verdict       VerdictResult
	MaxAmount     MaxAmountResult
	InterestRate  InterestRateResult
	EMICeiling    EMICeilingResult
	Confidence    Confidence
	ConfidenceNote string

	// Routed is true when the requested (usually unsecured) product isn't
	// realistically obtainable and the engine priced/recommends a safer one.
	Routed             bool
	RecommendedProduct LoanType
	RoutingReason      string

	NegotiationPoints []string
}
