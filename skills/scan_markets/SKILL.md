---
name: scan_markets
description: Scan major markets (crypto, forex, commodities) for price changes, volatility, and significant movements.
metadata: {"droidclaw":{"emoji":"📊","category":"economic","autonomous":true}}
---

# Scan Markets

Perform a comprehensive scan of major markets to detect significant price movements and volatility.

## When to Use
- Called automatically every 15 minutes by cron
- Can also be triggered manually via "/scan" command

## Procedure

1. **Get Multi-Ticker Data**: Use the `market_data` tool with action `multi_ticker` and symbols:
   `["BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT", "XAUUSD", "EURUSDT", "GBPUSDT"]`

2. **Get Forex Rates**: Use `market_data` with action `forex` and symbol `USD`

3. **Get Crypto News Headlines**: Use `news_feed` with action `crypto_news` and limit 5

4. **Get General Financial Headlines**: Use `news_feed` with action `headlines` and limit 5

5. **Analyze the Data**:
   - Identify any asset with >2% change in 24h (moderate signal)
   - Identify any asset with >5% change in 24h (strong signal)
   - Identify any asset with >10% change in 24h (critical signal)
   - Note any correlation between news and price movements
   - Track volume anomalies (unusually high volume)

6. **Save Scan Results**: Use `storage` tool to write results:
   - Write to `scans/YYYY-MM-DD_HH-MM.json` with the full scan data
   - Append summary to `scans/daily_log.json`

7. **Generate Summary**: Create a brief text summary with:
   - Market status (calm / volatile / trending)
   - Notable movements with percentage changes
   - Key news that may affect markets
   - Confidence level (1-10) for any detected patterns

## Output Format

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "market_status": "volatile",
  "signals": [
    {
      "symbol": "BTCUSDT",
      "price": "97500.00",
      "change_24h": "+3.5%",
      "signal_strength": "moderate",
      "note": "Breaking above resistance"
    }
  ],
  "news_impact": [
    "Fed signals potential rate pause - bullish for crypto"
  ],
  "overall_sentiment": "cautiously_bullish",
  "confidence": 7
}
```

## Safety Rules
- NEVER suggest executing trades
- NEVER provide financial advice
- Only observe, analyze, and report
- Mark all analysis as "for informational purposes only"
