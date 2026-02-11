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
   `["BTCUSDT", "ETHUSDT", "SOLUSDT", "XAUUSDT", "EURUSDT"]`

2. **Read Previous Check**: Use `storage` to read `alerts/last_check.json`

3. **Compare**:
   - If any asset moved >3% since last check (5 min): **ALERT**
   - If any asset moved >5% in 24h AND wasn't previously alerted: **ALERT**
   - If BTC moved >2% in 5 minutes: **CRITICAL ALERT**

4. **If Alert Triggered**:
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
