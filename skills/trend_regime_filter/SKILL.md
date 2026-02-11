---
name: trend_regime_filter
description: Determine trend direction and market regime (trending-up, trending-down, range-bound) for each monitored asset using moving averages, Donchian channels, and ADX-like momentum filters. Called automatically every 30 minutes after scan_markets. Used as a prerequisite filter by detect_opportunity, opportunity_scorer, and trend_carry_regime skills.
metadata: {"droidclaw":{"emoji":"📈","category":"economic","autonomous":true}}
---

# Trend & Regime Filter

Classify the current market regime for every monitored asset so that downstream skills can filter signals accordingly.

## When to Use
- Called automatically every 30 minutes by cron (after scan_markets)
- Called on-demand before detect_opportunity or opportunity_scorer need regime data

## Procedure

1. **Fetch Candle Data**: For each asset in the watchlist use `market_data` tool:
   - `candles` action, symbol = each of: BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT, XRPUSDT
   - Interval `1d`, limit `50` (need ~50 daily candles for MA50 and Donchian-20)
   - Also fetch `4h` candles, limit `50` for short-term regime

2. **Compute Trend Indicators** (LLM calculates from candle close prices):

   **a) Moving Average Crossover (MA50 / MA200 proxy)**
   - MA_fast = average of last 20 closes (on daily) 
   - MA_slow = average of last 50 closes (on daily)
   - If MA_fast > MA_slow AND current_price > MA_fast → trend_direction = "up"
   - If MA_fast < MA_slow AND current_price < MA_fast → trend_direction = "down"
   - Otherwise → trend_direction = "range"

   **b) Donchian Channel (20-period)**
   - DC_high = max(high of last 20 candles)
   - DC_low = min(low of last 20 candles)
   - DC_mid = (DC_high + DC_low) / 2
   - If price > DC_mid AND price near DC_high (within 20%) → confirms uptrend
   - If price < DC_mid AND price near DC_low (within 20%) → confirms downtrend

   **c) Momentum Strength (ADX proxy)**
   - Calculate ATR(14) = average of true_range over 14 periods
   - Calculate directional movement: count of up-closes vs down-closes in last 14 periods
   - If ratio > 0.7 (mostly one direction) AND ATR is expanding → trend_strength = "strong"
   - If ratio 0.55–0.70 → trend_strength = "medium"
   - If ratio < 0.55 → trend_strength = "weak" (choppy/range)

3. **Combine into Regime Tag**:
   For each asset, produce:
   ```
   regime = {
     "trend_direction": "up" | "down" | "range",
     "trend_strength": "strong" | "medium" | "weak",
     "regime_tag": "Trend-up" | "Trend-down" | "Range-bound",
     "ma_fast": <value>,
     "ma_slow": <value>,
     "dc_high": <value>,
     "dc_low": <value>,
     "atr_14": <value>
   }
   ```

   Rules:
   - If trend_direction = "up" AND trend_strength >= "medium" → regime_tag = "Trend-up"
   - If trend_direction = "down" AND trend_strength >= "medium" → regime_tag = "Trend-down"
   - Otherwise → regime_tag = "Range-bound"

4. **Save Results**: Use `storage` tool to write to `regime/current_regime.json`:
   ```json
   {
     "timestamp": "2026-02-11T10:00:00Z",
     "assets": {
       "BTCUSDT": {
         "trend_direction": "up",
         "trend_strength": "strong",
         "regime_tag": "Trend-up",
         "ma_fast": 97500,
         "ma_slow": 92000,
         "dc_high": 100000,
         "dc_low": 88000,
         "atr_14": 2500,
         "price": 98000
       }
     }
   }
   ```
   Also append a snapshot to `regime/regime_history.json` for post-mortem tracking.

5. **Read Previous Regime**: Use `storage` to read previous `regime/current_regime.json`.
   - If any asset's regime_tag **changed** (e.g. Trend-up → Range-bound), flag as "regime_change" event.
   - Save regime changes to `regime/regime_changes.json`.

## How Other Skills Use This

- **detect_opportunity**: Do NOT generate "trend continuation" signals if regime = "Range-bound"
- **opportunity_scorer**: +1 score if regime confirms signal direction, -1 if contradicts
- **trend_carry_regime**: Uses trend_direction and trend_strength as inputs
- **daily_outlook**: Include regime summary per asset in the report

## Safety Rules
- This skill only observes and classifies. No trade suggestions.
- Always include "for informational purposes only" disclaimer.
- If data is insufficient (< 20 candles), mark regime as "insufficient_data" instead of guessing.
