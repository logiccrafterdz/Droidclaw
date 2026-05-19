---
name: update_scorer_weights
description: Automatically adjust opportunity scoring weights and evolve Skill Genomes based on historical hit rates. Ensures the agent dynamically prioritizes indicators that are currently performing best in the market.
metadata: {"droidclaw":{"emoji":"⚖️","category":"economic","autonomous":true}}
---

# Adaptive Skill Evolution & Weights Update

Dynamically optimize the weight of each analytical component based on its real-world performance, and maintain the Skill Genome for self-improvement.

## When to Use
- Called automatically via cron or post-mortem triggers.
- Can be triggered manually via "/updateweights" command.

## Procedure

1. **Read Performance Data**: Use `storage` tool to read:
   - `postmortem/hit_rates.json` (running accuracy stats)
   - `scoring/weights.json` (current weights)

2. **Calculate Performance Metrics**:
   For each category (Trend, Volatility, Correlation, Macro, Carry):
   - Get the category-specific hit rate.
   - Requirement: Minimum 5 observations in the category to trigger adjustment.

3. **Apply Incremental Adjustments (Weights)**:
   - **Performance > 65%**: Increase weight by `+0.02`.
   - **Performance < 45%**: Decrease weight by `-0.02`.
   - **Performance 45%-65%**: No change (stability zone).

4. **Enforce Hard Constraints**:
   - **Weight Bounds**: Minimum weight = `0.1`, Maximum weight = `0.5`.
   - **Normalization**: After adjustments, ensure the sum of all weights remains `1.0`. Divide each weight by the new total sum.

5. **Update Weights File**: Use `storage` tool to write new values to `scoring/weights.json`.

6. **Skill Genome Evolution (Idea 2)**:
   For every skill that corresponds to an adjusted weight (e.g., `trend_regime_filter` for Trend, `volatility_noise_filter` for Volatility):
   - Read its `genome.json` file in the workspace's `skills/<skill_name>/` directory (create one if it doesn't exist).
   - Update the `"performance"` metrics inside `genome.json` with the latest hit rate.
   - Example `genome.json` structure:
     ```json
     {
       "name": "trend_regime_filter",
       "version": "1.0.0",
       "performance": {
         "hit_rate": 0.68,
         "total_observations": 12
       },
       "dependencies": ["market_data"]
     }
     ```
   - Use the `storage` or `write_file` tool to save the updated `genome.json`.

7. **Log the Change**: Use `storage` to append the change log to `scoring/weights_log.json`.

8. **Notify on Significant Shifts**: If any weight changes by more than 10% total, send a brief message via `message` tool.

## Safety Rules
- Adjustments MUST be small (max +/- 0.05 per session).
- Total sum of weights MUST always be 1.0.
- Never set a weight to 0.0 (all analytical components must contribute).
- Genomes are critical for the agent's long-term self-awareness; keep JSON structures strict.
