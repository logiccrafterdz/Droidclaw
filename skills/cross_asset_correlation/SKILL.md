---
name: cross_asset_correlation
description: Monitor rolling correlations between asset pairs (BTC-ETH, BTC-SOL, crypto-DXY, Gold-USD, etc.) and detect significant correlation changes or breakdowns. Identifies concentration risk and diversification opportunities. Called automatically every 2 hours. Used by opportunity_scorer and daily_outlook.
metadata: {"droidclaw":{"emoji":"🔗","category":"economic","autonomous":true}}
---

# Cross-Asset & Correlation Monitor

Track inter-asset correlations to detect regime shifts, concentration risks, and diversification signals.

## When to Use
- Called automatically every 2 hours by cron
- Called on-demand when opportunity_scorer needs correlation data

## Monitored Pairs

Core pairs to track:
- **Crypto internal**: BTC-ETH, BTC-SOL, BTC-BNB, BTC-XRP, ETH-SOL
- **Crypto vs Macro**: BTC-DXY (proxy via EURUSDT inverse), BTC-Gold (XAUUSD)
- **Traditional**: Gold-EUR, Gold-USD(DXY)

## Procedure

1. **Fetch Price Data**: Use `market_data` tool:
   - `candles` action for BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT, XRPUSDT
     - Interval `1d`, limit `30` (30-day rolling window)
   - `candles` for EURUSDT (as DXY inverse proxy)
   - For Gold: use `forex` action to get XAU rate (or read from latest scan)

2. **Compute Rolling Correlations**:
   For each pair (A, B):
   - Extract daily returns: r_A(i) and r_B(i)
   - Calculate Pearson correlation over the 30-day window:
     ```
     rho = sum((r_A - mean_A)(r_B - mean_B)) / (n * std_A * std_B)
     ```
   - Also calculate a shorter 7-day rolling correlation for detecting recent shifts

3. **Classify Correlations**:
   For each pair:
   - |rho| >= 0.7 → "high_correlation"
   - 0.4 <= |rho| < 0.7 → "moderate_correlation"
   - |rho| < 0.4 → "low_correlation"
   - rho < -0.3 → "negative_correlation" (potential hedge)

4. **Detect Correlation Changes**:
   - Read previous correlations from `storage`: `correlation/current_correlations.json`
   - For each pair, compare rho_30d_current vs rho_30d_previous:
     - If |delta_rho| > 0.2 → flag as "correlation_shift" event
     - If high→low (e.g., BTC-ETH drops from 0.9 to 0.5) → "correlation_breakdown" (significant event)
     - If low→high → "correlation_convergence"

5. **Build Cluster Groups**:
   Simple clustering based on correlations:
   - **High-beta crypto cluster**: Assets with |rho| >= 0.7 to BTC (typically ETH, SOL)
   - **Defensive/uncorrelated**: Assets with |rho| < 0.3 to BTC
   - **Inverse/hedge**: Assets with rho < -0.3 to crypto basket

6. **Save Results**: Use `storage` tool:
   Write to `correlation/current_correlations.json`:
   ```json
   {
     "timestamp": "2026-02-11T10:00:00Z",
     "pairs": {
       "BTC-ETH": {
         "rho_30d": 0.88,
         "rho_7d": 0.92,
         "classification": "high_correlation",
         "change_from_previous": -0.02,
         "event": null
       },
       "BTC-SOL": {
         "rho_30d": 0.72,
         "rho_7d": 0.55,
         "classification": "high_correlation",
         "change_from_previous": -0.15,
         "event": "correlation_weakening"
       }
     },
     "clusters": {
       "high_beta_crypto": ["ETH", "SOL", "BNB"],
       "uncorrelated": ["XAU"],
       "inverse": []
     },
     "concentration_risk": "high",
     "concentration_note": "BTC/ETH/SOL highly correlated - essentially one trade"
   }
   ```
   Append to `correlation/correlation_history.json`.

## How Other Skills Use This

- **opportunity_scorer**: -1 score if opportunity asset is in same high-correlation cluster as existing opportunities (concentration risk)
- **detect_opportunity**: Flag divergence signals when normally-correlated assets decouple
- **daily_outlook**: Include "Correlation Status" section with cluster summary and any shifts
- **summarize_opportunities**: Add "Correlation Risk" assessment

## Safety Rules
- Correlation does not imply causation - always note this in reports.
- Mark as "insufficient_data" if fewer than 14 data points available.
- All output is "for informational purposes only".
