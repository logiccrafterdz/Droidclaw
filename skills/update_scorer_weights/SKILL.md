---
name: update_scorer_weights
description: Automatically adjust opportunity scoring weights based on historical hit rates from post_mortem reports. Ensures the agent dynamically prioritizes indicators that are currently performing best in the market.
metadata: {"droidclaw":{"emoji":"⚖️","category":"economic","autonomous":true}}
---

# Update Scorer Weights

Dynamically optimize the weight of each analytical component based on its real-world performance.

## When to Use
- Called automatically once per week (Sunday evening)
- Can be triggered manually via "/updateweights" command

## Procedure

1. **Read Performance Data**: Use `storage` tool to read:
   - `postmortem/hit_rates.json` (running accuracy stats)
   - `scoring/weights.json` (current weights)

2. **Calculate Performance Metrics**:
   For each category (Trend, Volatility, Correlation, Macro, Carry):
   - Get the category-specific hit rate.
   - Requirement: Minimum 5 observations in the category to trigger adjustment.

3. **Apply Incremental Adjustments**:
   - **Performance > 65%**: Increase weight by `+0.02`.
   - **Performance < 45%**: Decrease weight by `-0.02`.
   - **Performance 45%-65%**: No change (stability zone).

4. **Enforce Hard Constraints**:
   - **Weight Bounds**: Minimum weight = `0.1`, Maximum weight = `0.5`.
   - **Normalization**: After adjustments, ensure the sum of all weights remains `1.0`. Divide each weight by the new total sum.

5. **Update Weights File**: Use `storage` tool to write new values to `scoring/weights.json`.

6. **Log the Change**: Use `storage` to append the change log to `scoring/weights_log.json`:
   ```json
   {
     "timestamp": "2026-02-12T13:00:00Z",
     "adjustments": {
       "macro_trigger": "+0.02",
       "volatility_signal": "-0.02"
     },
     "reasoning": "Macro hit rate at 72% over last 10 samples; Volatility hit rate dropped to 40%."
   }
   ```

7. **Notify on Significant Shifts**: If any weight changes by more than 10% total, send a brief message to Telegram.

## Safety Rules
- Adjustments MUST be small (max +/- 0.05 per week).
- Total sum of weights MUST always be 1.0.
- Never set a weight to 0.0 (all analytical components must contribute).
- All changes are "for informational purposes and internal optimization only".
