---
name: event_macro_trigger
description: Analyze economic calendar events (CPI, NFP, FOMC, PMI, GDP, rate decisions) and major news as macro triggers. Classifies events by importance and type, correlates them with observed market reactions, and builds a pattern database of macro-to-market effects. Called automatically every 4 hours and after any high-impact event. Used by opportunity_scorer and daily_outlook.
metadata: {"droidclaw":{"emoji":"🏛️","category":"economic","autonomous":true}}
---

# Event & Macro Trigger Analyzer

Convert economic events and major news into structured "triggers" with tracked market impact.

## When to Use
- Called automatically every 4 hours by cron
- Called immediately when high-impact news is detected in scan_markets
- Called before daily_outlook to populate "Macro Triggers" section

## Procedure

1. **Gather Event Data**: Use `news_feed` tool:
   - `headlines` action, limit 10
   - `search_news` action with queries: "Federal Reserve", "CPI inflation", "NFP jobs", "rate decision", "GDP growth"
   - `forex_news` action, limit 5

2. **Classify Each Event**:
   For each news item, determine:
   
   **a) Importance Level**:
   - **High**: FOMC rate decisions, CPI/PPI releases, NFP, GDP, major central bank announcements
   - **Medium**: PMI data, consumer confidence, housing data, trade balance
   - **Low**: Regional economic data, secondary indicators, analyst opinions
   
   **b) Event Type**:
   - `inflation`: CPI, PPI, PCE, inflation expectations
   - `employment`: NFP, unemployment claims, ADP, wage growth
   - `growth`: GDP, PMI, industrial production, retail sales
   - `central_bank`: Rate decisions, FOMC minutes, speeches, QE/QT
   - `geopolitical`: Trade wars, sanctions, conflict, elections
   - `crypto_specific`: ETF approvals, regulation, exchange news, whale movements
   
   **c) Expected Impact Direction**:
   - For each event type, apply standard macro logic:
     - CPI higher than expected → USD strong → BTC pressure (short-term)
     - Rate cut → risk-on → BTC/crypto positive
     - Strong NFP → USD strong → mixed crypto
     - Dovish Fed → risk-on → crypto positive
     - ETF approval/inflow → crypto positive

3. **Correlate with Market Data**:
   - Read latest market data from `storage`: `scans/daily_log.json`
   - For each high/medium event:
     - Note the market reaction that followed (BTC moved X%, DXY moved Y%)
     - Compare actual reaction with expected direction
     - Store as confirmed/contradicted pattern

4. **Build Macro Trigger Record**:
   ```json
   {
     "timestamp": "2026-02-11T14:00:00Z",
     "active_triggers": [
       {
         "event": "FOMC Minutes Released",
         "importance": "high",
         "type": "central_bank",
         "expected_impact": "hawkish tone → USD positive, BTC pressure",
         "actual_reaction": {
           "BTC_change": "-1.2%",
           "DXY_proxy_change": "+0.3%",
           "reaction_aligned": true
         },
         "active": true,
         "decay_hours": 72
       }
     ],
     "upcoming_events": [
       {
         "event": "CPI Release",
         "expected_date": "2026-02-14",
         "importance": "high",
         "type": "inflation",
         "pre_positioning_note": "Markets may reduce risk ahead of CPI"
       }
     ],
     "macro_regime": "hawkish_hold",
     "macro_sentiment": "cautious"
   }
   ```

5. **Update Macro Patterns**: Use `storage` to update `patterns/macro_patterns.json`:
   - Record pattern: "Event X + Condition Y → Market moved Z"
   - Track hit rate per pattern type
   - Example entries:
     ```json
     {
       "patterns": [
         {
           "trigger": "CPI above expectations",
           "conditions": ["high DXY volume", "pre-event risk-off"],
           "typical_reaction": "BTC -2% to -5% within 1-3 days",
           "times_observed": 5,
           "times_confirmed": 4,
           "confidence": 8,
           "last_seen": "2026-02-11"
         }
       ]
     }
     ```

6. **Save Results**: Use `storage` tool:
   - Write to `macro/current_triggers.json`
   - Append to `macro/trigger_history.json`

## Trigger Decay Logic
- High-importance events: active for 72 hours, then decay
- Medium-importance: active for 48 hours
- Low-importance: active for 24 hours
- After decay, move to history (not actively used in scoring)

## How Other Skills Use This

- **opportunity_scorer**: +1 if opportunity direction aligns with active macro trigger, -1 if contradicts
- **detect_opportunity**: Include macro context in opportunity reasoning
- **daily_outlook**: "Macro Triggers" section showing active triggers and upcoming events
- **post_mortem**: Validate whether macro-driven predictions were accurate

## Safety Rules
- Never predict specific event outcomes (e.g., "CPI will be 3.2%")
- Present macro analysis as context, not prediction
- Always include "for informational purposes only" disclaimer
- Note that macro-to-crypto linkages are probabilistic, not deterministic
