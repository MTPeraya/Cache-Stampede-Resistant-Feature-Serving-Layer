package features

import "time"

type Feature struct {
	EntityID        int64     `json:"entity_id"`
	ClickRate       float64   `json:"click_rate"`
	PopularityScore float64   `json:"popularity_score"`
	BotScore        float64   `json:"bot_score"`
	UpdatedAt       time.Time `json:"updated_at"`
}
