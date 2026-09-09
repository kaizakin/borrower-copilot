# Borrower Copilot

A loan self-assessment tool for borrowers in India. It answers "how much can I borrow?" but also separates a bank's likely sanction amount from what the borrower can safely repay. It gives an interest-rate range based on the person's risk profile, checks the loan against an income drop and rate rise, and recommends secured products to borrowers with assets but no credit history when that makes more sense. For borrowers who are not ready, it shows a "Path to Yes" with the changes that would help.

For the input, output, and reasoning behind each of the three sample borrowers, see [WALKTHROUGH.md](WALKTHROUGH.md).

The app is stateless. The browser sends the form payload back and forth through `web/handlers.go`, and there is no database.

## Running it

```
go run .
```

Open `http://localhost:8080`. The flow is a two-step HTMX wizard. Step 1 always shows nine core questions. Step 2 shows only the follow-up questions relevant to the borrower's employment type and requested amount. See `engine.AdditionalQuestions`. A "quick assessment" shortcut skips step 2 and returns a wider, clearly low-confidence range instead of guessing.

```
go test ./...
```

runs the rules-engine unit tests, including regression tests for Priya, Ravi, and Anita. See `engine/engine_test.go`.

## Project layout

- `engine/` contains the rules engine. Its functions do not depend on HTTP or templates. It holds FOIR calculations (`foir.go`), EMI and APR calculations (`calc.go`), rate bands (`rates.go`), verdict logic (`verdict.go`), adaptive questions (`questions.go`), and the Path to Yes changes (`pathtoyes.go`).
- `web/` contains the HTTP handlers, HTMX form parsing and rendering in `handlers.go` and `parse.go`, plus the templates.
- `main.go` registers the three routes and starts the server.
- `RULES.md` lists every engine constant and threshold, along with its reasoning and source. Each is either required by the spec or a judgment call.
- `WALKTHROUGH.md` contains the three persona walkthroughs with real engine output.
- `DEMO_SCRIPT.md` is a script for recording a demo video.

## What I would build next

- **Verified data instead of self-reported data.** Pull credit scores from a bureau API, income from bank-statement or UPI analysis or a payroll integration, and collateral value from a registry lookup. The current version trusts what the borrower enters.
- **Calibrate the constants in `RULES.md` with lender data.** Every FOIR percentage, rate band, and LTV is a judgment call, not a fetched or regulator-mandated number. The next step is to validate and periodically refresh them using actual approvals.
- **Product-specific pricing and eligibility.** The app currently uses one generic band for each employment type. It should account for subsidised government small-business schemes, education-loan terms, and vehicle-loan LTV rules. That would be especially useful for borrowers like Ravi.
- **`httptest`-based HTTP-layer tests.** The engine's pure functions have good coverage. `web/handlers.go` has only been smoke-tested manually with `curl`.
- **Regional-language support.** The product is aimed at borrowers across India, but the UI is English-only.
- **Save and resume, plus a proper negotiation-card export.** State currently exists only in the form payload sent between browser and server. The negotiation card uses the browser's print dialog instead of a generated PDF.

## What I cut for time

- HTTP-layer tests. The `engine` unit tests cover the same logic as pure functions.
- Generated PDF export. The app uses `window.print()` and a `@media print` stylesheet instead.
- A live rate or policy feed. `RULES.md` contains static constants instead.
- Product-specific pricing tables. The app uses one generic band for each employment type.
- Debounced client-side validation beyond HTML5 `required`/`number` attributes.
- A formal accessibility audit, beyond using semantic `<label>`/`<select>`/`<input>` elements.

## Assumptions and limitations

- **The FOIR percentages, rate bands, and LTV values in `RULES.md` are judgment calls.** RBI does not set one FOIR. The 40% to 50%, 60%, and 30% figures in the assignment brief are industry rules of thumb. The credit-score, bounced-payment, and high-income adjustments, along with the exact rate bands, were chosen for this assignment. Validate them with lender data before relying on the exact figures.
- **The specified APR formula can understate cost for a long reducing-balance loan.** The app implements `APR = ((Total Interest + Fees) / Principal) × (365 / Tenure in Days) × 100` exactly as requested. Since the outstanding balance drops each month, total interest is much lower than `rate × years × principal`. The formula can therefore return an APR below the nominal rate for multi-year loans, even with the 2% processing fee. `TestAPRIncludesProcessingFee` checks that a fee raises APR against the no-fee result, rather than assuming APR always exceeds the nominal rate.
- **Some persona inputs are estimates.** Ravi's cash income range, Anita's income range, Anita's app-loan repayments, and household expenses are converted to single figures in the sample profiles. A real product should ask borrowers for those numbers directly.
- **Nothing is verified.** Credit score, income, existing EMIs, and collateral value are self-reported. This is a self-assessment aid for preparing for a negotiation, not a replacement for a lender's underwriting.
