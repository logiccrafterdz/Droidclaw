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

2. **Read Analytical Skill Outputs** (prerequisite data from quantitative skills):
   - `regime/current_regime.json` (from trend_regime_filter) — check regime tags
   - `volatility/current_noise_profile.json` (from volatility_noise_filter) — check signal vs noise
   - `correlation/current_correlations.json` (from cross_asset_correlation) — check divergences
   - `macro/current_triggers.json` (from event_macro_trigger) — check active triggers

3. **Read Memory**: Use `read_file` to check `memory/MEMORY.md` for known patterns and past observations

4. **Read Past Patterns**: Use `storage` to read `patterns/known_patterns.json` if it exists

5. **Pre-Filter with Regime & Noise Data**:
   - Do NOT generate "trend continuation" opportunities for assets where regime = "Range-bound"
   - Do NOT generate opportunities where the current move is classified as "normal_move" (z_score < 1.0) unless there is a strong macro trigger
   - DO flag "divergence" opportunities when cross_asset_correlation shows correlation_breakdown

6. **Cross-Reference Analysis**:
   - Compare current prices with 24h and 7d trends
   - Look for divergences between correlated assets (use correlation data, not just BTC vs ETH)
   - Check if news events align with price movements (use macro triggers)
   - Identify support/resistance levels being tested (use Donchian channels from regime data)
   - Check for unusual volume-price divergences

7. **Pattern Matching**: Compare current conditions with known patterns:
   - **Breakout**: Price breaking above/below key levels with high volume AND regime supports direction
   - **Divergence**: Correlated assets moving in opposite directions (confirmed by correlation monitor)
   - **News-Driven**: Significant news with delayed price reaction (confirmed by macro trigger)
   - **Accumulation**: Low volatility with increasing volume (vol regime = "low")
   - **Distribution**: High volatility with decreasing volume after a run

8. **Generate Raw Opportunities** (preliminary scoring only):
   - Confidence: 1-10 (preliminary - will be refined by opportunity_scorer)
   - Impact: 1-10 (potential size of the move)
   - Timeframe: short (hours), medium (days), long (weeks)
   - Category: crypto, forex, commodity, macro
   - Include which analytical signals support/contradict

9. **Save Raw Findings**: Use `storage` tool to:
   - Write raw analysis to `opportunities/raw_YYYY-MM-DD.json`
   - These will be picked up by **opportunity_scorer** for systematic scoring and filtering
   - Also append to `opportunities/history.json`

10. **Trigger Opportunity Scorer**: After saving raw opportunities, the opportunity_scorer skill should be invoked to apply systematic scoring and produce the final filtered list.

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
