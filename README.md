<div align="center">
<img src="assets/logo.png" alt="DroidClaw" width="512">

<h1>DroidClaw: Economic Monitor on Android</h1>

<h3>Intelligent Economic Intelligence · Market Surveillance · Android (Termux) Native · 10MB RAM</h3>
<h3></h3>

<p>
<img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
<img src="https://img.shields.io/badge/Platform-Android%20(Termux)%20%7C%20Linux-green" alt="Platform">
<img src="https://img.shields.io/badge/Project-Economic%20Watcher-orange" alt="Type">
<img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
</p>

</div>

---

**DroidClaw** is a highly specialized, ultra-lightweight AI Assistant designed for 24/7 economic monitoring and financial analysis. Optimized to run within the **Android (Termux)** environment, it transforms a mobile device into a powerful, autonomous market watchtower.

Inspired by [PicoClaw](https://github.com/sipeed/picoclaw), DroidClaw has been refactored to focus exclusively on financial intelligence, multi-agent coordination, and automated reporting.

## � Key Features (Implemented)

### 🧠 Advanced Multi-Agent Intelligence
- **Named Agent Support**: Orchestrate multiple specialized agents (e.g., `econ_watcher`) with independent workspaces and configurations.
- **Deeper Analysis**: Configurable `max_tool_iterations` (up to 30) for complex financial reasoning.
- **Persistent Memory**: A long-term memory layer that tracks market patterns, user preferences, and historical trends.

### 📊 Economic Data Powerhouse
Custom Go-native tools designed for high-frequency monitoring with zero overhead:
- **`market_data`**: Real-time prices, OHLCV candles, and orderbook snapshots from Binance, Yahoo Finance, and CoinGecko.
- **`news_feed`**: Automated aggregation of economic calendars and news from RSS sources like Forex Factory, Investing.com, and Reuters.
- **`storage`**: Sandboxed JSON/CSV storage within the workspace for data archival and historical analysis.

### 🤖 Autonomous Workflows (Cron-Driven)
Fully automated monitoring pipeline via integrated cron service:
- **Market Scans**: Every 15 minutes to track volatility and trends.
- **Opportunity Detection**: Periodic analysis of scan data to identify high-confidence setups.
- **Daily Briefings**: Automated morning (7:00 AM) and evening (10:00 PM) reports delivered directly to Telegram.
- **Volatility Alerts**: Real-time notifications for price movements exceeding 3%.

### 📱 Optimized for Android
- **Termux Native**: Built to run efficiency on Android devices with minimal resources (<10MB RAM).
- **Proactive Reporting**: Delivering critical market alerts to your preferred messaging channel (Telegram, Discord, etc.).

---

## 🗺️ Development Roadmap

### 🛠️ Phase 3: Android System Bridge (Planned)
- **`android_ui` Tool**: Direct interaction with Android apps via Termux:API / ADB.
- **Notification Listener**: Monitoring mobile alerts for real-time news correlation.
- **Termux Bridge**: Enhanced HTTP bridge for Termux-specific system operations.

---

## 🛡️ Safety First
**DroidClaw is an observation and analysis tool only.**
- **No Trading Access**: The agent has no capability to execute orders or access financial accounts.
- **Informational Purpose**: All outputs are for data analysis and should not be considered financial advice.

---

## 🤝 Attribution
This project is a specialized fork of [PicoClaw](https://github.com/sipeed/picoclaw).
We maintain full compliance with the **MIT License**.

---

## 📦 Installation & Setup

1. **Environment**: Install **Go 1.24+** and **Termux** (if on Android).
2. **Initialization**:
   ```bash
   make deps
   make build
   ./picoclaw onboard econ
   ```
3. **Configuration**: Edit `~/.picoclaw/config.json` to add your LLM API keys and Telegram bot token.
4. **Run**:
   ```bash
   ./picoclaw gateway
   ```
