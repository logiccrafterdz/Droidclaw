---
name: trend_carry_regime
description: Combine trend signals with carry/yield proxies (funding rates, interest rate differentials) to classify assets into regime labels like carry-supported-trend, pure-momentum, or mean-reversion. Primarily useful for FX pairs and crypto funding. Called automatically every 4 hours. Used by opportunity_scorer to boost confidence when trend and carry align.
metadata: {"droidclaw":{"emoji":"💹","category":"economic","autonomous":true}}
---

# Trend-Carry & Regime Score

Assess whether trend direction is supported by carry/yield, creating a more robust regime classification.

## When to Use
- Called automatically every 4 hours by cron
- Called on-demand when opportunity_scorer needs carry data

## Concept

Research shows that combining trend-following with carry (yield differentials) produces more stable returns than either alone. When trend and carry point the same direction, confidence should be higher. When they diverge, the setup is riskier.

## Procedure

1. **Get Trend Data**: Read from `storage`: `regime/current_regime.json` (produced by trend_regime_filter)

2. **Compute Carry Proxies**:

   **a) For Crypto (Funding Rate proxy)**:
   - Use `market_data` tool: `candles` for BTCUSDT, ETHUSDT, SOLUSDT with interval `1h`, limit `24`
   - Calculate the "basis" between spot price moves and short-term momentum as funding proxy
   - Positive funding → longs paying shorts → market is overleveraged long
   - Negative funding → shorts paying longs → market is overleveraged short
   - Approximation: if price has been rising steadily (>5% in 7d) with high volume, funding likely positive
   - Read any funding data from `patterns/carry_data.json` if manually maintained

   **b) For FX pairs**:
   - Use `market_data` with `forex` action for USD rates
   - Interest rate differentials (use known approximate rates):
     - USD: ~5.25%, EUR: ~4.50%, GBP: ~5.25%, JPY: ~0.10%, CHF: ~1.75%
     - (These should be updated when central bank decisions are detected by event_macro_trigger)
   - Carry direction for EURUSD: if USD rate > EUR rate → carry favors short EURUSD (long USD)
   - Read latest rate estimates from `patterns/interest_rates.json` if available

   **c) For Gold**:
   - Gold has negative carry (storage cost, no yield)
   - Real yield proxy: if USD rates high and inflation expectations low → negative for gold (carry drag)
   - If real yields declining → gold carry improves (less opportunity cost)

3. **Combine Trend + Carry**:
   For each asset:
   ```
   trend_score = +1 (up), 0 (range), -1 (down)  [from trend_regime_filter]
   carry_score = +1 (carry supports direction), 0 (neutral), -1 (carry opposes)
   
   combined = trend_score + carry_score
   ```
   
   Classification:
   - combined = +2 or -2 → "carry_supported_trend" (strongest regime)
   - combined = +1 or -1 → "pure_momentum" (trend without carry support)
   - combined = 0, trend != range → "carry_opposed_trend" (trend fighting carry - risky)
   - trend = range → "mean_reversion_regime" (no trend, play ranges)

4. **Generate Regime Labels**:
   ```json
   {
     "timestamp": "2026-02-11T14:00:00Z",
     "assets": {
       "BTCUSDT": {
         "trend_score": 1,
         "carry_score": 0,
         "combined_score": 1,
         "regime_label": "pure_momentum",
         "carry_note": "Funding rate proxy neutral, no strong carry signal"
       },
       "EURUSDT": {
         "trend_score": -1,
         "carry_score": -1,
         "combined_score": -2,
         "regime_label": "carry_supported_trend",
         "carry_note": "USD yield advantage + EUR downtrend = aligned"
       }
     }
   }
   ```

5. **Save Results**: Use `storage` tool:
   - Write to `carry/current_carry_regime.json`
   - Append snapshot to `carry/carry_history.json`

## How Other Skills Use This

- **opportunity_scorer**: +1 bonus if regime_label = "carry_supported_trend" in the opportunity's direction
- **daily_outlook**: Include "Trend-Carry Alignment" section for FX and major assets
- **detect_opportunity**: Prefer carry-aligned trend setups over pure momentum

## Data Maintenance

- Interest rate estimates in `patterns/interest_rates.json` should be updated when event_macro_trigger detects rate decisions
- Funding rate approximations are rough - flag if not based on actual API data

## Safety Rules
- Carry proxies used here are approximations, not real-time funding rates
- Always disclose the approximation level in reports
- All output is "for informational purposes only"
