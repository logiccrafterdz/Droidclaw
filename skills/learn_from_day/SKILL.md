---
name: learn_from_day
description: Review the day's market events and update long-term memory with learned patterns and observations.
metadata: {"droidclaw":{"emoji":"🧠","category":"economic","autonomous":true}}
---

# Learn From Day

End-of-day reflection that reviews market events, validates or invalidates earlier predictions, and updates long-term memory.

## When to Use
- Called automatically at 10:00 PM daily by cron
- Also called hourly (lighter version) to keep patterns fresh

## Procedure

### Full Daily Review (10 PM)

1. **Read Today's Scans**: Use `storage` to read `scans/daily_log.json`

2. **Read Today's Opportunities**: Use `storage` to read `opportunities/` directory for today's files, including `opportunities/active_scored.json` (scored by opportunity_scorer)

3. **Read Post-Mortem Data**: Use `storage` to read:
   - `postmortem/hit_rates.json` (running accuracy statistics from post_mortem)
   - `postmortem/recommendations.json` (tuning suggestions)
   - `postmortem/review_history.json` (past validations)

4. **Read Current Memory**: Use `read_file` to read `memory/MEMORY.md`

5. **Read Analytical Skill History**:
   - `regime/regime_changes.json` (regime transitions from trend_regime_filter)
   - `correlation/correlation_history.json` (correlation shifts)
   - `macro/trigger_history.json` (macro event outcomes)

4. **Analyze What Happened**:
   - Which market movements were predicted correctly?
   - Which predictions were wrong? Why?
   - What unexpected events occurred?
   - What patterns repeated from previous days?
   - What correlations held or broke?

5. **Extract Learnings**:
   - New patterns discovered
   - Patterns confirmed (increase confidence)
   - Patterns invalidated (decrease confidence or remove)
   - New correlations observed
   - News source reliability (which sources led to accurate predictions?)

6. **Update Memory**: Use `write_file` to update `memory/MEMORY.md` with:
   - Add new observations under "## Market Observations"
   - Update confidence levels for known patterns
   - Add any new correlations under "## Correlations"
   - Keep total memory file under 2000 words (summarize older entries)

7. **Update Patterns File**: Use `storage` to update `patterns/known_patterns.json`:
```json
{
  "patterns": [
    {
      "name": "BTC Weekend Rally",
      "description": "BTC tends to rally on Sunday evenings",
      "confidence": 6,
      "times_observed": 3,
      "times_correct": 2,
      "last_seen": "2025-01-15",
      "conditions": ["low weekend volume", "positive Friday close"]
    }
  ],
  "correlations": [
    {
      "pair": ["BTCUSDT", "ETHUSDT"],
      "type": "positive",
      "strength": 0.85,
      "note": "ETH follows BTC with 2-4h delay"
    }
  ],
  "last_updated": "2025-01-15T22:00:00Z"
}
```

8. **Generate Evening Summary**: Create a brief summary and send via `message` tool:
```
🧠 Daily Learning Summary
📅 [Date]

✅ Correct Predictions: X/Y
❌ Missed: [what was missed]
📝 New Pattern: [if any]
🔄 Updated: [what was updated in memory]

Key Takeaway: [one-sentence insight]
```

9. **Save Learning Log**: Use `storage` to append to `learning/log.json`

### Hourly Light Review

When called hourly (not at 10 PM):
1. Read latest scan from `scans/daily_log.json`
2. Check if any significant changes since last check
3. If a notable pattern is forming, append to `patterns/intraday_notes.json`
4. No Telegram message needed for hourly checks

## Memory Management
- Keep MEMORY.md focused and under 2000 words
- Archive old observations monthly to `learning/archive_YYYY-MM.json`
- Maintain top 20 most reliable patterns in known_patterns.json
- Remove patterns with confidence < 3 after 5+ observations
