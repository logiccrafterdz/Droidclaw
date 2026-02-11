package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MarketDataTool provides real-time and historical market data from free APIs.
// Supports crypto (Binance), forex/commodities (exchangerate), and general markets.
type MarketDataTool struct {
	client *http.Client
}

func NewMarketDataTool() *MarketDataTool {
	return &MarketDataTool{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *MarketDataTool) Name() string {
	return "market_data"
}

func (t *MarketDataTool) Description() string {
	return `Get real-time market data for crypto, forex, and commodities.
Actions:
- "ticker": Get current price, 24h change, volume for a symbol (e.g., BTCUSDT, ETHUSDT)
- "candles": Get OHLCV candlestick data (intervals: 1m,5m,15m,1h,4h,1d)
- "orderbook": Get order book snapshot (top bids/asks)
- "multi_ticker": Get tickers for multiple symbols at once
- "forex": Get forex exchange rates (base currency like USD, EUR)
For crypto symbols, use Binance format: BTCUSDT, ETHUSDT, etc.`
}

func (t *MarketDataTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"ticker", "candles", "orderbook", "multi_ticker", "forex"},
				"description": "Action to perform",
			},
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Trading symbol (e.g., BTCUSDT, ETHUSDT for crypto; USD, EUR for forex)",
			},
			"symbols": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Multiple symbols for multi_ticker action",
			},
			"interval": map[string]interface{}{
				"type":        "string",
				"description": "Candle interval: 1m, 5m, 15m, 1h, 4h, 1d (default: 1h)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results (default: 24 for candles, 10 for orderbook)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *MarketDataTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "ticker":
		return t.getTicker(ctx, args)
	case "candles":
		return t.getCandles(ctx, args)
	case "orderbook":
		return t.getOrderbook(ctx, args)
	case "multi_ticker":
		return t.getMultiTicker(ctx, args)
	case "forex":
		return t.getForex(ctx, args)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *MarketDataTool) getTicker(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "Error: symbol is required for ticker", nil
	}
	symbol = strings.ToUpper(symbol)

	// Special handling for Spot Gold (XAUUSD)
	if symbol == "XAUUSD" || symbol == "XAU" {
		return t.getForexRateAsTicker(ctx, "XAU")
	}

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", symbol)
	body, err := t.httpGet(ctx, url)
	if err != nil {
		if strings.Contains(err.Error(), "400") {
			return fmt.Sprintf("Error: Invalid symbol '%s'. Note: For Spot Gold use 'XAUUSD'.", symbol), nil
		}
		return fmt.Sprintf("Error fetching ticker: %v", err), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Sprintf("Error parsing response: %v", err), nil
	}

	result := map[string]interface{}{
		"symbol":           data["symbol"],
		"price":            data["lastPrice"],
		"price_change":     data["priceChange"],
		"price_change_pct": data["priceChangePercent"],
		"high_24h":         data["highPrice"],
		"low_24h":          data["lowPrice"],
		"volume_24h":       data["volume"],
		"quote_volume_24h": data["quoteVolume"],
		"open_price":       data["openPrice"],
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) getCandles(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "Error: symbol is required for candles", nil
	}
	symbol = strings.ToUpper(symbol)

	interval := "1h"
	if i, ok := args["interval"].(string); ok && i != "" {
		interval = i
	}

	limit := 24
	if l, ok := args["limit"].(float64); ok && int(l) > 0 {
		limit = int(l)
		if limit > 100 {
			limit = 100
		}
	}

	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	body, err := t.httpGet(ctx, url)
	if err != nil {
		return fmt.Sprintf("Error fetching candles: %v", err), nil
	}

	var rawCandles [][]interface{}
	if err := json.Unmarshal(body, &rawCandles); err != nil {
		return fmt.Sprintf("Error parsing candles: %v", err), nil
	}

	type Candle struct {
		Time   string `json:"time"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}

	candles := make([]Candle, 0, len(rawCandles))
	for _, c := range rawCandles {
		if len(c) < 6 {
			continue
		}
		ts := int64(c[0].(float64))
		candles = append(candles, Candle{
			Time:   time.UnixMilli(ts).UTC().Format("2006-01-02 15:04"),
			Open:   fmt.Sprintf("%v", c[1]),
			High:   fmt.Sprintf("%v", c[2]),
			Low:    fmt.Sprintf("%v", c[3]),
			Close:  fmt.Sprintf("%v", c[4]),
			Volume: fmt.Sprintf("%v", c[5]),
		})
	}

	result := map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"count":    len(candles),
		"candles":  candles,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) getOrderbook(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "Error: symbol is required for orderbook", nil
	}
	symbol = strings.ToUpper(symbol)

	limit := 10
	if l, ok := args["limit"].(float64); ok && int(l) > 0 {
		limit = int(l)
		if limit > 50 {
			limit = 50
		}
	}

	url := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=%d", symbol, limit)
	body, err := t.httpGet(ctx, url)
	if err != nil {
		return fmt.Sprintf("Error fetching orderbook: %v", err), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Sprintf("Error parsing orderbook: %v", err), nil
	}

	result := map[string]interface{}{
		"symbol":    symbol,
		"bids":      data["bids"],
		"asks":      data["asks"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) getMultiTicker(ctx context.Context, args map[string]interface{}) (string, error) {
	symbolsRaw, ok := args["symbols"]
	if !ok {
		// Default watchlist for economic monitoring
		symbolsRaw = []interface{}{"BTCUSDT", "ETHUSDT", "XAUUSDT", "EURUSDT", "SOLUSDT"}
	}

	var symbols []string
	switch v := symbolsRaw.(type) {
	case []interface{}:
		for _, s := range v {
			if str, ok := s.(string); ok {
				symbols = append(symbols, strings.ToUpper(str))
			}
		}
	case []string:
		for _, s := range v {
			symbols = append(symbols, strings.ToUpper(s))
		}
	}

	if len(symbols) == 0 {
		return "Error: symbols list is empty", nil
	}

	// Binance allows fetching multiple tickers with a single request
	symbolsJSON, _ := json.Marshal(symbols)
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbols=%s", string(symbolsJSON))

	body, err := t.httpGet(ctx, url)
	if err != nil {
		return fmt.Sprintf("Error fetching multi ticker: %v", err), nil
	}

	var tickers []map[string]interface{}
	if err := json.Unmarshal(body, &tickers); err != nil {
		return fmt.Sprintf("Error parsing response: %v", err), nil
	}

	type TickerSummary struct {
		Symbol    string `json:"symbol"`
		Price     string `json:"price"`
		Change    string `json:"change_pct"`
		High      string `json:"high_24h"`
		Low       string `json:"low_24h"`
		Volume    string `json:"volume"`
	}

	summaries := make([]TickerSummary, 0, len(tickers))
	for _, t := range tickers {
		summaries = append(summaries, TickerSummary{
			Symbol: fmt.Sprintf("%v", t["symbol"]),
			Price:  fmt.Sprintf("%v", t["lastPrice"]),
			Change: fmt.Sprintf("%v%%", t["priceChangePercent"]),
			High:   fmt.Sprintf("%v", t["highPrice"]),
			Low:    fmt.Sprintf("%v", t["lowPrice"]),
			Volume: fmt.Sprintf("%v", t["quoteVolume"]),
		})
	}

	result := map[string]interface{}{
		"count":     len(summaries),
		"tickers":   summaries,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) getForex(ctx context.Context, args map[string]interface{}) (string, error) {
	base := "USD"
	if b, ok := args["symbol"].(string); ok && b != "" {
		base = strings.ToUpper(b)
	}

	// Use exchangerate.host free API (no key required)
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", base)
	body, err := t.httpGet(ctx, url)
	if err != nil {
		return fmt.Sprintf("Error fetching forex rates: %v", err), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Sprintf("Error parsing forex data: %v", err), nil
	}

	// Extract key currencies
	rates, ok := data["rates"].(map[string]interface{})
	if !ok {
		return "Error: unexpected forex API response format", nil
	}

	keyCurrencies := []string{"EUR", "GBP", "JPY", "CHF", "AUD", "CAD", "CNY", "SAR", "AED", "TRY", "RUB", "INR", "BRL", "XAU"}
	filtered := make(map[string]interface{})
	for _, cur := range keyCurrencies {
		if rate, exists := rates[cur]; exists {
			filtered[cur] = rate
		}
	}

	result := map[string]interface{}{
		"base":         base,
		"rates":        filtered,
		"all_rates":    len(rates),
		"last_updated": data["time_last_update_utc"],
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) getForexRateAsTicker(ctx context.Context, symbol string) (string, error) {
	// Use EUR as base to get XAU price in EUR, then convert or just get USD base
	// Actually ER-API with USD base gives XAU rate (1/price)
	body, err := t.httpGet(ctx, "https://open.er-api.com/v6/latest/USD")
	if err != nil {
		return fmt.Sprintf("Error fetching gold price: %v", err), nil
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	rates := data["rates"].(map[string]interface{})

	rate, ok := rates[symbol].(float64)
	if !ok || rate == 0 {
		return fmt.Sprintf("Error: Price data for %s currently unavailable via Forex API", symbol), nil
	}

	price := 1.0 / rate // XAU rate is 1 oz per USD, so price is 1/rate

	result := map[string]interface{}{
		"symbol":           symbol + "USD",
		"price":            fmt.Sprintf("%.2f", price),
		"price_change":     "0.00",
		"price_change_pct": "0.00",
		"source":           "Forex API (Spot)",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *MarketDataTool) httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
