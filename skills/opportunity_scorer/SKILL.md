---
name: opportunity_scorer
description: Central scoring engine that takes raw opportunity signals and scores them using inputs from all analytical skills (trend_regime_filter, volatility_noise_filter, cross_asset_correlation, event_macro_trigger, trend_carry_regime). Produces a composite confidence score 1-10, filters low-quality signals, and outputs only the top 3 opportunities. Called after detect_opportunity produces raw signals. Replaces ad-hoc confidence scoring with a systematic, rule-based approach.
metadata: {"droidclaw":{"emoji":"🎯","category":"economic","autonomous":true}}
---

# Opportunity Scorer & Filter

Systematic scoring of raw opportunity signals using quantitative inputs from all analytical skills.

## When to Use
- Called automatically after detect_opportunity produces raw signals
- Called on-demand via "/score" command
- Called before summarize_opportunities to ensure only scored opportunities are reported

## Procedure

1. **Read Raw Opportunities**: Use `storage` to read latest from `opportunities/` directory

2. **Read Analytical Inputs**: Use `storage` to read:
   - `regime/current_regime.json` (from trend_regime_filter)
   - `volatility/current_noise_profile.json` (from volatility_noise_filter)
   - `correlation/current_correlations.json` (from cross_asset_correlation)
   - `macro/current_triggers.json` (from event_macro_trigger)
   - `carry/current_carry_regime.json` (from trend_carry_regime)

3. **Score Each Opportunity**:

   Read current weights from `scoring/weights.json`. Let weights be `w_trend`, `w_vol`, `w_corr`, `w_macro`, `w_carry`.

   Base score = 0.

   **a) Trend & Regime Score** (Component Score `s_trend`):
   - Align with strong trend → 2
   - Align with medium trend → 1
   - Conflict with regime → -2
   - Weighted Contribution = `s_trend * w_trend`

   **b) Volatility & Signal Quality** (Component Score `s_vol`):
   - z_score >= 2.0 → 2
   - z_score 1.0–2.0 → 1
   - Noise → -1
   - Weighted Contribution = `s_vol * w_vol`

   **c) Correlation Risk** (Component Score `s_corr`):
   - Low correlation/Divergence → 1
   - High concentration/Cluster → -1
   - Weighted Contribution = `s_corr * w_corr`

   **d) Macro Trigger Alignment** (Component Score `s_macro`):
   - Support from high-importance trigger → 2
   - Support from medium-importance trigger → 1
   - Contradiction → -2
   - Weighted Contribution = `s_macro * w_macro`

   **e) Trend-Carry Alignment** (Component Score `s_carry`):
   - Carry supported → 1
   - Carry opposed → -1
   - Weighted Contribution = `s_carry * w_carry`

4. **Compute Final Score**:
   ```
   raw_score = (s_trend * w_trend) + (s_vol * w_vol) + (s_corr * w_corr) + (s_macro * w_macro) + (s_carry * w_carry)
   # Normalize raw_score (which is roughly -2 to +2 weighted sum) to 1-10
   confidence = clamp(round((raw_score + 2) / 4 * 9 + 1), 1, 10)
   ```

5. **Apply Filters** (reject opportunities that don't pass):
   - REJECT if confidence < 4 (too weak)
   - REJECT if regime = "Range-bound" AND type = "trend continuation"
   - REJECT if is_signal = false AND no macro trigger (pure noise)
   - REJECT if asset had same-direction opportunity rejected in last 24h

6. **Rank and Limit**:
   - Sort remaining opportunities by confidence DESC
   - Keep only **Top 3** opportunities
   - If fewer than 3 pass filters, that's fine - quality over quantity

7. **Generate Scored Output**:
   ```json
   {
     "timestamp": "2026-02-11T15:00:00Z",
     "scored_opportunities": [
       {
         "id": "opp-001",
         "asset": "BTCUSDT",
         "type": "breakout",
         "direction": "bullish",
         "confidence": 8,
         "impact": 7,
         "score_breakdown": {
           "trend_regime": "+2 (strong uptrend confirmed)",
           "volatility_signal": "+2 (z_score 2.5, significant move)",
           "correlation_risk": "+1 (uncorrelated to other opps)",
           "macro_trigger": "+1 (dovish Fed minutes supportive)",
           "carry_alignment": "+1 (carry-supported trend)"
         },
         "raw_score": 7,
         "filters_passed": true,
         "reasoning": "Strong uptrend confirmed by MA crossover, significant breakout above noise threshold, supported by macro dovish shift and carry alignment.",
         "risk_factors": ["High correlation with ETH (same cluster)", "CPI release in 3 days"]
       }
     ],
     "rejected": [
       {
         "id": "opp-003",
         "asset": "XRPUSDT",
         "reason": "confidence 3 (below threshold), regime Range-bound"
       }
     ],
     "total_raw": 5,
     "total_scored": 3,
     "total_rejected": 2
   }
   ```

8. **Save Results**: Use `storage` tool:
   - Write to `opportunities/scored_YYYY-MM-DD.json`
   - Update `opportunities/active_scored.json` with current top opportunities

## How Other Skills Use This

- **summarize_opportunities**: Read scored output instead of raw opportunities
- **daily_outlook**: Include only scored/filtered opportunities in report
- **post_mortem**: Track scored confidence vs actual outcome for calibration
- **volatility_alert**: If extreme move on a scored opportunity asset, flag urgently

## Score Calibration Notes
- Track whether confidence correlates with actual outcomes (via post_mortem)
- If hit rate at confidence 8+ is < 50%, consider adjusting weights
- Component weights can be manually tuned - current version uses equal weighting

## Safety Rules
- Scoring is systematic, not predictive. High score means "multiple factors align", not "guaranteed move"
- Always present score breakdown so reasoning is transparent
- All output is "for informational purposes only"
- Maximum 3 opportunities displayed to avoid information overload
