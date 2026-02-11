---
name: volatility_noise_filter
description: Measure realized volatility and classify price moves as signal vs noise for each asset. Computes noise thresholds using standard deviation bands so that only statistically significant moves trigger alerts. Called automatically every 15 minutes after scan_markets. Used by volatility_alert, detect_opportunity, and opportunity_scorer as a pre-filter.
metadata: {"droidclaw":{"emoji":"📉","category":"economic","autonomous":true}}
---

# Volatility & Noise Filter

Quantify whether a price move is meaningful signal or normal noise, per asset.

## When to Use
- Called automatically every 15 minutes by cron (after scan_markets)
- Called by other skills before issuing alerts or scoring opportunities

## Procedure

1. **Fetch Historical Data**: For each asset in watchlist use `market_data` tool:
   - `candles` action, symbol = BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT, XRPUSDT
   - Interval `1h`, limit `100` (covers ~4 days of hourly data)
   - Also fetch `1d` candles, limit `30` for longer-term vol baseline

2. **Compute Realized Volatility**:

   **a) Hourly volatility (short-term)**
   - Calculate returns: r_i = (close_i - close_{i-1}) / close_{i-1} for each hourly candle
   - vol_1h = standard_deviation(last 24 hourly returns)
   - vol_4h = standard_deviation(returns aggregated to 4h blocks, last 7 days)

   **b) Daily volatility (baseline)**
   - Calculate daily returns from daily candles
   - vol_daily = standard_deviation(last 20 daily returns)
   - vol_7d = standard_deviation(last 7 daily returns)

   **c) Annualized estimate** (for reference):
   - annual_vol = vol_daily * sqrt(365) for crypto, sqrt(252) for forex

3. **Classify Current Move**:
   - Get current 24h price change (from latest scan or ticker)
   - Get current 1h price change
   - Calculate z_score_24h = abs(change_24h) / vol_daily
   - Calculate z_score_1h = abs(change_1h) / vol_1h

   Classification:
   - z_score < 1.0 → "normal_move" (within 1 sigma, likely noise)
   - z_score 1.0–2.0 → "notable_move" (unusual but not extreme)
   - z_score 2.0–3.0 → "volatility_spike" (significant signal)
   - z_score > 3.0 → "extreme_move" (rare event, high signal)

4. **Build Noise Profile**: For each asset:
   ```json
   {
     "symbol": "BTCUSDT",
     "vol_1h": 0.008,
     "vol_daily": 0.035,
     "vol_7d": 0.04,
     "noise_threshold_1h": 0.016,
     "noise_threshold_daily": 0.07,
     "current_1h_change": 0.012,
     "current_24h_change": 0.035,
     "z_score_1h": 1.5,
     "z_score_24h": 1.0,
     "classification_1h": "notable_move",
     "classification_24h": "normal_move",
     "is_signal": false
   }
   ```
   - `noise_threshold` = 2 * realized_vol (the 2-sigma boundary)
   - `is_signal` = true only if z_score >= 2.0 on any timeframe

5. **Save Results**: Use `storage` tool:
   - Write to `volatility/current_noise_profile.json` with all assets
   - Append snapshot to `volatility/vol_history.json`

6. **Save Persistent Noise Profile**: Use `storage` to update `patterns/noise_profiles.json`:
   - Store rolling average noise levels per asset (updated each run)
   - This builds the "normal range" baseline over time

## How Other Skills Use This

- **volatility_alert**: Only trigger alert if `is_signal = true` (z_score >= 2.0), replacing the fixed 3%/5% thresholds with asset-specific dynamic thresholds
- **detect_opportunity**: Ignore opportunities where the move is classified as "normal_move"
- **opportunity_scorer**: +1 if move exceeds noise threshold, 0 if within noise
- **daily_outlook**: Include "Volatility Status" section showing which assets have elevated vol

## Volatility Regime Detection
- If vol_7d > 1.5 * vol_daily(20d average) → "High volatility regime" (widen thresholds)
- If vol_7d < 0.5 * vol_daily(20d average) → "Low volatility regime" (tighten thresholds, breakout watch)
- Otherwise → "Normal volatility regime"

## Safety Rules
- This skill only measures and classifies. No trade suggestions.
- When data is insufficient (< 20 data points), mark as "insufficient_data".
- All output is "for informational purposes only".
