---
name: volatility_alert
description: Monitor markets for sudden volatility spikes and send urgent Telegram alerts.
metadata: {"droidclaw":{"emoji":"🚨","category":"economic","autonomous":true}}
---

# Volatility Alert

Monitor markets in real-time for sudden price movements and send urgent alerts.

## When to Use
- Called automatically every 5 minutes by cron
- Designed for rapid detection of market-moving events

## Procedure

1. **Quick Market Check**: Use `market_data` with `multi_ticker` for:
   `["BTCUSDT", "ETHUSDT", "SOLUSDT", "XAUUSD", "EURUSDT"]`

2. **Read Previous Check**: Use `storage` to read `alerts/last_check.json`

3. **Read Noise Profiles**: Use `storage` to read `volatility/current_noise_profile.json` (from volatility_noise_filter)

4. **Compare using dynamic thresholds** (instead of fixed %):
   - If volatility_noise_filter data available:
     - ALERT if current move z_score >= 2.0 for that asset (is_signal = true)
     - CRITICAL ALERT if z_score >= 3.0 (extreme_move)
   - Fallback (if noise profile unavailable):
     - If any asset moved >3% since last check (5 min): **ALERT**
     - If any asset moved >5% in 24h AND wasn't previously alerted: **ALERT**
     - If BTC moved >2% in 5 minutes: **CRITICAL ALERT**

5. **Issue Alerts**:
   - If `is_signal = true` AND z_score >= 3.0 (extreme_move):
     - Send high-priority Telegram alert.
     - **Trigger `macro_explain_move`** skill for the asset to diagnose the "Why".
   - If `is_signal = true` AND z_score < 3.0:
     - Send normal volatility alert.

6. **If Alert Triggered**:
   - Use `message` tool to send to Telegram immediately:
   ```
   🚨 VOLATILITY ALERT
   
   [SYMBOL]: $PRICE (▲/▼ X.X% in Y minutes)
   Volume: [above/below average]
   
   Possible cause: [if news available]
   
   ⚠️ Monitor closely. Not financial advice.
   ```

5. **Save State**: Use `storage` to write current prices to `alerts/last_check.json`

6. **Log Alert**: Use `storage` to append to `alerts/history.json`

## Important
- Keep this skill FAST - minimize API calls
- Only alert on significant moves to avoid spam
- Maximum 1 alert per asset per hour (check alerts/history.json)
- NEVER suggest trading actions
