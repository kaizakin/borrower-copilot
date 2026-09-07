# WALKTHROUGH.md — Borrower Copilot

How to run, the three persona run-throughs with actual engine output, what was cut for time, and the limitations of the hardcoded assumptions baked into `RULES.md`.

## Running it

```
go run .
```

Then open `http://localhost:8080`. `go test ./...` runs the rules-engine unit tests, including the three personas below as automated regression tests (`engine/engine_test.go`).

The flow is a two-tier HTMX wizard: tier 1 (the 9 core questions, `POST /additional`) always renders; tier 2 (`POST /assess`) adaptively shows only the additional questions that are still relevant to the borrower's employment type and requested amount (see `engine.AdditionalQuestions`). A "Get a quick assessment now" button on tier 1 skips straight to `/assess`, producing the wide, low-confidence range the spec calls for.

---

## Persona 1 — Priya (29, Bengaluru, salaried)

**Input:** salaried, ₹1,10,000/month net, car EMI ₹14,000, rent ₹28,000, CIBIL 780, wants ₹8,00,000 personal loan. *Tenure isn't part of the original narrative — 36 months is assumed as a typical personal-loan tenure, since neither the EMI nor the APR formula is computable without one (see RULES.md's non-goals note on this).*

**Output:**
- **Verdict: Borrow** — "your safe capacity of ₹12,52,340 comfortably covers the ₹8,00,000 you want"
- **Max amount:** Safe Capacity ₹12,52,340 · Sanction Limit ₹12,66,030 (recommend using the Safe Capacity Limit)
- **Fair rate:** 9.5%–11.0% (tightened for her 780 CIBIL), all-in APR **6.7%** on the ₹8L loan
- **EMI ceiling:** ₹41,000/month, driven by her FOIR (50% = 45% salaried base + 5 for good credit)
- **Stress tests:** both a 20% income drop and a 2-point rate rise still leave her ceiling above the ₹25,908–26,667 EMI this loan needs

This is the comfortable case: good credit, verifiable salary, existing obligations well within her means. The engine correctly does not route her anywhere else or flag low confidence (she answered her credit score).

## Persona 2 — Ravi (42, Mysuru, self-employed)

**Input:** self-employed kirana owner, cash income ₹40,000–80,000/month, ITR ₹4.2L/year, unencumbered shop worth ₹45,00,000, no credit score, wife earns ₹18,000, wants ₹15,00,000 for business/vehicle.

**Judgement calls made to build this Profile:**
- **Declared income:** used the ₹40k–80k midpoint (₹60,000) as `NetMonthlyIncome`, and modelled roughly 40% of it as unverifiable cash on top of his provable ITR base (`VariableIncomeShare = 0.4`) — his ITR of ₹4.2L/yr (₹35,000/month) is close to the "stable" 60% of ₹60,000. A real underwriter would likely anchor to the ITR figure alone; this copilot instead shows the borrower what his own claimed income supports, discounted for the part he can't prove.
- **Household expenses (₹25,000/month)** and **existing EMIs (₹0)** are not stated in the narrative and are estimation gaps — a real deployment would ask, not assume.
- Wife's ₹18,000 income modelled as a co-applicant.

**Output:**
- **Routed:** yes — "as a self-employed applicant without an established credit score, an unsecured business loan is unlikely to be sanctioned at a fair rate; securing it against your ₹45,00,000 asset (loan-against-property/gold) instead unlocks a far lower, safer rate." **Recommended product: secured.**
- **Verdict: Borrow less** — safe capacity ₹8,06,385 against the ₹15,00,000 asked
- **Fair rate:** 9.0%–11.0% secured band (not the 30%+ unsecured band he'd otherwise face with no credit score)
- **EMI ceiling:** ₹26,400/month
- **Stress tests:** at his requested ₹15L amount, both stress scenarios fail — the copilot is explicit that the full ask does not survive a 20% income drop or a 2-point rate rise, reinforcing "borrow less" rather than the full amount

This is the case the assignment specifically asked for: an asset-rich, credit-invisible borrower who should never be quoted an unsecured rate, and whose true safe amount is well below what he's asking for even once routed correctly.

## Persona 3 — Anita (35, Hubballi, informal)

**Input:** informal delivery/tailoring income ₹26,000–30,000/month, 2 kids, 3 outstanding app loans (₹35,000 principal at 30%+), 1 bounced EMI last month, wants ₹1,50,000 for an EV scooter.

**Judgement calls made to build this Profile:**
- **Income:** midpoint ₹28,000.
- **Existing EMIs (₹9,000/month):** the narrative gives a principal (₹35,000 across 3 app loans) and a rate ("30%+"), not a monthly repayment figure — ₹9,000 is an estimated combined instalment for short-tenure, high-cost app loans of that size. A real tool would ask for this directly rather than infer it.
- **Household expenses (₹16,000/month)** for her and 2 kids is likewise an estimation gap, not stated in the narrative.

**Output:**
- **Verdict: Don't borrow** — "you've missed a payment in the last 3 months and your safe EMI ceiling is only ₹0 — taking on more debt now risks another bounce and further credit damage"
- **EMI ceiling: ₹0/month.** Her FOIR bottoms out at the 20% floor (30% informal base − 10 for the bounce), and 20% of ₹28,000 doesn't even cover her existing ₹9,000/month app-loan burden.
- **Max amount:** Safe Capacity ₹0. Sanction Limit still shown (₹13,795, what a bank's looser formula alone would say) — precisely to make the point that a bank's number can look survivable while the borrower's actual budget says otherwise.
- **Fair rate:** 22%–46% (informal base 24–36%, widened further for no credit score and the bounce) — deliberately in the same range as the app loans she's already stuck in.
- **Confidence: high**, even though only one additional question (bounce history) was answered — see `RULES.md` rule 38: disclosing even one high-impact risk factor is enough to stop the engine from defaulting to the flat "only core answered" low-confidence fallback.

This is the "actively tell high-risk profiles not to borrow" case the assignment calls out by name, and it's the one persona where the Sanction Limit and Safe Capacity Limit genuinely diverge — showing why the tool insists on recommending the safe number, not the bank's.

---

## What got cut for time

- **No automated HTTP-layer tests** — the three personas are covered by `engine` unit tests (pure functions, easy to assert on) and were manually smoke-tested end-to-end through the real HTTP/HTMX routes via `curl` during development, but there's no `httptest`-based test suite for `web/handlers.go` itself.
- **No PDF export** for the negotiation card — it relies on the browser's native print/"Save as PDF" dialog (`window.print()` with a `@media print` stylesheet), not a generated PDF file.
- **No live rate/policy feed** — every rate band, FOIR number, and LTV is a static constant (see `RULES.md`). A production version would need these calibrated against, and periodically refreshed from, actual lender data.
- **No product-specific pricing tables** — e.g. subsidised government schemes for small business loans (which could matter a great deal for Ravi), education-loan-specific terms, or vehicle-loan-specific LTV rules are all out of scope; every non-salaried, non-secured request is priced through one generic band per employment type.
- **No inline client-side validation** beyond HTML5 `required`/`number` attributes — no debounced field-level error messages, no guardrails against a borrower typing an obviously wrong number (e.g. income of ₹1).
- **No accessibility audit** beyond using semantic `<label>`/`<select>`/`<input>` elements — no screen-reader testing was performed.
- **English only** — no regional-language support, despite the product targeting borrowers across India.

## Limitations of the hardcoded financial assumptions

- **Every FOIR percentage, rate band, and LTV in `RULES.md` is a judgement call, not a fetched or regulator-mandated number.** RBI does not itself fix a single FOIR; the 40–50%/60%/30% figures in the assignment brief are industry rules of thumb, and this engine's specific point values within those ranges (the credit-score/bounce/high-income adjustments, the exact rate bands) are reasonable but ultimately invented for this assignment. They should be treated as a starting calibration to be validated against real lender data before anyone relies on the exact numbers, not as ground truth.
- **The given APR formula can understate cost for long reducing-balance tenures.** `APR = ((Total Interest + Fees) / Principal) × (365 / Tenure in Days) × 100` was implemented exactly as specified. But because a reducing-balance loan's total interest is much smaller than `rate × years × principal` (the outstanding balance shrinks every month), this formula — which the spec ties to a 2% processing fee disclosure — actually produces an APR *below* the nominal rate over multi-year tenures (see `engine/engine_test.go`'s `TestAPRIncludesProcessingFee`, which had to be written to assert "a fee raises APR relative to no fee" rather than "APR exceeds the nominal rate," because the latter isn't true in general for this formula). It's implemented faithfully to the spec, but this is a real mathematical quirk worth flagging rather than quietly hiding.
- **All three personas required translating narrative ranges into single point figures** (Ravi's ₹40k–80k cash income, Anita's ₹26k–30k income and her app-loan repayment burden) — the judgement calls made are documented above and in `engine/engine_test.go`'s comments. A production tool should ask for these directly instead of a developer guessing on the borrower's behalf.
- **Nothing is verified.** Credit score, income, existing EMIs, and collateral value are all self-reported and untrusted by construction — this is a self-assessment aid to prepare for a negotiation, not a substitute for a lender's actual underwriting, and the copy on the page says as much.
