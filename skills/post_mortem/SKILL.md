---
name: post_mortem
description: Review past opportunities after their expected timeframe has elapsed. Validate whether predicted directions materialized, compute hit rates by category, update pattern confidence, and record recommendations for threshold tuning. Does NOT auto-adjust rules - records findings for manual review. Called daily at 9 PM and weekly on Sundays for comprehensive review.
metadata: {"droidclaw":{"emoji":"📊","category":"economic","autonomous":true}}
---

# Post-Mortem & Pattern Update

Disciplined review loop that tracks prediction accuracy without auto-modifying rules.

## When to Use
- Called automatically at 9:00 PM daily by cron (daily review)
- Called automatically on Sundays at 8:00 PM (weekly comprehensive review)
- Can be triggered manually via "/postmortem" command

## Procedure

### Daily Review (9 PM)

1. **Read Scored Opportunities**: Use `storage` to read:
   - `opportunities/scored_*.json` from 3+ days ago (for short-term review)
   - `opportunities/active_scored.json` for currently active opportunities

2. **Read Current Market Data**: Use `market_data` tool:
   - `multi_ticker` for all assets that had scored opportunities
   - Compare current price to price at time of opportunity

3. **Evaluate Each Past Opportunity**:
   For each opportunity where enough time has passed (based on timeframe):
   - Short-term (hours): review after 24 hours
   - Medium-term (days): review after 3 days
   - Long-term (weeks): review after 14 days

   Determine:
   ```
   actual_move = (current_price - entry_price) / entry_price
   predicted_direction = opportunity.direction ("bullish" or "bearish")
   
   If predicted_direction = "bullish":
     hit = actual_move > 0
     magnitude = actual_move
   If predicted_direction = "bearish":
     hit = actual_move < 0
     magnitude = -actual_move
   ```

   Classification:
   - `confirmed_strong`: hit AND |magnitude| > 2% 
   - `confirmed_weak`: hit AND |magnitude| 0–2%
   - `invalidated`: NOT hit AND |magnitude| > 1%
   - `neutral`: |magnitude| < 1% (no meaningful move either way)

4. **Compute Hit Rates by Category**:
   Read all past reviews from `postmortem/review_history.json`:
   
   Calculate:
   - Overall hit rate = confirmed / (confirmed + invalidated)
   - Hit rate by opportunity type (breakout, divergence, news-driven, etc.)
   - Hit rate by confidence level (7+, 5-6, 4)
   - Hit rate by source skill (trend-driven, macro-driven, volatility-driven)
   - Hit rate by asset

5. **Update Pattern Confidence**:
   Use `storage` to update `patterns/known_patterns.json`:
   - For confirmed patterns: increment `times_correct`, maintain/increase confidence
   - For invalidated patterns: increment `times_observed` without `times_correct`, decrease confidence by 1
   - Remove patterns with confidence < 3 AND times_observed >= 5

6. **Generate Recommendations** (stored, not auto-applied):
   Based on hit rates, generate tuning suggestions:
   ```json
   {
     "recommendations": [
       {
         "type": "weight_adjustment",
         "detail": "Macro-driven opportunities have 65% hit rate vs 45% for pure volatility. Consider increasing macro_trigger weight.",
         "priority": "medium"
       },
       {
         "type": "threshold_adjustment",
         "detail": "Opportunities with confidence 4-5 have only 35% hit rate. Consider raising minimum threshold to 5.",
         "priority": "high"
       },
       {
         "type": "asset_note",
         "detail": "XRPUSDT opportunities consistently underperform (30% hit rate). Consider reducing weight or adding extra filter.",
         "priority": "low"
       }
     ]
   }
   ```

7. **Save Review**: Use `storage` tool:
   - Write to `postmortem/review_YYYY-MM-DD.json`
   - Append to `postmortem/review_history.json`
   - Update `postmortem/hit_rates.json` with running statistics
   - Write recommendations to `postmortem/recommendations.json`

8. **Send Summary** (daily): Use `message` tool:
   ```
   📊 Daily Post-Mortem
   📅 [Date]

   ━━━ Results ━━━
   Reviewed: X opportunities
   ✅ Confirmed: X (strong: X, weak: X)
   ❌ Invalidated: X
   ➖ Neutral: X

   ━━━ Hit Rates ━━━
   Overall: XX%
   High confidence (7+): XX%
   Trend-driven: XX%
   Macro-driven: XX%

   ━━━ Pattern Updates ━━━
   Updated: X patterns
   Removed: X low-confidence patterns

   ━━━ Recommendations ━━━
   [Top recommendation if any]

   📝 Full report: postmortem/review_YYYY-MM-DD.json
   ```

### Weekly Review (Sunday 8 PM)

All of the above, plus:

9. **Aggregate Weekly Stats**:
   - Total opportunities scored this week
   - Win rate by day of week
   - Best/worst performing category
   - Trend regime accuracy (how often regime classification was correct)
   - Correlation predictions accuracy

10. **Update Memory**: Use `write_file` to update `memory/MEMORY.md`:
    - Add weekly performance summary under "## Performance Tracking"
    - Update "## Reliable Patterns" with high-confidence patterns
    - Keep total file under 2000 words

11. **Send Weekly Report**: Extended version with weekly aggregates via `message` tool

## Key Principles
- **No auto-tuning**: This skill RECORDS findings and SUGGESTS changes. It does NOT modify thresholds or weights automatically.
- **Transparency**: All hit rates and recommendations are stored and accessible.
- **Conservatism**: Only remove patterns after sufficient observations (5+) with consistently low confidence.
- **Your role**: Review `postmortem/recommendations.json` periodically and manually adjust skill parameters if warranted.

## Safety Rules
- Past performance does not predict future results - always include this caveat
- All output is "for informational purposes only"
- Never suggest the system is "learning to trade" - it is tracking observation accuracy
