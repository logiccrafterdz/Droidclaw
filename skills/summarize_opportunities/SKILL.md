---
name: summarize_opportunities
description: Generate a comprehensive summary of detected opportunities with detailed analysis for Telegram delivery.
metadata: {"droidclaw":{"emoji":"📋","category":"economic","autonomous":true}}
---

# Summarize Opportunities

Produce a well-structured text report of all currently tracked opportunities.

## When to Use
- Called on-demand via Telegram "/opportunities" command
- Called as part of evening summary flow

## Procedure

1. **Read Active Opportunities**: Use `storage` to list and read from `opportunities/` directory

2. **Read Patterns**: Use `storage` to read `patterns/known_patterns.json`

3. **Read Today's Scans**: Use `storage` to read `scans/daily_log.json`

4. **Compile Report**:

```
📋 Opportunity Report
📅 [Date] | ⏰ [Time]

━━━ Active Opportunities ━━━

🔵 #1: [Type] - [Asset]
   Direction: [Bullish/Bearish]
   Confidence: [X/10] ████████░░
   Impact: [X/10]
   Timeframe: [short/medium/long]
   Key Levels: [support] → [resistance]
   Reasoning: [brief explanation]
   Status: [new/developing/mature]

🔵 #2: [Type] - [Asset]
   ...

━━━ Pattern Reliability ━━━
✅ Confirmed: [X] patterns
❌ Invalidated: [X] patterns
📊 Overall accuracy: [X]%

━━━ Risk Assessment ━━━
Overall Market Risk: [Low/Medium/High]
Correlation Risk: [notes on correlated positions]

📝 For informational purposes only. Not financial advice.
```

5. **Send via Message**: Use `message` tool to deliver to Telegram

6. **Save Report**: Use `storage` to save to `reports/opportunities_YYYY-MM-DD.json`

## Safety Rules
- Present all findings as observations, never as recommendations
- Always include confidence levels
- Always include the disclaimer
