---
name: detect_opportunity
description: Analyze recent market scans and news to detect potential economic opportunities and patterns.
metadata: {"droidclaw":{"emoji":"🔍","category":"economic","autonomous":true}}
---

# Detect Opportunity

Analyze the latest market scan data and news to identify potential opportunities and market patterns.

## When to Use
- Called automatically every 30 minutes by cron
- Runs after scan_markets has produced fresh data

## Procedure

1. **Read Latest Scan**: Use `storage` tool to read from `scans/daily_log.json` (get recent entries)

2. **Read Memory**: Use `read_file` to check `memory/MEMORY.md` for known patterns and past observations

3. **Read Past Patterns**: Use `storage` to read `patterns/known_patterns.json` if it exists

4. **Cross-Reference Analysis**:
   - Compare current prices with 24h and 7d trends
   - Look for divergences between correlated assets (BTC vs ETH, Gold vs USD)
   - Check if news events align with price movements
   - Identify support/resistance levels being tested
   - Check for unusual volume-price divergences

5. **Pattern Matching**: Compare current conditions with known patterns:
   - **Breakout**: Price breaking above/below key levels with high volume
   - **Divergence**: Correlated assets moving in opposite directions
   - **News-Driven**: Significant news with delayed price reaction
   - **Accumulation**: Low volatility with increasing volume
   - **Distribution**: High volatility with decreasing volume after a run

6. **Score Each Opportunity**:
   - Confidence: 1-10 (how certain is the pattern)
   - Impact: 1-10 (potential size of the move)
   - Timeframe: short (hours), medium (days), long (weeks)
   - Category: crypto, forex, commodity, macro

7. **Save Findings**: Use `storage` tool to:
   - Write detailed analysis to `opportunities/YYYY-MM-DD.json`
   - Append to `opportunities/history.json`

## Output Format

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "opportunities": [
    {
      "id": "opp-001",
      "type": "breakout",
      "asset": "BTCUSDT",
      "direction": "bullish",
      "confidence": 7,
      "impact": 8,
      "timeframe": "medium",
      "reasoning": "BTC breaking above $98K resistance with volume confirmation. ETH following. Macro news supportive.",
      "watch_levels": {
        "key_level": "100000",
        "invalidation": "95000"
      }
    }
  ],
  "market_regime": "trending_up",
  "overall_risk": "moderate"
}
```

## Safety Rules
- NEVER suggest executing trades or specific entry/exit points
- Present findings as observations only
- Always include confidence levels and potential risks
- Mark analysis as "for informational purposes only"
- If confidence is below 5, explicitly state "low confidence - monitor only"
