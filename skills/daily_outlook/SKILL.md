---
name: daily_outlook
description: Generate comprehensive daily economic outlook report for morning briefing via Telegram.
metadata: {"droidclaw":{"emoji":"📰","category":"economic","autonomous":true}}
---

# Daily Outlook

Generate a comprehensive morning briefing covering all major markets, upcoming events, and actionable insights.

## When to Use
- Called automatically at 7:00 AM daily by cron
- Should send the report to Telegram using the `message` tool

## Procedure

1. **Gather Market Data**:
   - Use `market_data` with `multi_ticker` for: BTCUSDT, ETHUSDT, SOLUSDT, XAUUSD, EURUSDT, GBPUSDT
   - Use `market_data` with `forex` for USD rates
   - Use `market_data` with `candles` for BTC (1d interval, 7 candles) to get weekly trend

2. **Gather News**:
   - Use `news_feed` with `headlines` (limit: 5)
   - Use `news_feed` with `crypto_news` (limit: 5)

3. **Review Yesterday's Data**:
   - Use `storage` to read `scans/daily_log.json` for yesterday's scans
   - Use `storage` to read `opportunities/active_scored.json` for scored active opportunities

4. **Read Analytical Skill Outputs**:
   - `regime/current_regime.json` (from trend_regime_filter) — regime per asset
   - `volatility/current_noise_profile.json` (from volatility_noise_filter) — vol status
   - `correlation/current_correlations.json` (from cross_asset_correlation) — cluster summary
   - `macro/current_triggers.json` (from event_macro_trigger) — active triggers & upcoming events
   - `carry/current_carry_regime.json` (from trend_carry_regime) — carry alignment
   - `postmortem/hit_rates.json` (from post_mortem) — recent accuracy

5. **Review Memory**: Read `memory/MEMORY.md` for context and patterns

5. **Compose the Report** in this format:

```
📰 Daily Economic Outlook
📅 [Date] | ⏰ [Time]

━━━ 🔵 CRYPTO ━━━
BTC: $XX,XXX (▲/▼ X.X%)
ETH: $X,XXX (▲/▼ X.X%)
SOL: $XXX (▲/▼ X.X%)

━━━ 💱 FOREX ━━━
EUR/USD: X.XXXX (▲/▼ X.X%)
GBP/USD: X.XXXX (▲/▼ X.X%)

━━━ 🥇 COMMODITIES ━━━
Gold: $X,XXX (▲/▼ X.X%)

━━━ 📊 MARKET STATUS ━━━
[Overall sentiment: Bullish/Bearish/Neutral]
[Key observation from overnight]

━━━ 📰 KEY NEWS ━━━
1. [Headline 1]
2. [Headline 2]
3. [Headline 3]

━━━ 🔍 WATCH TODAY ━━━
• [Thing to watch #1]
• [Thing to watch #2]

━━━ ⚠️ ACTIVE OPPORTUNITIES ━━━
[List scored opportunities from opportunity_scorer - max top 3]

━━━ 📈 REGIME STATUS ━━━
[Per asset: Trend-up/Trend-down/Range-bound + strength from trend_regime_filter]

━━━ 📉 VOLATILITY STATUS ━━━
[Vol regime: High/Normal/Low + any assets with elevated z-scores]

━━━ 🔗 CORRELATION STATUS ━━━
[Cluster summary + any correlation shifts from cross_asset_correlation]

━━━ 🏛️ MACRO TRIGGERS ━━━
[Active triggers + upcoming high-impact events from event_macro_trigger]

━━━ 💹 CARRY ALIGNMENT ━━━
[Carry-supported trends vs carry-opposed from trend_carry_regime]

━━━ 📊 ACCURACY TRACKER ━━━
[Recent hit rate from post_mortem if available]

📝 Note: For informational purposes only. Not financial advice.
```

6. **Send Report**: Use `message` tool to send to the configured Telegram channel

7. **Save Report**: Use `storage` to save to `reports/outlook_YYYY-MM-DD.json`

## Important
- Keep the report concise and scannable
- Use emojis for quick visual parsing
- Always include the disclaimer
- If markets are closed (weekend), note it and focus on crypto (24/7)
