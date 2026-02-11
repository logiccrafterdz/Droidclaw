package tools

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NewsFeedTool provides economic news and calendar from free RSS/API sources.
type NewsFeedTool struct {
	client *http.Client
}

func NewNewsFeedTool() *NewsFeedTool {
	return &NewsFeedTool{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (t *NewsFeedTool) Name() string {
	return "news_feed"
}

func (t *NewsFeedTool) Description() string {
	return `Get economic news and market updates from RSS feeds and APIs.
Actions:
- "headlines": Get latest financial/economic headlines from multiple sources
- "crypto_news": Get crypto-specific news and updates
- "forex_news": Get forex/currency market news
- "search_news": Search news by keyword/topic
- "rss": Fetch any RSS feed URL and extract articles
Sources include: CoinDesk, CoinTelegraph, Investing.com, Reuters, Bloomberg RSS feeds.`
}

func (t *NewsFeedTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"headlines", "crypto_news", "forex_news", "search_news", "rss"},
				"description": "Action to perform",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query for search_news action",
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "RSS feed URL for rss action",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of articles to return (default: 10, max: 25)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *NewsFeedTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok && int(l) > 0 {
		limit = int(l)
		if limit > 25 {
			limit = 25
		}
	}

	switch action {
	case "headlines":
		return t.getHeadlines(ctx, limit)
	case "crypto_news":
		return t.getCryptoNews(ctx, limit)
	case "forex_news":
		return t.getForexNews(ctx, limit)
	case "search_news":
		query, _ := args["query"].(string)
		if query == "" {
			return "Error: query is required for search_news", nil
		}
		return t.searchNews(ctx, query, limit)
	case "rss":
		url, _ := args["url"].(string)
		if url == "" {
			return "Error: url is required for rss action", nil
		}
		return t.fetchRSS(ctx, url, limit)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// RSS XML structures
type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category"`
}

// Atom feed structures for some feeds
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   string    `xml:"title"`
	Link    atomLink  `xml:"link"`
	Summary string    `xml:"summary"`
	Updated string    `xml:"updated"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type newsArticle struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Summary     string `json:"summary"`
	Published   string `json:"published"`
	Source      string `json:"source"`
}

func (t *NewsFeedTool) getHeadlines(ctx context.Context, limit int) (string, error) {
	feeds := []struct {
		url    string
		source string
	}{
		{"https://feeds.finance.yahoo.com/rss/2.0/headline?s=^GSPC&region=US&lang=en-US", "Yahoo Finance"},
		{"https://www.investing.com/rss/news.rss", "Investing.com"},
		{"https://feeds.reuters.com/reuters/businessNews", "Reuters Business"},
	}

	return t.aggregateFeeds(ctx, feeds, limit)
}

func (t *NewsFeedTool) getCryptoNews(ctx context.Context, limit int) (string, error) {
	feeds := []struct {
		url    string
		source string
	}{
		{"https://www.coindesk.com/arc/outboundfeeds/rss/", "CoinDesk"},
		{"https://cointelegraph.com/rss", "CoinTelegraph"},
		{"https://cryptonews.com/news/feed/", "CryptoNews"},
	}

	return t.aggregateFeeds(ctx, feeds, limit)
}

func (t *NewsFeedTool) getForexNews(ctx context.Context, limit int) (string, error) {
	feeds := []struct {
		url    string
		source string
	}{
		{"https://www.dailyfx.com/feeds/all", "DailyFX"},
		{"https://www.forexlive.com/feed", "ForexLive"},
		{"https://feeds.finance.yahoo.com/rss/2.0/headline?s=EURUSD=X&region=US&lang=en-US", "Yahoo Forex"},
	}

	return t.aggregateFeeds(ctx, feeds, limit)
}

func (t *NewsFeedTool) searchNews(ctx context.Context, query string, limit int) (string, error) {
	// Use Google News RSS for search
	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s+economy+finance&hl=en&gl=US&ceid=US:en",
		strings.ReplaceAll(query, " ", "+"))

	articles, err := t.parseRSSFeed(ctx, url, "Google News")
	if err != nil {
		return fmt.Sprintf("Error searching news: %v", err), nil
	}

	if len(articles) > limit {
		articles = articles[:limit]
	}

	result := map[string]interface{}{
		"query":    query,
		"count":    len(articles),
		"articles": articles,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *NewsFeedTool) fetchRSS(ctx context.Context, url string, limit int) (string, error) {
	articles, err := t.parseRSSFeed(ctx, url, "Custom RSS")
	if err != nil {
		return fmt.Sprintf("Error fetching RSS: %v", err), nil
	}

	if len(articles) > limit {
		articles = articles[:limit]
	}

	result := map[string]interface{}{
		"url":       url,
		"count":     len(articles),
		"articles":  articles,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *NewsFeedTool) aggregateFeeds(ctx context.Context, feeds []struct {
	url    string
	source string
}, limit int) (string, error) {
	var allArticles []newsArticle

	for _, feed := range feeds {
		articles, err := t.parseRSSFeed(ctx, feed.url, feed.source)
		if err != nil {
			// Skip failed feeds, don't fail entirely
			continue
		}
		allArticles = append(allArticles, articles...)
	}

	if len(allArticles) == 0 {
		return "No articles found from any source.", nil
	}

	if len(allArticles) > limit {
		allArticles = allArticles[:limit]
	}

	result := map[string]interface{}{
		"count":     len(allArticles),
		"articles":  allArticles,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *NewsFeedTool) parseRSSFeed(ctx context.Context, feedURL, source string) ([]newsArticle, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, application/atom+xml")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try RSS format first
	var rss rssDocument
	if err := xml.Unmarshal(body, &rss); err == nil && len(rss.Channel.Items) > 0 {
		articles := make([]newsArticle, 0, len(rss.Channel.Items))
		for _, item := range rss.Channel.Items {
			summary := stripXMLTags(item.Description)
			if len(summary) > 300 {
				summary = summary[:300] + "..."
			}
			articles = append(articles, newsArticle{
				Title:     item.Title,
				Link:      item.Link,
				Summary:   summary,
				Published: item.PubDate,
				Source:    source,
			})
		}
		return articles, nil
	}

	// Try Atom format
	var atom atomFeed
	if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
		articles := make([]newsArticle, 0, len(atom.Entries))
		for _, entry := range atom.Entries {
			summary := stripXMLTags(entry.Summary)
			if len(summary) > 300 {
				summary = summary[:300] + "..."
			}
			articles = append(articles, newsArticle{
				Title:     entry.Title,
				Link:      entry.Link.Href,
				Summary:   summary,
				Published: entry.Updated,
				Source:    source,
			})
		}
		return articles, nil
	}

	return nil, fmt.Errorf("could not parse feed as RSS or Atom")
}

// stripXMLTags removes HTML/XML tags from a string.
func stripXMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, ch := range s {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}
	return strings.TrimSpace(result.String())
}
