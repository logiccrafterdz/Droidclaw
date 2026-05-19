package skills

import (
	"strings"
)

type SkillRouter struct {
	keywordMap  map[string]string   // keyword -> skill name
	cronSkills  map[string][]string // cron phrase -> skills
	chainSkills map[string][]string // skill -> next skills
}

func NewSkillRouter() *SkillRouter {
	return &SkillRouter{
		keywordMap: map[string]string{
			"market":      "market_data",
			"price":       "market_data",
			"news":        "news_feed",
			"trend":       "trend_regime_filter",
			"regime":      "trend_regime_filter",
			"volatility":  "volatility_noise_filter",
			"noise":       "volatility_noise_filter",
			"correlation": "cross_asset_correlation",
			"macro":       "event_macro_trigger",
			"carry":       "trend_carry_regime",
			"opportunity": "detect_opportunity",
			"score":       "opportunity_scorer",
			"summarize":   "summarize_opportunities",
			"postmortem":  "post_mortem",
			"learn":       "learn_from_day",
			"alert":       "volatility_alert",
			"weights":     "update_scorer_weights",
		},
		cronSkills: map[string][]string{
			"detect_opportunity": {"trend_regime_filter", "volatility_noise_filter", "cross_asset_correlation", "event_macro_trigger", "trend_carry_regime", "detect_opportunity"},
			"score":              {"opportunity_scorer"},
			"summarize":          {"summarize_opportunities"},
			"learn":              {"learn_from_day"},
			"postmortem":         {"post_mortem"},
		},
		chainSkills: map[string][]string{
			"detect_opportunity": {"opportunity_scorer"},
			"opportunity_scorer": {"summarize_opportunities"},
			"post_mortem":        {"update_scorer_weights", "learn_from_day"},
		},
	}
}

func (sr *SkillRouter) ResolveSkills(message string, channel string, allSkills []SkillInfo) []SkillInfo {
	msgLower := strings.ToLower(message)
	selectedNames := make(map[string]bool)

	// 1. Cron triggers
	if channel == "cron" {
		for key, skills := range sr.cronSkills {
			if strings.Contains(msgLower, key) {
				for _, s := range skills {
					selectedNames[s] = true
				}
			}
		}
	}

	// 2. Keyword matching
	for kw, skill := range sr.keywordMap {
		if strings.Contains(msgLower, kw) {
			selectedNames[skill] = true
		}
	}

	// Also match direct skill names
	for _, s := range allSkills {
		nameLower := strings.ToLower(s.Name)
		nameSpaced := strings.ReplaceAll(nameLower, "_", " ")
		if strings.Contains(msgLower, nameLower) || strings.Contains(msgLower, nameSpaced) {
			selectedNames[s.Name] = true
		}
	}

	// 3. Chain skills
	for name := range selectedNames {
		if next, ok := sr.chainSkills[name]; ok {
			for _, n := range next {
				selectedNames[n] = true
			}
		}
	}

	// 4. Build final list
	var result []SkillInfo
	for _, s := range allSkills {
		if selectedNames[s.Name] {
			result = append(result, s)
		}
	}

	// Fallback: if nothing matched, return all
	if len(result) == 0 {
		return allSkills
	}

	return result
}
