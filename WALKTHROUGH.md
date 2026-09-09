# Persona walkthroughs

These examples show the inputs and results for Priya, Ravi, and Anita.

## Persona 1: Priya, 29, Bengaluru, salaried

**Input.** Salaried, ₹1,10,000/month net, car EMI ₹14,000, rent ₹28,000, CIBIL 780, seeking an ₹8,00,000 personal loan. *The original narrative does not specify a tenure. This uses 36 months, a typical personal-loan term, because the EMI and APR cannot be calculated without one. See the non-goals note in `RULES.md`.*

**Output:**
- **Verdict: Borrow.** "Your safe capacity of ₹12,52,340 comfortably covers the ₹8,00,000 you want."
- **Maximum amount.** Safe Capacity ₹12,52,340. Sanction Limit ₹12,66,030. The recommendation uses the Safe Capacity Limit.
- **Fair rate.** 9.5% to 11.0%, tightened for her 780 CIBIL score. The all-in APR on the ₹8L loan is **6.7%**.
- **EMI ceiling.** ₹41,000/month, based on a 50% FOIR. That is the 45% salaried base plus 5 points for good credit.
- **Stress tests.** A 20% income drop and a 2-point rate rise both leave her ceiling above the ₹25,908 to ₹26,667 EMI required for this loan.

This is a comfortable case. She has good credit, a verifiable salary, and manageable existing obligations. The engine does not route her to another product or mark the result as low confidence because she provided her credit score.

## Persona 2: Ravi, 42, Mysuru, self-employed

**Input.** Self-employed kirana owner, cash income of ₹40,000 to ₹80,000/month, ITR of ₹4.2L/year, unencumbered shop worth ₹45,00,000, no credit score, wife earns ₹18,000, seeking ₹15,00,000 for a business or vehicle.

**Judgment calls used for this profile.**
- **Declared income.** The ₹40k to ₹80k midpoint, ₹60,000, is used as `NetMonthlyIncome`. About 40% is treated as unverifiable cash above his provable ITR income (`VariableIncomeShare = 0.4`). His ₹4.2L/year ITR, or ₹35,000/month, is close to the stable 60% of ₹60,000. A real underwriter would probably rely on the ITR alone. This copilot instead shows what his claimed income could support after discounting the unverified share.
- **Household expenses of ₹25,000/month** and **existing EMIs of ₹0** are not in the narrative, so they are estimates. A real product would ask for them.
- His wife's ₹18,000 income is modeled as co-applicant income.

**Output:**
- **Routed.** Yes. "As a self-employed applicant without an established credit score, an unsecured business loan is unlikely to be sanctioned at a fair rate. Securing it against your ₹45,00,000 asset, through a loan against property or gold, unlocks a much lower rate." **Recommended product: secured.**
- **Verdict: Borrow less.** Safe capacity is ₹8,06,385, against the ₹15,00,000 requested.
- **Fair rate.** 9.0% to 11.0% for the secured product, instead of the 30%+ unsecured band he would likely face without a credit score.
- **EMI ceiling.** ₹26,400/month.
- **Stress tests.** At the requested ₹15L, both scenarios fail. The full amount does not hold up to a 20% income drop or a 2-point rate rise, which supports the "borrow less" recommendation.

This is the asset-rich borrower with no credit history that the assignment asks for. He should not be quoted an unsecured rate, and his safe amount is well below what he requested even after the correct routing.

## Persona 3: Anita, 35, Hubballi, informal

**Input.** Informal delivery and tailoring income of ₹26,000 to ₹30,000/month, two children, three outstanding app loans with ₹35,000 principal at 30%+, one bounced EMI last month, seeking ₹1,50,000 for an EV scooter.

**Judgment calls used for this profile.**
- **Income.** The midpoint is ₹28,000.
- **Existing EMIs of ₹9,000/month.** The narrative gives principal, ₹35,000 across three app loans, and a rate of 30%+, but not a monthly repayment. ₹9,000 is an estimated combined instalment for short-term, high-cost loans of that size. A real tool would ask directly instead of inferring it.
- **Household expenses of ₹16,000/month** for her and two children are also an estimate, not part of the narrative.

**Output:**
- **Verdict: Don't borrow.** "You've missed a payment in the last 3 months and your safe EMI ceiling is only ₹0. Taking on more debt now risks another bounce and further credit damage."
- **EMI ceiling: ₹0/month.** Her FOIR falls to its 20% floor, the 30% informal base less 10 points for the bounce. Twenty percent of ₹28,000 does not cover her existing ₹9,000/month app-loan payments.
- **Maximum amount.** Safe Capacity is ₹0. The Sanction Limit is still shown as ₹13,795, which is what a bank's looser formula would say. This shows how a bank's number can look manageable while the borrower's actual budget says otherwise.
- **Fair rate.** 22% to 46%. The informal base is 24% to 36%, widened for no credit score and the bounce. It is in the same range as the app loans she already has.
- **Confidence: high.** Only one additional question, bounce history, was answered. Under rule 38 in `RULES.md`, one disclosed high-impact risk factor keeps the engine from using its flat low-confidence fallback for core answers only.

This is the "actively tell high-risk profiles not to borrow" case named in the assignment. It is also the only persona where the Sanction Limit and Safe Capacity Limit meaningfully diverge. That is why the tool recommends the safe number, not the bank's number.
