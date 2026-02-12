---
name: macro_explain_move
description: Perform deep, hypothesis-driven reasoning to explain significant price movements (e.g., >3% move in 24h). Instead of general news reading, it identifies potential macro drivers (Fed, USD, Geopolitics) and searches for evidence to support or refute them.
metadata: {"droidclaw":{"emoji":"⚖️","category":"economic","autonomous":false}}
---

# Macro Explain Move

Advanced reasoning to diagnose the fundamental "Why" behind a significant market move.

## When to Use
- Triggered automatically by `volatility_alert` when z_score > 3.0
- Triggered manually via `/explain [asset]`

## Procedure

1. **Identify the Move**:
   - Asset: (e.g., XAUUSD, BTCUSDT)
   - Magnitude: (e.g., +2.5% in 4h)
   - Current Price vs Previous baseline.

2. **Hypothesis Generation (Self-Questioning)**:
   Before searching, the agent must ask: "What are the most likely reasons for this move in [Asset]?"
   
   Potential drivers to consider:
   - **For Gold (XAUUSD)**: Real yields, USD Index (DXY), Geopolitical tension, Central Bank demand.
   - **For Crypto (BTC/ETH)**: ETF flows, Liquidity/Stablecoin changes, Macro risk sentiment, Regulatory news.
   - **For Forex (EUR/GBP)**: Central bank divergence (Fed vs ECB/BoE), Inflation data, Debt ceiling/Politics.

3. **Targeted Verification Search**:
   Use `web_search` or `news_feed` with specific hypothesis-driven queries:
   - Bad query: "gold news today"
   - Good query: "gold price move dollar index correlation today" OR "US 10-year yield impact on gold price [Date]"

4. **Structured Cross-Reference**:
   - Compare the timing of the price move with the timing of specific news releases (e.g., CPI release at 8:30 AM).
   - Check if related assets (DXY, SPX, yields) moved in coordination.

5. **Diagnostic Output**:
   Generate a structured report:
   
   ### 🦞 Macro Diagnosis: [Asset] [Move %]
   
   **Primary Driver**: (e.g., "Sharp decline in US Treasury yields following weak NFP data")
   
   **Supporting Evidence**:
   - "Yields fell 15bps immediately after the 8:30 AM release."
   - "DXY broke below 103.50 support."
   
   **Secondary Factors**: (e.g., "Increased safe-haven demand due to Middle East headlines")
   
   **Confidence Level**: (High/Medium/Low) - Based on how well news timing matches the price chart.
   
   **Outlook Impact**: Does this change the `regime`?

6. **Update Memory**: Save the insight to `macro/reasoning_log.json` and update `memory/MEMORY.md`.

## Safety Rules
- Be conservative: Use phrases like "Likely driven by" or "Correlates with".
- Avoid causality claims if timing doesn't match perfectly.
- Remind the user: "This is a diagnostic hypothesis, not financial advice."
